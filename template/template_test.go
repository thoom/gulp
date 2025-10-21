package template

import (
	"os"
	"reflect"
	"testing"
)

func TestParseTemplateVars(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected map[string]string
	}{
		{
			name:     "Valid Variables",
			input:    []string{"name=test", "version=1.0"},
			expected: map[string]string{"name": "test", "version": "1.0"},
		},
		{
			name:     "Variables with Whitespace",
			input:    []string{"  key  =  value  "},
			expected: map[string]string{"key": "value"},
		},
		{
			name:     "Malformed Variable",
			input:    []string{"no-equal-sign"},
			expected: map[string]string{},
		},
		{
			name:     "Empty Input",
			input:    []string{},
			expected: map[string]string{},
		},
		{
			name:     "Mixed Valid and Invalid",
			input:    []string{"valid=true", "invalid"},
			expected: map[string]string{"valid": "true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseTemplateVars(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseTemplateVars() got = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestProcessTemplate(t *testing.T) {
	// Create a temporary template file
	tmpFile, err := os.CreateTemp("", "template-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	templateContent := "Hello, {{.Vars.name}}!"
	tmpFile.WriteString(templateContent)
	tmpFile.Close()

	t.Run("Successful Processing", func(t *testing.T) {
		vars := []string{"name=World"}
		result, err := ProcessTemplate(tmpFile.Name(), vars)
		if err != nil {
			t.Fatalf("ProcessTemplate() returned an unexpected error: %v", err)
		}
		expected := "Hello, World!"
		if string(result) != expected {
			t.Errorf("ProcessTemplate() got = %s, want %s", result, expected)
		}
	})

	t.Run("No Variables", func(t *testing.T) {
		result, err := ProcessTemplate(tmpFile.Name(), nil)
		if err != nil {
			t.Fatalf("ProcessTemplate() returned an unexpected error: %v", err)
		}
		if string(result) != templateContent {
			t.Errorf("ProcessTemplate() should return original content when no vars are provided")
		}
	})

	t.Run("File Not Found", func(t *testing.T) {
		_, err := ProcessTemplate("/no/such/file.txt", nil)
		if err == nil {
			t.Fatal("ProcessTemplate() expected an error for a non-existent file but got none")
		}
	})

	t.Run("Empty file path", func(t *testing.T) {
		_, err := ProcessTemplate("", nil)
		if err == nil {
			t.Fatal("ProcessTemplate() expected an error for empty file path but got none")
		}
	})
}

func TestReadPayloadFile(t *testing.T) {
	// Create a temporary payload file
	tmpFile, err := os.CreateTemp("", "payload-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	payloadContent := `{"key":"value"}`
	tmpFile.WriteString(payloadContent)
	tmpFile.Close()

	t.Run("Successful Read", func(t *testing.T) {
		result, err := ReadPayloadFile(tmpFile.Name())
		if err != nil {
			t.Fatalf("ReadPayloadFile() returned an unexpected error: %v", err)
		}
		if string(result) != payloadContent {
			t.Errorf("ReadPayloadFile() got = %s, want %s", result, payloadContent)
		}
	})

	t.Run("File Not Found", func(t *testing.T) {
		_, err := ReadPayloadFile("/no/such/payload.json")
		if err == nil {
			t.Fatal("ReadPayloadFile() expected an error for a non-existent file but got none")
		}
	})

	t.Run("Empty file path", func(t *testing.T) {
		_, err := ReadPayloadFile("")
		if err == nil {
			t.Fatal("ReadPayloadFile() expected an error for empty file path but got none")
		}
	})
}

func TestProcessStdin(t *testing.T) {
	content := []byte("User: {{.Vars.user}}")
	vars := []string{"user=gopher"}
	result, err := ProcessStdin(content, vars)
	if err != nil {
		t.Fatalf("ProcessStdin() returned an unexpected error: %v", err)
	}
	expected := "User: gopher"
	if string(result) != expected {
		t.Errorf("ProcessStdin() got = %s, want %s", result, expected)
	}
}

func TestProcessInlineTemplate(t *testing.T) {
	content := "ID: {{.Vars.id}}"
	vars := []string{"id=123"}
	result, err := ProcessInlineTemplate(content, vars)
	if err != nil {
		t.Fatalf("ProcessInlineTemplate() returned an unexpected error: %v", err)
	}
	expected := "ID: 123"
	if string(result) != expected {
		t.Errorf("ProcessInlineTemplate() got = %s, want %s", result, expected)
	}
}
