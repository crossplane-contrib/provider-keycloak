/*
Copyright 2021 Upbound Inc.
*/

package clients

import (
	"context"
	"encoding/json"
	"fmt"

	terraformSDK "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/v2/apis/common"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"

	"github.com/crossplane/upjet/v2/pkg/terraform"
	keycloakProvider "github.com/keycloak/terraform-provider-keycloak/provider"

	clusterv1beta1 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/v1beta1"
	namespacedv1beta1 "github.com/crossplane-contrib/provider-keycloak/apis/namespaced/v1beta1"
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
// https://registry.terraform.io/providers/keycloak/keycloak/latest/docs#argument-reference
// https://registry.terraform.io/providers/keycloak/keycloak/latest/docs#example-usage-client-credentials-grant
// https://registry.terraform.io/providers/keycloak/keycloak/latest/docs#example-usage-password-grant

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
// nolint: gocyclo
func TerraformSetupBuilder() terraform.SetupFn {
	return func(ctx context.Context, client client.Client, mg resource.Managed) (terraform.Setup, error) {
		ps := terraform.Setup{}

		pcSpec, err := resolveProviderConfig(ctx, client, mg)
		if err != nil {
			return terraform.Setup{}, err
		}

		creds, err := ExtractCredentials(ctx, pcSpec.Credentials.Source, client, pcSpec.Credentials.CommonCredentialSelectors)
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

		return ps, errors.Wrap(
			configureNoForkKeycloakClient(ctx, &ps),
			"failed to configure the no-fork client")
	}
}

// ExtractCredentials Function that extracts credentials from the secret provided to providerconfig
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

// Function to setup provider that uses terraform SDK
func configureNoForkKeycloakClient(ctx context.Context, ps *terraform.Setup) error {

	cb := keycloakProvider.KeycloakProvider(nil)

	diags := cb.Configure(ctx, terraformSDK.NewResourceConfigRaw(ps.Configuration))
	if diags.HasError() {
		return fmt.Errorf("failed to configure the Keycloak provider: %v", diags)
	}

	ps.Meta = cb.Meta()
	return nil
}

func legacyToModernProviderConfigSpec(pc *clusterv1beta1.ProviderConfig) (*namespacedv1beta1.ClusterProviderConfigSpec, error) {
	// TODO(erhan): this is hacky and potentially lossy, generate or manually implement
	if pc == nil {
		return nil, nil
	}
	data, err := json.Marshal(pc.Spec)
	if err != nil {
		return nil, err
	}

	var mSpec namespacedv1beta1.ClusterProviderConfigSpec
	err = json.Unmarshal(data, &mSpec)
	return &mSpec, err
}

func resolveProviderConfig(ctx context.Context, crClient client.Client, mg resource.Managed) (*namespacedv1beta1.ClusterProviderConfigSpec, error) {
	switch managed := mg.(type) {
	case resource.LegacyManaged:
		return resolveProviderConfigLegacy(ctx, crClient, managed)
	case resource.ModernManaged:
		return resolveProviderConfigModern(ctx, crClient, managed)
	default:
		return nil, errors.New("resource is not a managed")
	}
}

func resolveProviderConfigLegacy(ctx context.Context, client client.Client, mg resource.LegacyManaged) (*namespacedv1beta1.ClusterProviderConfigSpec, error) {
	configRef := mg.GetProviderConfigReference()
	if configRef == nil {
		return nil, errors.New(errNoProviderConfig)
	}
	pc := &clusterv1beta1.ProviderConfig{}
	if err := client.Get(ctx, types.NamespacedName{Name: configRef.Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetProviderConfig)
	}

	t := resource.NewLegacyProviderConfigUsageTracker(client, &clusterv1beta1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackUsage)
	}

	return legacyToModernProviderConfigSpec(pc)
}

func resolveProviderConfigModern(ctx context.Context, crClient client.Client, mg resource.ModernManaged) (*namespacedv1beta1.ClusterProviderConfigSpec, error) {
	configRef := mg.GetProviderConfigReference()
	if configRef == nil {
		return nil, errors.New(errNoProviderConfig)
	}

	pcRuntimeObj, err := crClient.Scheme().New(namespacedv1beta1.SchemeGroupVersion.WithKind(configRef.Kind))
	if err != nil {
		return nil, errors.Wrapf(err, "referenced provider config kind %q is invalid for %s/%s", configRef.Kind, mg.GetNamespace(), mg.GetName())
	}
	pcObj, ok := pcRuntimeObj.(resource.ProviderConfig)
	if !ok {
		return nil, errors.Errorf("referenced provider config kind %q is not a provider config type %s/%s", configRef.Kind, mg.GetNamespace(), mg.GetName())
	}

	// Namespace will be ignored if the PC is a cluster-scoped type
	if err := crClient.Get(ctx, types.NamespacedName{Name: configRef.Name, Namespace: mg.GetNamespace()}, pcObj); err != nil {
		return nil, errors.Wrap(err, errGetProviderConfig)
	}

	var pcSpec namespacedv1beta1.ClusterProviderConfigSpec
	switch pc := pcObj.(type) {
	case *namespacedv1beta1.ProviderConfig:
		pcSpec = namespacedv1beta1.ClusterProviderConfigSpec{
			Credentials: namespacedv1beta1.ClusterProviderCredentials{
				Source: "Secret",
				CommonCredentialSelectors: common.CommonCredentialSelectors{
					SecretRef: &common.SecretKeySelector{
						Key: pc.Spec.CredentialsSecretRef.Key,
						SecretReference: common.SecretReference{
							Name:      pc.Spec.CredentialsSecretRef.Name,
							Namespace: mg.GetNamespace(),
						},
					},
				},
			},
		}
	case *namespacedv1beta1.ClusterProviderConfig:
		pcSpec = pc.Spec
	default:
		return nil, errors.New("unknown provider config kind")
	}
	t := resource.NewProviderConfigUsageTracker(crClient, &namespacedv1beta1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackUsage)
	}
	return &pcSpec, nil
}
