package form

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// FormData represents form field data
type FormData struct {
	Fields map[string]string
	Files  map[string]string // field name -> file path
}

// ParseFormData parses key=value pairs and file@path pairs from input
func ParseFormData(data []byte) (*FormData, error) {
	form := &FormData{
		Fields: make(map[string]string),
		Files:  make(map[string]string),
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if err := parseFormLine(form, line); err != nil {
			return nil, err
		}
	}

	return form, nil
}

// parseFormLine parses a single form line and adds it to the appropriate collection
func parseFormLine(form *FormData, line string) error {
	// Split into key=value format
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return nil // Skip malformed lines
	}

	key := parts[0]
	value := parts[1]

	// Check if it's a file upload (value starts with @)
	if isFileUpload(value) {
		filePath := strings.TrimPrefix(value, "@")
		form.Files[key] = filePath
	} else {
		// Regular field
		form.Fields[key] = value
	}

	return nil
}

// isFileUpload determines if a value represents a file upload
func isFileUpload(value string) bool {
	return strings.HasPrefix(value, "@")
}

// ToURLEncoded converts form data to application/x-www-form-urlencoded format
func (f *FormData) ToURLEncoded() ([]byte, error) {
	if len(f.Files) > 0 {
		return nil, fmt.Errorf("file uploads not supported with URL encoding, use multipart form data instead")
	}

	values := url.Values{}
	for key, value := range f.Fields {
		values.Set(key, value)
	}

	return []byte(values.Encode()), nil
}

// ToMultipart converts form data to multipart/form-data format
func (f *FormData) ToMultipart() ([]byte, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add regular fields
	for key, value := range f.Fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, "", fmt.Errorf("failed to write field %s: %v", key, err)
		}
	}

	// Add file fields
	for fieldName, filePath := range f.Files {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, "", fmt.Errorf("failed to open file %s: %v", filePath, err)
		}

		part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
		if err != nil {
			file.Close() // Ensure file is closed before returning
			return nil, "", fmt.Errorf("failed to create form file for %s: %v", fieldName, err)
		}

		if _, err := io.Copy(part, file); err != nil {
			file.Close() // Ensure file is closed before returning
			return nil, "", fmt.Errorf("failed to copy file %s: %v", filePath, err)
		}

		if err := file.Close(); err != nil {
			return nil, "", fmt.Errorf("failed to close file %s: %v", filePath, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("failed to close multipart writer: %v", err)
	}

	return buf.Bytes(), writer.FormDataContentType(), nil
}

// ProcessFormData determines the appropriate form encoding and returns the body and content type
func ProcessFormData(data []byte) ([]byte, string, error) {
	form, err := ParseFormData(data)
	if err != nil {
		return nil, "", err
	}

	// Use multipart if there are file uploads, otherwise use URL encoding
	if len(form.Files) > 0 {
		body, contentType, err := form.ToMultipart()
		return body, contentType, err
	}

	body, err := form.ToURLEncoded()
	if err != nil {
		return nil, "", err
	}

	return body, "application/x-www-form-urlencoded", nil
}
