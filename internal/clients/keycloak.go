/*
Copyright 2021 Upbound Inc.
*/

package clients

import (
	"context"
	"encoding/json"
	"fmt"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/upjet/pkg/terraform"

	"github.com/crossplane-contrib/provider-keycloak/apis/v1beta1"
	terraformSDK "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	keycloakProvider "github.com/mrparkers/terraform-provider-keycloak/provider"
)

const (
	// error messages
	errNoProviderConfig     = "no providerConfigRef provided"
	errGetProviderConfig    = "cannot get referenced ProviderConfig"
	errTrackUsage           = "cannot track ProviderConfig usage"
	errExtractCredentials   = "cannot extract credentials"
	errUnmarshalCredentials = "cannot unmarshal keycloak credentials as JSON"
	errExtractSecretKey     = "cannot extract from secret key when none specified"
	errGetCredentialsSecret = "cannot get credentials secret"
)

// Password and client secret auth parameters  + general config parameters
// https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs#argument-reference
// https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs#example-usage-client-credentials-grant
// https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs#example-usage-password-grant

var requiredKeycloakConfigKeys = []string{
	"client_id",
	"url",
}

var optionalKeycloakConfigKeys = []string{
	"client_secret",
	"username",
	"password",
	"realm",
	"initial_login",
	"client_timeout",
	"tls_insecure_skip_verify",
	"root_ca_certificate",
	"base_path",
	"additional_headers",
	"red_hat_sso",
}

// TerraformSetupBuilder builds Terraform a terraform.SetupFn function which
// returns Terraform provider setup configuration
func TerraformSetupBuilder() terraform.SetupFn { // nolint: gocyclo
	return func(ctx context.Context, client client.Client, mg resource.Managed) (terraform.Setup, error) {
		ps := terraform.Setup{}

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

		creds, err := ExtractCredentials(ctx, pc.Spec.Credentials.Source, client, pc.Spec.Credentials.CommonCredentialSelectors)
		if err != nil {
			return ps, errors.Wrap(err, errExtractCredentials)
		}

		// set provider configuration
		ps.Configuration = map[string]any{}
		// Iterate over the requiredKeycloakConfigKeys, they must be set
		for _, key := range requiredKeycloakConfigKeys {
			if value, ok := creds[key]; ok {
				if !ok {
					// Return an error if a required key is missing
					return ps, errors.Errorf("required Keycloak configuration key '%s' is missing", key)
				}
				ps.Configuration[key] = value
			}
		}

		// Iterate over the optionalKeycloakConfigKeys, they can be set and do not have to be in the creds map
		for _, key := range optionalKeycloakConfigKeys {
			if value, ok := creds[key]; ok {
				ps.Configuration[key] = value
			}
		}

		return ps, errors.Wrap(configureNoForkKeycloakClient(ctx, &ps), "failed to configure the no-fork client")
	}
}

func configureNoForkKeycloakClient(ctx context.Context, ps *terraform.Setup) error {

	cb := keycloakProvider.KeycloakProvider(nil)

	diags := cb.Configure(ctx, terraformSDK.NewResourceConfigRaw(ps.Configuration))
	if diags.HasError() {
		return fmt.Errorf("failed to configure the Grafana provider: %v", diags)
	}

	ps.Meta = cb.Meta()
	return nil
}

func ExtractCredentials(ctx context.Context, source xpv1.CredentialsSource, client client.Client, selector xpv1.CommonCredentialSelectors) (map[string]string, error) {
	creds := make(map[string]string)

	// first try to see if the secret contains a proper key-value map
	if selector.SecretRef == nil {
		return nil, errors.New(errExtractSecretKey)
	}
	secret := &corev1.Secret{}
	if err := client.Get(ctx, types.NamespacedName{Namespace: selector.SecretRef.Namespace, Name: selector.SecretRef.Name}, secret); err != nil {
		return nil, errors.Wrap(err, errGetCredentialsSecret)
	}
	if _, ok := secret.Data[selector.SecretRef.Key]; !ok {
		for k, v := range secret.Data {
			creds[k] = string(v)
		}
		return creds, nil
	}

	// if that fails, use Crossplane's way of extracting a JSON document
	rawData, err := resource.CommonCredentialExtractor(ctx, source, client, selector)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(rawData, &creds); err != nil {
		return nil, errors.Wrap(err, errUnmarshalCredentials)
	}

	return creds, nil
}
