/*
Copyright 2022 Upbound Inc.
*/

package connectionsecrettransform

import (
	"context"
	"net/url"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	clusterv1beta1 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/v1beta1"
	"github.com/crossplane-contrib/provider-keycloak/internal/clients"
)

// Recognized "keycloak:" field names, i.e. the values that are computed from
// the ProviderConfig's Keycloak URL and the owning managed resource's realm
// rather than read from an object field.
const (
	keycloakFieldURL           = "url"
	keycloakFieldRealm         = "realm"
	keycloakFieldIssuerURL     = "issuerUrl"
	keycloakFieldWellKnownURL  = "wellKnownUrl"
	keycloakFieldAuthzURL      = "authorizationUrl"
	keycloakFieldTokenURL      = "tokenUrl"
	keycloakFieldUserInfoURL   = "userinfoUrl"
	keycloakFieldJWKSURL       = "jwksUrl"
	keycloakFieldEndSessionURL = "endSessionUrl"
)

// credentialsKeyURL and credentialsKeyBasePath are the only two entries of
// the ProviderConfig's credentials this controller ever reads. Both are
// addressing information rather than secret material (they are what a
// browser is redirected to), so publishing values derived from them into a
// connection secret cannot leak a credential. Every other entry - most
// notably client_secret, password and the JWT signing key - is ignored.
const (
	credentialsKeyURL      = "url"
	credentialsKeyBasePath = "base_path"
)

// realmPaths are the fields a managed resource's realm is looked up in, in
// order. Terraform's realm_id is the realm's name (that is the Keycloak
// realm identifier), so a resource that lives in a realm carries it in
// status.atProvider.realmId once observed and in spec.forProvider.realmId
// beforehand; a Realm itself carries its name in the "realm" field instead.
var realmPaths = [][]string{
	{statusField, atProviderField, realmIDField},
	{specField, forProviderField, realmIDField},
	{statusField, atProviderField, realmField},
	{specField, forProviderField, realmField},
}

// The managed resource fields realmPaths is built from: the desired
// (spec.forProvider) and observed (status.atProvider) state of a Crossplane
// managed resource, and the two names a realm is spelled with - "realmId"
// for a resource that lives in a realm, "realm" for a Realm itself.
const (
	specField        = "spec"
	statusField      = "status"
	atProviderField  = "atProvider"
	forProviderField = "forProvider"
	realmIDField     = "realmId"
	realmField       = "realm"
)

// keycloakSource resolves "keycloak:<field>" add-field sources, i.e. values
// derived from the Keycloak deployment a resource was created in rather than
// read from a field of an existing object: most importantly the realm's OIDC
// issuer URL, which no single object carries because it combines the
// ProviderConfig's Keycloak URL with the resource's realm.
//
// It is resolved lazily and at most once per reconcile: the base URL comes
// from the ProviderConfig's credentials Secret, which is not worth reading
// when no source needs it.
type keycloakSource struct {
	resolve func() (string, string, error)

	done  bool
	base  string
	realm string
	err   error
}

// newKeycloakSource returns a keycloakSource for the given ProviderConfig and
// managed resource. pc may be nil (a namespaced resource, or one without a
// resolvable ProviderConfig), in which case every lookup fails with an
// explanatory error rather than panicking.
func newKeycloakSource(ctx context.Context, r *Reconciler, pc *clusterv1beta1.ProviderConfig, mrObj map[string]interface{}) *keycloakSource {
	return &keycloakSource{resolve: func() (string, string, error) {
		base, err := r.keycloakBaseURL(ctx, pc)
		if err != nil {
			return "", "", err
		}
		return base, realmOf(mrObj), nil
	}}
}

// value returns the value of a single "keycloak:" field.
func (k *keycloakSource) value(field string) (string, error) {
	if !k.done {
		k.base, k.realm, k.err = k.resolve()
		k.done = true
	}
	if k.err != nil {
		return "", k.err
	}

	if field == keycloakFieldURL {
		return k.base, nil
	}
	if k.realm == "" {
		return "", errors.Errorf("cannot determine the realm of the owning managed resource, which %q requires", addSourceKeycloak+field)
	}

	issuer := k.base + "/realms/" + url.PathEscape(k.realm)
	oidc := issuer + "/protocol/openid-connect"
	switch field {
	case keycloakFieldRealm:
		return k.realm, nil
	case keycloakFieldIssuerURL:
		return issuer, nil
	case keycloakFieldWellKnownURL:
		return issuer + "/.well-known/openid-configuration", nil
	case keycloakFieldAuthzURL:
		return oidc + "/auth", nil
	case keycloakFieldTokenURL:
		return oidc + "/token", nil
	case keycloakFieldUserInfoURL:
		return oidc + "/userinfo", nil
	case keycloakFieldJWKSURL:
		return oidc + "/certs", nil
	case keycloakFieldEndSessionURL:
		return oidc + "/logout", nil
	default:
		return "", errors.Errorf("unknown field %q (supported: %s)", field, strings.Join(keycloakFields(), ", "))
	}
}

// keycloakFields lists the recognized "keycloak:" field names, for error
// messages.
func keycloakFields() []string {
	f := []string{
		keycloakFieldURL,
		keycloakFieldRealm,
		keycloakFieldIssuerURL,
		keycloakFieldWellKnownURL,
		keycloakFieldAuthzURL,
		keycloakFieldTokenURL,
		keycloakFieldUserInfoURL,
		keycloakFieldJWKSURL,
		keycloakFieldEndSessionURL,
	}
	sort.Strings(f)
	return f
}

// realmOf returns the Keycloak realm the managed resource lives in, or "" if
// it cannot be determined (e.g. the resource is not realm-scoped, or has not
// been observed yet and does not spell its realm out in its spec because it
// uses a reference that has not been resolved yet).
func realmOf(mrObj map[string]interface{}) string {
	for _, p := range realmPaths {
		if v, found, err := unstructured.NestedString(mrObj, p...); err == nil && found && v != "" {
			return v
		}
	}
	return ""
}

// keycloakBaseURL returns the Keycloak base URL (including a configured base
// path, without a trailing slash) of the ProviderConfig's credentials, i.e.
// the URL every realm endpoint is built on.
func (r *Reconciler) keycloakBaseURL(ctx context.Context, pc *clusterv1beta1.ProviderConfig) (string, error) {
	if pc == nil {
		return "", errors.New("no ProviderConfig available for this resource")
	}

	creds, err := clients.ExtractCredentials(ctx, pc.Spec.Credentials.Source, r.client, pc.Spec.Credentials.CommonCredentialSelectors)
	if err != nil {
		return "", errors.Wrap(err, "cannot read the ProviderConfig credentials")
	}

	raw, _ := creds[credentialsKeyURL].(string)
	basePath, _ := creds[credentialsKeyBasePath].(string)
	return keycloakBaseURL(raw, basePath)
}

// keycloakBaseURL joins the credentials' "url" and "base_path" into the base
// URL of the Keycloak deployment, normalized the same way
// clients.TerraformSetupBuilder normalizes them for the Terraform provider:
// no trailing slash, base path prefixed with a slash.
func keycloakBaseURL(rawURL, basePath string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", errors.Errorf("the ProviderConfig credentials have no %q", credentialsKeyURL)
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", errors.Wrapf(err, "the ProviderConfig credentials have an invalid %q", credentialsKeyURL)
	}
	if u.Scheme == "" || u.Host == "" {
		return "", errors.Errorf("the ProviderConfig credentials have an invalid %q", credentialsKeyURL)
	}
	u.Path = strings.TrimRight(u.Path, "/")
	u.RawQuery = ""
	u.Fragment = ""

	basePath = strings.TrimRight(strings.TrimSpace(basePath), "/")
	if basePath != "" && !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	return u.String() + basePath, nil
}
