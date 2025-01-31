package utils

import (
	"bytes"
	"github.com/crossplane/upjet/pkg/resource/json"
	"strings"
	"text/template"
)

// GetExternalNameFromTemplate returns the external name of a resource based on a template and tfState
func GetExternalNameFromTemplate(externalNameTemplate string, tfState map[string]any) (string, error) {
	t, err := template.New("getExternalName").Funcs(template.FuncMap{
		"ToLower": strings.ToLower,
		"ToUpper": strings.ToUpper,
	}).Parse(externalNameTemplate)

	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, tfState)
	if err != nil {
		return "", err
	}
	externalName := buf.String()
	return externalName, nil
}

// UnmarshalTerraformParamsToObject unmarshal´s the terraform parameters into the ParametersObject
func UnmarshalTerraformParamsToObject(parameters map[string]any, v interface{}) error {
	p, err := json.TFParser.Marshal(parameters)
	if err != nil {
		return err
	}
	err = json.TFParser.Unmarshal(p, &v)
	if err != nil {
		return err
	}
	return nil
}
