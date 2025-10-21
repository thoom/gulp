package ui

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ghodss/yaml"
)

//go:embed static/*
var staticFiles embed.FS

// Template represents a discovered GULP configuration template
type Template struct {
	Path      string    `json:"path"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	Variables []string  `json:"variables"`
	Folder    string    `json:"folder"`
	Size      int64     `json:"size"`
	Modified  time.Time `json:"modified"`
	IsValid   bool      `json:"is_valid"`
	Error     string    `json:"error,omitempty"`
}

// TemplateRequest represents a request to execute a template
type TemplateRequest struct {
	TemplatePath string            `json:"template_path"`
	Variables    map[string]string `json:"variables"`
	URL          string            `json:"url,omitempty"`
	Method       string            `json:"method,omitempty"`
}

// ExecutionResponse represents the response from executing a template
type ExecutionResponse struct {
	Success        bool              `json:"success"`
	StatusCode     *int              `json:"status_code,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	Body           string            `json:"body"`
	Error          string            `json:"error,omitempty"`
	Duration       float64           `json:"duration"`
	RequestURL     string            `json:"request_url"`
	RequestHeaders map[string]string `json:"request_headers,omitempty"`
}

// Allows for mocking exec.Command in tests
var execCommand = exec.Command

var (
	statusCodeRegex = regexp.MustCompile(`\[GULP\] Status Code: (\d+)`)
	headersRegex    = regexp.MustCompile(`(?m)^> ([a-zA-Z\-]+): (.*)$`)
	requestURLRegex = regexp.MustCompile(`\[GULP\] Request URL: (.*)`)
)

// Server represents the UI web server
type Server struct {
	port       string
	templates  []Template
	workingDir string
	gulpBinary string // Path to the gulp binary to execute
	staticFS   fs.FS  // Filesystem for the static UI assets
}

// StartServer initializes and starts the web UI server
func StartServer(address string) error {
	// Create the filesystem for the embedded static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return fmt.Errorf("could not create static filesystem: %w", err)
	}

	server := &Server{
		gulpBinary: os.Args[0], // Default to current executable
		staticFS:   staticFS,
	}

	// Parse address
	if err := server.parseAddress(address); err != nil {
		return fmt.Errorf("invalid address '%s': %w", address, err)
	}

	// Get working directory
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get working directory: %w", err)
	}
	server.workingDir = workingDir

	// Discover templates
	if err := server.discoverTemplates(); err != nil {
		return fmt.Errorf("could not discover templates: %w", err)
	}

	// Setup routes
	server.setupRoutes()

	// Start server
	fmt.Printf("🚀 Visual GULP starting...\n")
	fmt.Printf("📁 Scanning for templates in: %s\n", workingDir)
	fmt.Printf("📋 Found %d templates\n", len(server.templates))
	fmt.Printf("🌐 Server running at: http://%s\n", server.getFullAddress())
	fmt.Printf("🔗 Open your browser to the URL above\n")

	return http.ListenAndServe(server.getFullAddress(), nil)
}

// parseAddress parses the address flag into host:port
func (s *Server) parseAddress(address string) error {
	if address == "" {
		s.port = "8080"
		return nil
	}

	// If no colon, treat as port only
	if !strings.Contains(address, ":") {
		s.port = address
		return nil
	}

	// Full address provided
	parts := strings.Split(address, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid format, expected 'port' or 'host:port'")
	}

	s.port = parts[1]
	return nil
}

// getFullAddress returns the full address for the server
func (s *Server) getFullAddress() string {
	return "localhost:" + s.port
}

// discoverTemplates scans the working directory for YAML/YML templates
func (s *Server) discoverTemplates() error {
	s.templates = []Template{}

	return filepath.WalkDir(s.workingDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Get relative path for checking
		relPath, err := filepath.Rel(s.workingDir, path)
		if err != nil {
			relPath = path
		}

		// Skip hidden directories and files
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip common directories that shouldn't contain templates
		if d.IsDir() {
			skipDirs := []string{"node_modules", "build", "dist", "vendor", ".git", ".svn", "target", "bin"}
			for _, skipDir := range skipDirs {
				if d.Name() == skipDir || strings.Contains(relPath, skipDir) {
					return filepath.SkipDir
				}
			}
		}

		// Only process YAML/YML files
		if !d.IsDir() && (strings.HasSuffix(strings.ToLower(d.Name()), ".yml") || strings.HasSuffix(strings.ToLower(d.Name()), ".yaml")) {
			template, err := s.parseTemplate(path)
			if err != nil {
				// Still add invalid templates but mark them as such
				template = Template{
					Path:    relPath,
					Name:    d.Name(),
					Folder:  filepath.Dir(relPath),
					IsValid: false,
					Error:   err.Error(),
				}
			}
			s.templates = append(s.templates, template)
		}

		return nil
	})
}

// parseTemplate reads and parses a template file
func (s *Server) parseTemplate(path string) (Template, error) {
	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		return Template{}, fmt.Errorf("could not read file: %w", err)
	}

	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return Template{}, fmt.Errorf("could not get file info: %w", err)
	}

	// Parse YAML to validate
	var config interface{}
	if err := yaml.Unmarshal(content, &config); err != nil {
		return Template{}, fmt.Errorf("invalid YAML: %w", err)
	}

	// Extract template variables
	variables := extractTemplateVariables(string(content))

	// Get relative path for display
	relPath, err := filepath.Rel(s.workingDir, path)
	if err != nil {
		relPath = path
	}

	return Template{
		Path:      relPath,
		Name:      info.Name(),
		Content:   string(content),
		Variables: variables,
		Folder:    filepath.Dir(relPath),
		Size:      info.Size(),
		Modified:  info.ModTime(),
		IsValid:   true,
	}, nil
}

// extractTemplateVariables finds all {{.Vars.variableName}} patterns in content
func extractTemplateVariables(content string) []string {
	// Regex to match {{.Vars.variableName}} patterns
	re := regexp.MustCompile(`\{\{\.Vars\.([a-zA-Z_][a-zA-Z0-9_]*)\}\}`)
	matches := re.FindAllStringSubmatch(content, -1)

	// Extract unique variable names
	varMap := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			varMap[match[1]] = true
		}
	}

	// Convert to sorted slice
	variables := make([]string, 0, len(varMap))
	for varName := range varMap {
		variables = append(variables, varName)
	}
	sort.Strings(variables)

	return variables
}

// setupRoutes configures HTTP routes
func (s *Server) setupRoutes() {
	// API routes
	http.HandleFunc("/api/templates", s.handleTemplates)
	http.HandleFunc("/api/template/", s.handleTemplate)
	http.HandleFunc("/api/execute", s.handleExecute)
	http.HandleFunc("/api/health", s.handleHealth)

	// Static file serving for React app
	// Serve React's static assets (JS, CSS files)
	staticAssets, err := fs.Sub(staticFiles, "static/static")
	if err == nil {
		fileServer := http.FileServer(http.FS(staticAssets))
		http.Handle("/static/", http.StripPrefix("/static/", fileServer))

		// Serve React app for all other routes (SPA routing)
		http.HandleFunc("/", s.handleReactApp)
	} else {
		// Fallback to simple HTML if embed fails (development)
		http.HandleFunc("/", s.handleRoot)
	}
}

// handleTemplates returns the list of discovered templates
func (s *Server) handleTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.templates)
}

// handleTemplate returns a single template by its path
func (s *Server) handleTemplate(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	for _, t := range s.templates {
		if t.Path == path {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(t)
			return
		}
	}
	// If no template is found
	s.sendError(w, "Template not found", http.StatusNotFound)
}

// handleExecute executes a template with the provided variables
func (s *Server) handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := s.executeTemplate(req)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// executeTemplate integrates with actual GULP execution
func (s *Server) executeTemplate(req TemplateRequest) ExecutionResponse {
	startTime := time.Now()

	// Find the full path for the template
	templateFullPath := filepath.Join(s.workingDir, req.TemplatePath)

	// Construct the command arguments
	args := []string{
		"-p", templateFullPath,
		"-u", req.URL,
		"-m", req.Method,
		"--json-output", // Ensure output is in JSON format
	}
	for key, value := range req.Variables {
		args = append(args, "-v", fmt.Sprintf("%s=%s", key, value))
	}

	cmd := execCommand(s.gulpBinary, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(startTime).Seconds()

	if err != nil {
		return ExecutionResponse{
			Success:  false,
			Error:    fmt.Sprintf("GULP execution failed: %s", stderr.String()),
			Duration: duration,
		}
	}

	// The `parseUIOutput` function is designed for interactive output, not JSON.
	// We will unmarshal the JSON directly.
	var execResponse ExecutionResponse
	if err := json.Unmarshal(out.Bytes(), &execResponse); err != nil {
		// Fallback for older GULP versions or unexpected output
		return s.parseUIOutput(out.String(), duration, req.URL)
	}

	// The JSON output from gulp might not include duration, so we set it here.
	execResponse.Duration = duration
	return execResponse
}

// parseUIOutput parses GULP UI output to extract request/response details
func (s *Server) parseUIOutput(output string, duration float64, fallbackURL string) ExecutionResponse {
	bodySeparator := "\n---\n"

	// Find the first occurrence of the separator
	sepIndex := strings.Index(output, bodySeparator)
	if sepIndex == -1 {
		return ExecutionResponse{Success: false, Error: "Could not parse GULP output"}
	}

	header := output[:sepIndex]
	body := strings.TrimSpace(output[sepIndex+len(bodySeparator):])

	// Extract status code
	statusCodeMatch := statusCodeRegex.FindStringSubmatch(header)
	statusCode := 200 // Default to 200 OK
	if len(statusCodeMatch) > 1 {
		if code, err := strconv.Atoi(statusCodeMatch[1]); err == nil {
			statusCode = code
		}
	}

	// Extract request URL
	requestURLMatch := requestURLRegex.FindStringSubmatch(header)
	requestURL := fallbackURL
	if len(requestURLMatch) > 1 {
		requestURL = requestURLMatch[1]
	}

	// Extract headers
	headers := make(map[string]string)
	headerMatches := headersRegex.FindAllStringSubmatch(header, -1)
	for _, match := range headerMatches {
		if len(match) == 3 {
			headers[match[1]] = match[2]
		}
	}

	return ExecutionResponse{
		Success:    true,
		StatusCode: &statusCode,
		Headers:    headers,
		Body:       body,
		Duration:   duration,
		RequestURL: requestURL,
	}
}

// handleHealth provides a health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":      "healthy",
		"templates":   len(s.templates),
		"working_dir": s.workingDir,
		"timestamp":   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleReactApp serves the React application
func (s *Server) handleReactApp(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// If the path is a directory, serve index.html.
	// This handles the root path "/" as well.
	if strings.HasSuffix(path, "/") {
		path = "index.html"
	}

	// Check if the file exists in the static filesystem.
	if _, err := fs.Stat(s.staticFS, strings.TrimPrefix(path, "/")); os.IsNotExist(err) {
		// If the file doesn't exist, serve the root index.html.
		// This is the key for single-page application routing.
		http.ServeFileFS(w, r, s.staticFS, "index.html")
		return
	}

	// Serve the file from the embedded filesystem.
	http.FileServer(http.FS(s.staticFS)).ServeHTTP(w, r)
}

// handleRoot serves a simple HTML page for fallback
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Visual GULP</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 2px solid #007acc; padding-bottom: 10px; }
        .endpoint { background: #f8f9fa; padding: 15px; margin: 10px 0; border-radius: 4px; border-left: 4px solid #007acc; }
        .method { background: #007acc; color: white; padding: 2px 6px; border-radius: 3px; font-size: 12px; }
        code { background: #e9ecef; padding: 2px 6px; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Visual GULP</h1>
        <p>Welcome to Visual GULP! The React frontend will be added soon.</p>
        
        <h2>Available API Endpoints</h2>
        
        <div class="endpoint">
            <span class="method">GET</span> <code>/api/templates</code>
            <p>List all discovered templates with metadata</p>
        </div>
        
        <div class="endpoint">
            <span class="method">GET</span> <code>/api/template/{path}</code>
            <p>Get specific template content and variables</p>
        </div>
        
        <div class="endpoint">
            <span class="method">POST</span> <code>/api/execute</code>
            <p>Execute GULP with template and variables</p>
        </div>
        
        <div class="endpoint">
            <span class="method">GET</span> <code>/api/health</code>
            <p>Health check and server status</p>
        </div>
        
        <h2>Template Discovery</h2>
        <p>Server is scanning: <code>` + s.workingDir + `</code></p>
        <p>Found <strong>` + fmt.Sprintf("%d", len(s.templates)) + `</strong> templates</p>
        
        <p><a href="/api/templates">View Templates JSON</a> | <a href="/api/health">Health Check</a></p>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// sendError sends a JSON error response
func (s *Server) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
