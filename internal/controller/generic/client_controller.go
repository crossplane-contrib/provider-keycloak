/*
Copyright 2022 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package generic

import (
	"context"
	"encoding/json"
	"fmt"

	gocloak "github.com/Nerzal/gocloak/v13"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-keycloak/apis/generic/v1alpha1"
	v1alpha1Generic "github.com/crossplane-contrib/provider-keycloak/apis/generic/v1alpha1"
	apisv1alpha1 "github.com/crossplane-contrib/provider-keycloak/apis/v1alpha1"
	apisv1beta1 "github.com/crossplane-contrib/provider-keycloak/apis/v1beta1"
	"github.com/crossplane-contrib/provider-keycloak/internal/features"
)

const (
	errNotClient    = "managed resource is not a Client custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"

	errNewClient = "cannot create new Service"
)

// A NoOpService does nothing.
type NoOpService struct{}

var (
	newNoOpService = func(_ []byte) (interface{}, error) { return &NoOpService{}, nil }
)

// Setup adds a controller that reconciles Client managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1Generic.ClientGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1Generic.ClientGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1beta1.ProviderConfigUsage{}),
			newServiceFn: newNoOpService}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1Generic.Client{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	usage        resource.Tracker
	newServiceFn func(creds []byte) (interface{}, error)
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1Generic.Client)
	if !ok {
		return nil, errors.New(errNotClient)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &apisv1beta1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	cd := pc.Spec.Credentials
	data, err := resource.CommonCredentialExtractor(ctx, cd.Source, c.kube, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	// Unmarshal the decoded data into a map
	var jsonData map[string]string
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, errors.Wrap(err, "error unmarshalling JSON data")
	}

	// Get credentials from JSON data
	url := jsonData["url"]
	basePath := jsonData["base_path"]
	user := jsonData["username"]
	password := jsonData["password"]
	realm := jsonData["realm"]
	clientID := jsonData["client_id"]
	clientSecret := jsonData["client_secret"]

	serverUrl := url + basePath

	// Create a GoCloak client
	kc_client := gocloak.NewClient(serverUrl)

	var token string
	// Authenticate based on available credentials
	if user != "" && password != "" {
		jwt, err := kc_client.LoginAdmin(ctx, user, password, realm)
		if err != nil {
			return nil, errors.Wrap(err, "failed to authenticate")
		}
		token = jwt.AccessToken
	} else if clientID != "" && clientSecret != "" {
		jwt, err := kc_client.LoginClient(ctx, clientID, clientSecret, realm)
		if err != nil {
			return nil, errors.Wrap(err, "failed to authenticate")
		}
		token = jwt.AccessToken
	} else {
		return nil, errors.New("no credentials provided")
	}

	// Return the external client and token
	return &external{client: kc_client, token: token}, nil

}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	// A 'client' used to connect to the external resource API. In practice this
	// would be something like an AWS SDK client.
	client *gocloak.GoCloak
	token  string
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Client)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotClient)
	}

	// These fmt statements should be removed in the real implementation.
	fmt.Println("Observing: %+v", cr)
	annotations := cr.GetAnnotations()
	fmt.Println("Annotations: %+v", annotations)
	externalName := annotations["crossplane.io/external-name"]
	fmt.Println("ExternalName: %+v", externalName)

	existing_client, err := c.client.GetClient(ctx, c.token, *cr.Spec.ForProvider.RealmID, externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to get client")
	}

	fmt.Println("ExistingClient: %+v", existing_client)
	/*
			ExistingClient: %+v {
		        "access": {
		                "configure": true,
		                "manage": true,
		                "view": true
		        },
		        "adminUrl": "",
		        "attributes": {
		                "backchannel.logout.revoke.offline.tokens": "false",
		                "backchannel.logout.session.required": "true",
		                "client.secret.creation.time": "1720725218",
		                "oauth2.device.authorization.grant.enabled": "false",
		                "oidc.ciba.grant.enabled": "false"
		        },
		        "authenticationFlowBindingOverrides": {},
		        "baseUrl": "",
		        "bearerOnly": false,
		        "clientAuthenticatorType": "client-secret",
		        "clientId": "test",
		        "consentRequired": false,
		        "defaultClientScopes": [
		                "web-origins",
		                "acr",
		                "roles",
		                "profile",
		                "email"
		        ],
		        "description": "",
		        "directAccessGrantsEnabled": true,
		        "enabled": true,
		        "frontchannelLogout": true,
		        "fullScopeAllowed": true,
		        "id": "90878f8a-b8a4-4514-9189-5d2fb95dd9e3",
		        "implicitFlowEnabled": false,
		        "name": "test",
		        "nodeReRegistrationTimeout": -1,
		        "notBefore": 0,
		        "optionalClientScopes": [
		                "address",
		                "phone",
		                "offline_access",
		                "microprofile-jwt"
		        ],
		        "protocol": "openid-connect",
		        "publicClient": false,
		        "redirectUris": [
		                "/*"
		        ],
		        "rootUrl": "",
		        "secret": "CXwTPItWvexFa6H0UbH0QqYXJv3anWAL",
		        "serviceAccountsEnabled": false,
		        "standardFlowEnabled": true,
		        "surrogateAuthRequired": false,
		        "webOrigins": [
		                "/*"
		        ]
		}
	*/

	cr.Status.AtProvider.ClientID = *existing_client.ClientID
	cr.Status.AtProvider.RealmID = cr.Spec.ForProvider.RealmID
	cr.Status.AtProvider.Name = existing_client.Name
	cr.Status.AtProvider.Description = existing_client.Description
	cr.Status.AtProvider.Enabled = existing_client.Enabled
	

	authflow := *existing_client.AuthenticationFlowBindingOverrides
	if authflow != nil {
		// browser_id
		browser_id := authflow["browser_id"]

		// direct_grant_id
		direct_grant_id := authflow["direct_grant_id"]
		cr.Status.AtProvider.AuthenticationFlowBindingOverrides = []v1alpha1Generic.AuthenticationFlowBindingOverridesParameters{
			{
				BrowserID:     &browser_id,
				DirectGrantID: &direct_grant_id,
			},
		}

	}

	cr.SetConditions(v1.Available())

	return managed.ExternalObservation{
		// Return false when the external resource does not exist. This lets
		// the managed resource reconciler know that it needs to call Create to
		// (re)create the resource, or that it has successfully been deleted.
		ResourceExists: true,

		// Return false when the external resource exists, but it not up to date
		// with the desired managed resource state. This lets the managed
		// resource reconciler know that it needs to call Update.
		ResourceUpToDate: true,

		// Return any details that may be required to connect to the external
		// resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Client)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotClient)
	}

	fmt.Printf("Creating: %+v", cr)

	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Client)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotClient)
	}

	fmt.Printf("Updating: %+v", cr)

	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Client)
	if !ok {
		return errors.New(errNotClient)
	}

	fmt.Printf("Deleting: %+v", cr)

	return nil
}
