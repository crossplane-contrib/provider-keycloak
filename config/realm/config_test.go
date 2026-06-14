package realm

import (
	"testing"
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
