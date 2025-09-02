package clients

import (
	"context"
	"reflect"
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
		want    map[string]string
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
  "client_id": "test-client",
  "username":  "tester",
  "password":  "53cr37",
  "url":       "my-keycloak.nmspc.svc.cluster.local"
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
			want: map[string]string{
				"client_id": "test-client",
				"username":  "tester",
				"password":  "53cr37",
				"url":       "my-keycloak.nmspc.svc.cluster.local",
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
							"client_id": []byte("test-client"),
							"username":  []byte("tester"),
							"password":  []byte("53cr37"),
							"url":       []byte("my-keycloak.nmspc.svc.cluster.local"),
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
			want: map[string]string{
				"client_id": "test-client",
				"username":  "tester",
				"password":  "53cr37",
				"url":       "my-keycloak.nmspc.svc.cluster.local",
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
