/*
Copyright 2024 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tf2crossplane

import "strings"

// sanitizeName converts a Terraform local name into an RFC 1123 compliant
// Kubernetes object name (lowercase alphanumerics and '-').
func sanitizeName(in string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(in) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			b.WriteRune(r)
		case r == '_' || r == '.' || r == ' ':
			b.WriteRune('-')
		default:
			// drop any other character
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = "resource"
	}
	// Names must start with an alphanumeric character.
	if out[0] == '-' {
		out = strings.TrimLeft(out, "-")
	}
	return out
}

// providerConfigScaffold and credentialsSecretScaffold mirror
// examples/providerconfig and examples/credentials. Secret values are
// intentionally placeholders and must be filled in by the user.
const providerConfigScaffold = `apiVersion: keycloak.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: keycloak-config
  namespace: crossplane-system
spec:
  credentials:
    source: Secret
    secretRef:
      name: keycloak-credentials
      key: credentials
      namespace: crossplane-system
`

const credentialsSecretScaffold = `apiVersion: v1
kind: Secret
metadata:
  name: keycloak-credentials
  namespace: crossplane-system
type: Opaque
stringData:
  # Replace the placeholder values below. Never commit real secrets.
  credentials: |
    {
      "client_id": "xxxxxx",
      "client_secret": "xxxxxx",
      "url": "https://keycloak.example.com",
      "base_path": "",
      "realm": "master"
    }
`
