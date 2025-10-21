package form

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseFormData(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *FormData
		hasError bool
	}{
		{
			name:  "Valid Fields and Files",
			input: "name=test\nemail=test@example.com\nfile=@/tmp/file.txt",
			expected: &FormData{
				Fields: map[string]string{"name": "test", "email": "test@example.com"},
				Files:  map[string]string{"file": "/tmp/file.txt"},
			},
			hasError: false,
		},
		{
			name:  "Only Fields",
			input: "key1=value1\nkey2=value2",
			expected: &FormData{
				Fields: map[string]string{"key1": "value1", "key2": "value2"},
				Files:  map[string]string{},
			},
			hasError: false,
		},
		{
			name:  "Only Files",
			input: "resume=@/path/to/resume.pdf\nphoto=@/path/to/photo.jpg",
			expected: &FormData{
				Fields: map[string]string{},
				Files:  map[string]string{"resume": "/path/to/resume.pdf", "photo": "/path/to/photo.jpg"},
			},
			hasError: false,
		},
		{
			name:  "Empty and Whitespace Lines",
			input: "\n  key=value\n\n  file=@/tmp/data  \n",
			expected: &FormData{
				Fields: map[string]string{"key": "value"},
				Files:  map[string]string{"file": "/tmp/data"},
			},
			hasError: false,
		},
		{
			name:     "Malformed Line",
			input:    "malformed-line",
			expected: &FormData{Fields: map[string]string{}, Files: map[string]string{}},
			hasError: false, // Malformed lines are skipped
		},
		{
			name:     "Empty Input",
			input:    "",
			expected: &FormData{Fields: map[string]string{}, Files: map[string]string{}},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form, err := ParseFormData([]byte(tt.input))

			if (err != nil) != tt.hasError {
				t.Errorf("ParseFormData() error = %v, wantErr %v", err, tt.hasError)
				return
			}

			if !reflect.DeepEqual(form, tt.expected) {
				t.Errorf("ParseFormData() got = %v, want %v", form, tt.expected)
			}
		})
	}
}

func TestFormData_ToURLEncoded(t *testing.T) {
	t.Run("Valid URL Encoding", func(t *testing.T) {
		form := &FormData{
			Fields: map[string]string{"name": "John Doe", "email": "johndoe@example.com"},
			Files:  map[string]string{},
		}
		encoded, err := form.ToURLEncoded()
		if err != nil {
			t.Fatalf("ToURLEncoded() returned an unexpected error: %v", err)
		}
		// Note: url.Values.Encode() sorts keys
		expected := "email=johndoe%40example.com&name=John+Doe"
		if string(encoded) != expected {
			t.Errorf("ToURLEncoded() got = %s, want %s", encoded, expected)
		}
	})

	t.Run("Error on File Uploads", func(t *testing.T) {
		form := &FormData{
			Fields: map[string]string{"name": "test"},
			Files:  map[string]string{"file": "/tmp/file.txt"},
		}
		_, err := form.ToURLEncoded()
		if err == nil {
			t.Fatal("ToURLEncoded() expected an error but got none")
		}
	})
}

func TestFormData_ToMultipart(t *testing.T) {
	// Create a temporary file for testing file uploads
	tmpFile, err := os.CreateTemp("", "testfile-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("hello world")
	tmpFile.Close()

	form := &FormData{
		Fields: map[string]string{"user": "gopher"},
		Files:  map[string]string{"upload": tmpFile.Name()},
	}

	body, contentType, err := form.ToMultipart()
	if err != nil {
		t.Fatalf("ToMultipart() returned an unexpected error: %v", err)
	}

	if !strings.HasPrefix(contentType, "multipart/form-data; boundary=") {
		t.Errorf("Expected multipart content type, got %s", contentType)
	}

	bodyStr := string(body)
	if !strings.Contains(bodyStr, `Content-Disposition: form-data; name="user"`) ||
		!strings.Contains(bodyStr, "gopher") {
		t.Error("Multipart body does not contain the form field")
	}

	if !strings.Contains(bodyStr, `Content-Disposition: form-data; name="upload"; filename="`+filepath.Base(tmpFile.Name())+`"`) ||
		!strings.Contains(bodyStr, "hello world") {
		t.Error("Multipart body does not contain the file content")
	}
}

func TestProcessFormData(t *testing.T) {
	// Setup temp file for multipart test
	tmpFile, err := os.CreateTemp("", "test-process-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("content")
	tmpFile.Close()

	tests := []struct {
		name                string
		input               string
		expectedContentType string
		expectError         bool
	}{
		{
			name:                "URL Encoded",
			input:               "field1=value1",
			expectedContentType: "application/x-www-form-urlencoded",
			expectError:         false,
		},
		{
			name:                "Multipart",
			input:               "field1=value1\nfile1=@" + tmpFile.Name(),
			expectedContentType: "multipart/form-data",
			expectError:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, contentType, err := ProcessFormData([]byte(tt.input))
			if (err != nil) != tt.expectError {
				t.Fatalf("ProcessFormData() error = %v, wantErr %v", err, tt.expectError)
			}

			if !strings.HasPrefix(contentType, tt.expectedContentType) {
				t.Errorf("ProcessFormData() contentType = %s, wantPrefix %s", contentType, tt.expectedContentType)
			}
		})
	}
}
