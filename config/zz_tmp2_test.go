package config

import (
	"encoding/json"
	"testing"

	jsonpatch "github.com/evanphx/json-patch"

	clientv1alpha2 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/openidclient/v1alpha2"
)

func TestTmpPatch(t *testing.T) {
	for _, s := range []string{
		`{"spec":{"initProvider":{"authenticationFlowBindingOverrides":[{"browserIdRef":{"name":"my-flow"}}]},"forProvider":{"clientId":"c"}}}`,
		`{"spec":{"forProvider":{"clientId":"c","authenticationFlowBindingOverrides":[{"browserIdRef":{"name":"my-flow"}}]}}}`,
	} {
		ex := &clientv1alpha2.Client{}
		_ = json.Unmarshal([]byte(s), ex)
		res := ex.DeepCopy()
		u := "uuid"
		if len(res.Spec.InitProvider.AuthenticationFlowBindingOverrides) > 0 {
			res.Spec.InitProvider.AuthenticationFlowBindingOverrides[0].BrowserID = &u
		} else {
			res.Spec.ForProvider.AuthenticationFlowBindingOverrides[0].BrowserID = &u
		}
		a, _ := json.Marshal(ex)
		b, _ := json.Marshal(res)
		p, err := jsonpatch.CreateMergePatch(a, b)
		t.Logf("%s\n%s\npatch=%s err=%v", a, b, p, err)
	}
}
