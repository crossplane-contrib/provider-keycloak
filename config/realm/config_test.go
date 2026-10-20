package realm

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestValidateDurationString(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{name: "valid seconds", value: "300s", wantErr: false},
		{name: "valid minutes", value: "5m", wantErr: false},
		{name: "valid hours", value: "1h", wantErr: false},
		{name: "valid composite", value: "1h30m", wantErr: false},
		{name: "valid zero", value: "0s", wantErr: false},
		{name: "valid full format", value: "10h0m0s", wantErr: false},
		{name: "empty string", value: "", wantErr: false},
		{name: "invalid bare number", value: "900", wantErr: true},
		{name: "invalid unit", value: "5x", wantErr: true},
		{name: "invalid text", value: "invalid", wantErr: true},
		{name: "invalid with spaces", value: "5 m", wantErr: true},
		{name: "wrong type", value: 123, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateDurationString(tt.value, "test_field")
			if tt.wantErr && len(errs) == 0 {
				t.Errorf("validateDurationString(%v) expected error, got none", tt.value)
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("validateDurationString(%v) unexpected error: %v", tt.value, errs)
			}
		})
	}
}

func TestPreserveUnconfiguredFlowBindings(t *testing.T) {
	flowFields := []string{
		"browser_flow",
		"registration_flow",
		"direct_grant_flow",
		"reset_credentials_flow",
		"client_authentication_flow",
		"docker_authentication_flow",
		"first_broker_login_flow",
	}
	resourceSchema := make(map[string]*schema.Schema, len(flowFields))
	observedValues := make(map[string]string, len(flowFields))
	for _, name := range flowFields {
		resourceSchema[name] = &schema.Schema{Type: schema.TypeString, Optional: true, Computed: true}
		observedValues[name] = "observed-" + name
	}
	resourceData := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{
		"browser_flow": "configured-browser",
	})
	observed := &keycloak.Realm{
		BrowserFlow:              stringPointer(observedValues["browser_flow"]),
		RegistrationFlow:         stringPointer(observedValues["registration_flow"]),
		DirectGrantFlow:          stringPointer(observedValues["direct_grant_flow"]),
		ResetCredentialsFlow:     stringPointer(observedValues["reset_credentials_flow"]),
		ClientAuthenticationFlow: stringPointer(observedValues["client_authentication_flow"]),
		DockerAuthenticationFlow: stringPointer(observedValues["docker_authentication_flow"]),
		FirstBrokerLoginFlow:     stringPointer(observedValues["first_broker_login_flow"]),
	}

	if err := preserveUnconfiguredFlowBindings(resourceData, observed); err != nil {
		t.Fatalf("preserveUnconfiguredFlowBindings() error = %v", err)
	}

	if got := resourceData.Get("browser_flow"); got != "configured-browser" {
		t.Errorf("configured browser_flow = %v, want configured-browser", got)
	}
	for _, name := range flowFields[1:] {
		if got, want := resourceData.Get(name), observedValues[name]; got != want {
			t.Errorf("%s = %v, want observed value %s", name, got, want)
		}
	}
}

func stringPointer(value string) *string {
	return &value
}
