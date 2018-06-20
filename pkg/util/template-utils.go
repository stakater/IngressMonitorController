package util

import (
	"bytes"
	"html/template"

	"github.com/pkg/errors"
)

type MonitorNameTemplateParts struct {
	IngressName string
	Namespace   string
}

func GetNameTemplateFormat(nameTemplate string) (string, error) {
	if nameTemplate == "" {
		nameTemplate = "{{.IngressName}}-{{.Namespace}}"
	}
	placeholders := MonitorNameTemplateParts{"%[1]s", "%[2]s"}
	tmpl, err := template.New("format").Parse(nameTemplate)
	if err != nil {
		errors.Wrap(err, "Failed to parse nameTempalte")
	}
	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, placeholders)
	if err != nil {
		errors.Wrap(err, "Failed to execute nameTempalte")
	}
	return buffer.String(), nil
}
