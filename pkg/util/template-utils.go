package util

import (
	"bytes"
	"html/template"
)

const (
	defaultNameTemplate = "{{.Name}}-{{.Namespace}}"
)

type MonitorNameTemplateParts struct {
	Name      string
	Namespace string
}

func GetNameTemplateFormat(nameTemplate string) (string, error) {
	if nameTemplate == "" {
		nameTemplate = defaultNameTemplate
	}
	placeholders := MonitorNameTemplateParts{"%[1]s", "%[2]s"}

	tmpl, err := template.New("format").Parse(nameTemplate)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, placeholders)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
