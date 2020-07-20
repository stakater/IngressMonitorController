package util

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/pkg/errors"
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
		errors.Wrap(err, "Failed to parse nameTemplate")
	}

	if tmpl == nil {
		return "", fmt.Errorf("Failed to parse nameTemplate")
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, placeholders)
	if err != nil {
		errors.Wrap(err, "Failed to execute nameTemplate")
	}
	return buffer.String(), nil
}
