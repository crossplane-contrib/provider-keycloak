/*
Copyright 2021 Upbound Inc.
*/

package clients

import (
	"context"
	"encoding/json"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/upbound/upjet/pkg/terraform"

	"github.com/corewire/apis/v1beta1"
)

const (
	// error messages
	errNoProviderConfig     = "no providerConfigRef provided"
	errGetProviderConfig    = "cannot get referenced ProviderConfig"
	errTrackUsage           = "cannot track ProviderConfig usage"
	errExtractCredentials   = "cannot extract credentials"
	errUnmarshalCredentials = "cannot unmarshal keycloak credentials as JSON"
)


// Password and client secret auth parameters  + general config parameters
// https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs#argument-reference
// https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs#example-usage-client-credentials-grant
// https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs#example-usage-password-grant
const (
	usernameKey = "username"
	passwordKey = "password"
	clientIDKey = "client_id"
	clientSecretKey = "client_secret"
	urlKey      = "url"
	realmKey      = "realm"
	basePathKey      = "base_path"
	additionalHeadersKey = "additional_headers"
)


// TerraformSetupBuilder builds Terraform a terraform.SetupFn function which
// returns Terraform provider setup configuration
func TerraformSetupBuilder(version, providerSource, providerVersion string) terraform.SetupFn {
	return func(ctx context.Context, client client.Client, mg resource.Managed) (terraform.Setup, error) {
		ps := terraform.Setup{
			Version: version,
			Requirement: terraform.ProviderRequirement{
				Source:  providerSource,
				Version: providerVersion,
			},
		}

		configRef := mg.GetProviderConfigReference()
		if configRef == nil {
			return ps, errors.New(errNoProviderConfig)
		}
		pc := &v1beta1.ProviderConfig{}
		if err := client.Get(ctx, types.NamespacedName{Name: configRef.Name}, pc); err != nil {
			return ps, errors.Wrap(err, errGetProviderConfig)
		}

		t := resource.NewProviderConfigUsageTracker(client, &v1beta1.ProviderConfigUsage{})
		if err := t.Track(ctx, mg); err != nil {
			return ps, errors.Wrap(err, errTrackUsage)
		}

		data, err := resource.CommonCredentialExtractor(ctx, pc.Spec.Credentials.Source, client, pc.Spec.Credentials.CommonCredentialSelectors)
		if err != nil {
			return ps, errors.Wrap(err, errExtractCredentials)
		}
		creds := map[string]string{}
		if err := json.Unmarshal(data, &creds); err != nil {
			return ps, errors.Wrap(err, errUnmarshalCredentials)
		}

        // set provider configuration
        ps.Configuration = map[string]any{}
		// username
		if v, ok := creds[usernameKey]; ok {
		  ps.Configuration[usernameKey] = v
		}
		// password
		if v, ok := creds[passwordKey]; ok {
		  ps.Configuration[passwordKey] = v
		}
		// client
        if v, ok := creds[clientIDKey]; ok {
          ps.Configuration[clientIDKey] = v
        }
		// secret
        if v, ok := creds[clientSecretKey]; ok {
          ps.Configuration[clientSecretKey] = v
        }
		// url 
        if v, ok := creds[urlKey]; ok {
          ps.Configuration[urlKey] = v
        }
		// realm
        if v, ok := creds[realmKey]; ok {
          ps.Configuration[realmKey] = v
        }
		// basepath
        if v, ok := creds[basePathKey]; ok {
          ps.Configuration[basePathKey] = v
        }
		// additional headers 
        if v, ok := creds[additionalHeadersKey]; ok {
          ps.Configuration[additionalHeadersKey] = v
        }
		return ps, nil
	}
}
