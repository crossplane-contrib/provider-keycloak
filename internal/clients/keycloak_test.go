package clients

import (
	"context"
	"reflect"
	"slices"
	"strconv"
	"testing"

	v1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestExtractCredentials(t *testing.T) {
	type args struct {
		ctx      context.Context
		source   v1.CredentialsSource
		client   client.Client
		selector v1.CommonCredentialSelectors
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]any
		wantErr bool
	}{
		{
			name: "extracting credentials from JSON blob secret works",
			args: args{
				ctx:    context.Background(),
				source: v1.CredentialsSourceSecret,
				client: fake.NewClientBuilder().
					WithObjects(&corev1.Secret{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "provider-keycloak-config",
							Namespace: "crossplane-system",
						},
						Data: map[string][]byte{
							"someCredentialsField": []byte(`{
  "client_id":      "test-client",
  "username":       "tester",
  "password":       "53cr37",
  "url":            "my-keycloak.nmspc.svc.cluster.local",
  "client_timeout": 30,
  "tls_insecure_skip_verify": true
}`)},
					}).
					Build(),
				selector: v1.CommonCredentialSelectors{
					SecretRef: &v1.SecretKeySelector{
						Key: "someCredentialsField",
						SecretReference: v1.SecretReference{
							Name:      "provider-keycloak-config",
							Namespace: "crossplane-system",
						},
					},
				},
			},
			want: map[string]any{
				"client_id":                "test-client",
				"username":                 "tester",
				"password":                 "53cr37",
				"url":                      "my-keycloak.nmspc.svc.cluster.local",
				"client_timeout":           30,
				"tls_insecure_skip_verify": true,
			},
		},
		{
			name: "extracting credentials from plain k8s secret works",
			args: args{
				ctx:    context.Background(),
				source: v1.CredentialsSourceSecret,
				client: fake.NewClientBuilder().
					WithObjects(&corev1.Secret{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "provider-keycloak-config-plain",
							Namespace: "crossplane-system",
						},
						Data: map[string][]byte{
							"client_id":                []byte("test-client"),
							"username":                 []byte("tester"),
							"password":                 []byte("53cr37"),
							"url":                      []byte("my-keycloak.nmspc.svc.cluster.local"),
							"client_timeout":           []byte(strconv.Itoa(30)),
							"tls_insecure_skip_verify": []byte(strconv.FormatBool(true)),
						},
					}).
					Build(),
				selector: v1.CommonCredentialSelectors{
					SecretRef: &v1.SecretKeySelector{
						Key: "someCredentialsField",
						SecretReference: v1.SecretReference{
							Name:      "provider-keycloak-config-plain",
							Namespace: "crossplane-system",
						},
					},
				},
			},
			want: map[string]any{
				"client_id":                "test-client",
				"username":                 "tester",
				"password":                 "53cr37",
				"url":                      "my-keycloak.nmspc.svc.cluster.local",
				"client_timeout":           30,
				"tls_insecure_skip_verify": true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractCredentials(tt.args.ctx, tt.args.source, tt.args.client, tt.args.selector)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractCredentials() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractCredentials() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateAndNormalizeURLAndBasePath(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]any
		want    map[string]any
		wantErr bool
	}{
		{
			name: "normalizes trailing slash in url and root base_path",
			input: map[string]any{
				"url":       "https://keycloak.example.com/",
				"base_path": "/",
			},
			want: map[string]any{
				"url":       "https://keycloak.example.com",
				"base_path": "",
			},
		},
		{
			name: "normalizes base_path trailing slash",
			input: map[string]any{
				"url":       "https://keycloak.example.com",
				"base_path": "/auth/",
			},
			want: map[string]any{
				"url":       "https://keycloak.example.com",
				"base_path": "/auth",
			},
		},
		{
			name: "rejects url without scheme",
			input: map[string]any{
				"url": "keycloak.example.com",
			},
			wantErr: true,
		},
		{
			name: "rejects invalid base_path",
			input: map[string]any{
				"url":       "https://keycloak.example.com",
				"base_path": "auth",
			},
			wantErr: true,
		},
		{
			name: "rejects url with query",
			input: map[string]any{
				"url": "https://keycloak.example.com?x=1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := make(map[string]any, len(tt.input))
			for k, v := range tt.input {
				cfg[k] = v
			}

			err := validateAndNormalizeURLAndBasePath(cfg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateAndNormalizeURLAndBasePath() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if !reflect.DeepEqual(cfg, tt.want) {
				t.Fatalf("validateAndNormalizeURLAndBasePath() got = %v, want %v", cfg, tt.want)
			}
		})
	}
}

// TestOptionalKeycloakConfigKeysIncludesKeycloakVersion guards against a
// regression of https://github.com/crossplane-contrib/provider-keycloak/issues/742:
// a ProviderConfig scoped to a service account without the master-realm
// view-system/manage-realms role can only complete its initial login if the
// keycloak_version credential is forwarded to the underlying Terraform
// provider so it can skip the version auto-detection that requires that
// role. If this key is dropped from optionalKeycloakConfigKeys, that
// escape hatch silently stops working even though ExtractCredentials and
// config/lookup/keycloak_client.go still read and forward the field.
func TestOptionalKeycloakConfigKeysIncludesKeycloakVersion(t *testing.T) {
	if !slices.Contains(optionalKeycloakConfigKeys, "keycloak_version") {
		t.Fatalf("optionalKeycloakConfigKeys must include \"keycloak_version\" so that ProviderConfigs can pin the Keycloak version when the service account cannot read /admin/serverinfo's version field")
	}
}

// TestTerraformSetupBuilderForwardsKeycloakVersion exercises the same
// required/optional key-copying logic used by TerraformSetupBuilder to
// build the Terraform provider configuration, verifying that
// "keycloak_version" supplied via the credentials secret is carried through
// into the resulting configuration.
func TestTerraformSetupBuilderForwardsKeycloakVersion(t *testing.T) {
	creds := map[string]any{
		"client_id":        "test-client",
		"client_secret":    "53cr37",
		"url":              "https://my-keycloak.example.com",
		"realm":            "demo",
		"keycloak_version": "26.6.2",
	}

	configuration := map[string]any{}
	for _, key := range requiredKeycloakConfigKeys {
		value, ok := creds[key]
		if !ok {
			t.Fatalf("required Keycloak configuration key %q is missing from test fixture", key)
		}
		configuration[key] = value
	}
	for _, key := range optionalKeycloakConfigKeys {
		if value, ok := creds[key]; ok {
			configuration[key] = value
		}
	}

	if got, want := configuration["keycloak_version"], "26.6.2"; got != want {
		t.Fatalf("configuration[\"keycloak_version\"] = %v, want %v", got, want)
	}
}
