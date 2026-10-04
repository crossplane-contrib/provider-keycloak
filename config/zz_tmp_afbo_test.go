package config

import (
	"encoding/json"
	"testing"

	"github.com/crossplane/upjet/v2/pkg/controller/conversion"
	"k8s.io/apimachinery/pkg/runtime"

	clientv1alpha1 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/openidclient/v1alpha1"
	clientv1alpha2 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/openidclient/v1alpha2"
)

func TestTmpAFBO(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientv1alpha1.SchemeBuilder.AddToScheme(scheme)
	_ = clientv1alpha2.SchemeBuilder.AddToScheme(scheme)
	pcC, _ := GetProvider(true)
	pcN, _ := GetProviderNamespaced(true)
	if err := conversion.RegisterConversions(pcC, pcN, scheme); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{
		`{"spec":{"forProvider":{"clientId":"c","authenticationFlowBindingOverrides":[{"browserIdRef":{"name":"my-flow"}}]}}}`,
		`{"spec":{"forProvider":{"clientId":"c","authenticationFlowBindingOverrides":[{"browserId":"154a"}]}}}`,
		`{"spec":{"forProvider":{"clientId":"c","authenticationFlowBindingOverrides":[{"browserId":"154a","browserIdRef":{"name":"my-flow"}}]}}}`,
	} {
		src := &clientv1alpha1.Client{}
		if err := json.Unmarshal([]byte(s), src); err != nil {
			t.Fatal(err)
		}
		dst := &clientv1alpha2.Client{}
		if err := src.ConvertTo(dst); err != nil {
			t.Fatal(err)
		}
		b, _ := json.Marshal(dst.Spec)
		t.Logf("up: %s", b)
		back := &clientv1alpha1.Client{}
		if err := back.ConvertFrom(dst); err != nil {
			t.Fatal(err)
		}
		b, _ = json.Marshal(back.Spec)
		t.Logf("down: %s", b)
	}
}
