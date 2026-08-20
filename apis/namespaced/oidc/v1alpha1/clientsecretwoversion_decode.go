/*
Copyright 2022 Upbound Inc.
*/

package v1alpha1

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// coerceClientSecretWoVersion rewrites a string-encoded clientSecretWoVersion
// to the number encoding of this frozen API version. Main builds between the
// terraform-provider-keycloak v5.9.0 bump (588b321) and the v1alpha2 split
// (990f048) typed the field as a string at v1alpha1 and persisted such values,
// which would otherwise fail the strict decode in the conversion webhook and
// crash-loop the provider (see crossplane-contrib/provider-keycloak#669).
func coerceClientSecretWoVersion(data []byte) ([]byte, error) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}
	raw, ok := obj["clientSecretWoVersion"]
	if !ok {
		return data, nil
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("clientSecretWoVersion %q was stored as a non-numeric string by a pre-v3.0.0 development build and cannot be represented in the numeric v1alpha1 schema: update the field to a numeric value or recreate the object at v1alpha2", s)
		}
		obj["clientSecretWoVersion"] = json.RawMessage(strconv.FormatFloat(f, 'f', -1, 64))
		return json.Marshal(obj)
	}
	// Not a string, let the default decoding handle it.
	return data, nil
}

// UnmarshalJSON tolerates the string encoding of clientSecretWoVersion written
// at v1alpha1 by pre-v3.0.0 development builds.
func (in *IdentityProviderParameters) UnmarshalJSON(data []byte) error {
	data, err := coerceClientSecretWoVersion(data)
	if err != nil {
		return err
	}
	type noMethods IdentityProviderParameters
	return json.Unmarshal(data, (*noMethods)(in))
}

// UnmarshalJSON tolerates the string encoding of clientSecretWoVersion written
// at v1alpha1 by pre-v3.0.0 development builds.
func (in *IdentityProviderInitParameters) UnmarshalJSON(data []byte) error {
	data, err := coerceClientSecretWoVersion(data)
	if err != nil {
		return err
	}
	type noMethods IdentityProviderInitParameters
	return json.Unmarshal(data, (*noMethods)(in))
}

// UnmarshalJSON tolerates the string encoding of clientSecretWoVersion written
// at v1alpha1 by pre-v3.0.0 development builds.
func (in *IdentityProviderObservation) UnmarshalJSON(data []byte) error {
	data, err := coerceClientSecretWoVersion(data)
	if err != nil {
		return err
	}
	type noMethods IdentityProviderObservation
	return json.Unmarshal(data, (*noMethods)(in))
}
