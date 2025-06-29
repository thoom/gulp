package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"testing/fstest"
)

// newTestServer creates a server instance for testing purposes.
func newTestServer(t *testing.T) (*Server, func()) {
	tmpDir, err := os.MkdirTemp("", "gulp-ui-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	templateContent := "url: http://example.com"
	if err := os.WriteFile(filepath.Join(tmpDir, "test.yml"), []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write dummy template: %v", err)
	}

	dummyExec, err := os.CreateTemp("", "dummy-gulp-exec")
	if err != nil {
		t.Fatalf("Failed to create dummy executable: %v", err)
	}
	dummyExec.Close()

	server := &Server{
		port:       "8081",
		workingDir: tmpDir,
		gulpBinary: dummyExec.Name(),
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
		os.Remove(dummyExec.Name())
	}

	return server, cleanup
}

func TestServer_parseAddress(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		wantPort string
		wantErr  bool
	}{
		{"Empty Address", "", "8080", false},
		{"Port Only", "9090", "9090", false},
		{"Full Address", "localhost:1234", "1234", false},
		{"Invalid Format", "localhost:1234:56", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{}
			err := s.parseAddress(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && s.port != tt.wantPort {
				t.Errorf("parseAddress() port = %v, want %v", s.port, tt.wantPort)
			}
		})
	}
}

func TestServer_discoverTemplates(t *testing.T) {
	server, cleanup := newTestServer(t)
	defer cleanup()

	os.Mkdir(filepath.Join(server.workingDir, ".hidden"), 0755)
	os.WriteFile(filepath.Join(server.workingDir, ".hidden", "ignored.yml"), []byte("..."), 0644)
	os.WriteFile(filepath.Join(server.workingDir, "not-a-template.txt"), []byte("..."), 0644)
	os.Mkdir(filepath.Join(server.workingDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(server.workingDir, "subdir", "sub.yaml"), []byte("url: http://sub.com"), 0644)
	os.WriteFile(filepath.Join(server.workingDir, "another.yml"), []byte("key: '{{.Invalid}}'"), 0644)

	if err := server.discoverTemplates(); err != nil {
		t.Fatalf("discoverTemplates() failed: %v", err)
	}

	if len(server.templates) != 3 {
		t.Errorf("discoverTemplates() expected 3 templates, got %d", len(server.templates))
	}
}

func TestExtractTemplateVariables(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{"Simple Var", "Hello, {{.Vars.name}}!", []string{"name"}},
		{"Multiple Vars", "{{.Vars.user}} and {{.Vars.pass}}", []string{"pass", "user"}},
		{"No Vars", "Just a plain string.", []string{}},
		{"Duplicate Vars", "{{.Vars.id}} and {{.Vars.id}}", []string{"id"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTemplateVariables(tt.content)
			sort.Strings(got)
			sort.Strings(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractTemplateVariables() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_handleHealth(t *testing.T) {
	server, cleanup := newTestServer(t)
	defer cleanup()

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.handleHealth(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var healthResponse map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &healthResponse); err != nil {
		t.Fatalf("Failed to unmarshal health response: %v", err)
	}

	if status, ok := healthResponse["status"]; !ok || status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%v'", status)
	}
}

func TestServer_handleTemplates(t *testing.T) {
	server, cleanup := newTestServer(t)
	defer cleanup()

	server.discoverTemplates()

	req, err := http.NewRequest("GET", "/api/templates", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.handleTemplates)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "test.yml") {
		t.Errorf("handler body does not contain expected template: got %s", rr.Body.String())
	}
}

func TestServer_handleExecute(t *testing.T) {
	server, cleanup := newTestServer(t)
	defer cleanup()

	// Mock the exec.Command call
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = func(command string, args ...string) *exec.Cmd {
		// This is the mock implementation. It will "succeed" and write JSON to stdout.
		cmd := exec.Command("echo", `{"success": true, "body": "mocked response"}`)
		return cmd
	}

	server.discoverTemplates()

	// Create a request to execute the template
	reqBody := `{"template_path": "test.yml", "variables": {"key": "value"}}`
	req, err := http.NewRequest("POST", "/api/execute", strings.NewReader(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.handleExecute)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var execResponse ExecutionResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &execResponse); err != nil {
		t.Fatalf("Failed to unmarshal execution response: %v", err)
	}

	if !execResponse.Success || execResponse.Body != "mocked response" {
		t.Errorf("Unexpected execution response: got %+v", execResponse)
	}
}

func TestServer_handleReactApp(t *testing.T) {
	server, cleanup := newTestServer(t)
	defer cleanup()

	// Create a mock filesystem for the UI assets
	mockFS := fstest.MapFS{
		"index.html":  {Data: []byte("<title>GULP</title>")},
		"asset-1.txt": {Data: []byte("asset 1")},
	}
	server.staticFS = mockFS

	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handleReactApp)

	t.Run("Serves index.html for root", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		if !strings.Contains(rr.Body.String(), "<title>GULP</title>") {
			t.Error("Did not serve index.html for root path")
		}
	})

	t.Run("Serves existing static asset", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/asset-1.txt", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		if rr.Body.String() != "asset 1" {
			t.Errorf("Did not serve correct asset content, got: %s", rr.Body.String())
		}
	})

	t.Run("Serves index.html as fallback for non-existent path", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/non-existent/path", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code for fallback: got %v want %v", status, http.StatusOK)
		}
		if !strings.Contains(rr.Body.String(), "<title>GULP</title>") {
			t.Error("Did not serve index.html as a fallback")
		}
	})
}
