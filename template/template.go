package template

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
)

// TemplateData holds variables for template processing
type TemplateData struct {
	Vars map[string]string
}

// ParseTemplateVars parses CLI variables in format "key=value" into a map
func ParseTemplateVars(vars []string) map[string]string {
	templateVars := make(map[string]string)
	for _, v := range vars {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			templateVars[parts[0]] = parts[1]
		}
	}
	return templateVars
}

// ProcessTemplate reads a template file, processes it with variables, and returns the result
func ProcessTemplate(templateFile string, vars []string) ([]byte, error) {
	if templateFile == "" {
		return nil, fmt.Errorf("template file path cannot be empty")
	}

	// Read the template file
	templateContent, err := os.ReadFile(templateFile)
	if err != nil {
		return nil, fmt.Errorf("could not read template file '%s': %v", templateFile, err)
	}

	// If no variables are provided, return the file content as-is (no template processing)
	if len(vars) == 0 {
		return templateContent, nil
	}

	// Parse template variables
	templateVars := ParseTemplateVars(vars)

	// Create template data
	data := TemplateData{
		Vars: templateVars,
	}

	// Parse and execute the template
	tmpl, err := template.New("payload").Parse(string(templateContent))
	if err != nil {
		return nil, fmt.Errorf("could not parse template file '%s': %v", templateFile, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("could not execute template file '%s': %v", templateFile, err)
	}

	return buf.Bytes(), nil
}

// ReadPayloadFile reads a JSON or YAML file for use as payload
func ReadPayloadFile(payloadFile string) ([]byte, error) {
	if payloadFile == "" {
		return nil, fmt.Errorf("payload file path cannot be empty")
	}

	content, err := os.ReadFile(payloadFile)
	if err != nil {
		return nil, fmt.Errorf("could not read payload file '%s': %v", payloadFile, err)
	}

	return content, nil
}

// ProcessStdin processes stdin content as a Go template with the provided variables
func ProcessStdin(content []byte, vars []string) ([]byte, error) {
	// If no variables are provided, return the content as-is
	if len(vars) == 0 {
		return content, nil
	}

	// Parse template variables
	templateVars := ParseTemplateVars(vars)

	// Create template data
	data := TemplateData{
		Vars: templateVars,
	}

	// Parse and execute the template
	tmpl, err := template.New("stdin").Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("could not parse stdin as template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("could not execute stdin template: %v", err)
	}

	return buf.Bytes(), nil
}

// ProcessInlineTemplate processes inline template content with direct variable access
func ProcessInlineTemplate(content string, vars []string) ([]byte, error) {
	// If no variables are provided, return the content as-is
	if len(vars) == 0 {
		return []byte(content), nil
	}

	// Parse template variables
	templateVars := ParseTemplateVars(vars)

	// Create template data
	data := TemplateData{
		Vars: templateVars,
	}

	// Parse and execute the template with proper data structure
	tmpl, err := template.New("inline").Parse(content)
	if err != nil {
		return nil, fmt.Errorf("could not parse inline template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("could not execute inline template: %v", err)
	}

	return buf.Bytes(), nil
}
