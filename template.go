package go2json

import (
	"bytes"
	"text/template"
)

// RenderTemplate renders a Go template string with the provided data.
// Returns the rendered string or an error if the template is invalid or execution fails.
func RenderTemplate(tmpl string, data interface{}) (string, error) {
	t, err := template.New("tpl").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RenderTemplateNoError renders a Go template with the given data, ignoring any errors.
// Use this when you're certain the template is valid and want simpler error handling.
func RenderTemplateNoError(tmpl string, data interface{}) string {
	res, _ := RenderTemplate(tmpl, data)
	return res
}

// MustRenderTemplate renders a template with the given data and panics if the template
// is invalid or execution fails.
func MustRenderTemplate(tmpl string, data interface{}) string {
	t, err := template.New("tpl").Parse(tmpl)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		panic(err)
	}

	return buf.String()
}
