# GULP ![Builds](https://github.com/thoom/gulp/actions/workflows/main.yml/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/thoom/gulp)](https://goreportcard.com/report/github.com/thoom/gulp) [![Coverage](https://sonarcloud.io/api/project_badges/measure?project=gulp&metric=coverage)](https://sonarcloud.io/summary/overall?id=gulp) [![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=gulp&metric=security_rating)](https://sonarcloud.io/summary/overall?id=gulp) [![GoDoc](https://godoc.org/github.com/thoom/gulp?status.svg)](https://godoc.org/github.com/thoom/gulp)

GULP is a powerful HTTP client designed for API testing and automation. It supports JSON, YAML, form data, templates, and more.

> **🎉 New in v1.0!** Template processing, unified configuration format, load testing, and more! See the [**Migration Guide**](MIGRATION-V1.md) for comprehensive examples and upgrade instructions.

**Key Features:**
- 🚀 **Fast and lightweight** - Single binary with no dependencies
- 🔧 **Template processing** - Go templates with variable substitution (NEW in v1.0)
- 📄 **Unified configuration** - Method, URL, and data all in one config file (NEW in v1.0)
- 📝 **Multiple data formats** - JSON, YAML, form data, file uploads
- ⚡ **Load testing** - Concurrent requests and repeat functionality (NEW in v1.0)
- 🔐 **Comprehensive authentication** - Basic auth, client certificates, custom CA
- 🎨 **Rich output options** - Verbose, status-only, or body-only modes
- 📄 **Flexible configuration** - YAML config files with clean structure
- 🔄 **Pipeline friendly** - Works great with jq, curl, and other tools

## Quick Start

```bash
# Simple GET request
gulp https://api.example.com
# or with --url flag (preferred by many users)
gulp --url https://api.example.com

# POST with JSON data
echo '{"name": "John", "age": 30}' | gulp -m POST --url https://api.example.com

# Using templates with variables (v1.0)
gulp --template @request.json --var env=prod --var user=john -m POST --url https://api.example.com

# Complete configuration file (v1.0)
gulp -c config.yml
```

## Installation

### Binary Releases (Recommended)

Download the latest binary for your platform from [GitHub Releases](https://github.com/thoom/gulp/releases).

### Using Docker

```bash
# Basic usage
docker run --rm -it -v $PWD:/gulp ghcr.io/thoom/gulp

# With configuration file
docker run --rm -it -v $PWD:/gulp ghcr.io/thoom/gulp -c config.yml https://api.example.com
```

### Using Go

```bash
go install github.com/thoom/gulp@latest
```

## Core Usage

### Basic Requests (v1.0)

```bash
# GET request
gulp https://api.github.com/users/octocat
# or using --url flag
gulp --url https://api.github.com/users/octocat

# POST with inline JSON (v1.0)
gulp -m POST --body '{"message": "Hello World"}' https://api.example.com
# or with --url flag
gulp -m POST --body '{"message": "Hello World"}' --url https://api.example.com

# PUT with file content (v1.0)
gulp -m PUT --body @data.json https://api.example.com/resource/123
# or with --url flag
gulp -m PUT --body @data.json --url https://api.example.com/resource/123

# Custom headers
gulp -H "Authorization: Bearer token" -H "Content-Type: application/json" https://api.example.com
# or with --url flag
gulp -H "Authorization: Bearer token" -H "Content-Type: application/json" --url https://api.example.com
```

### Template Processing (v1.0)

One of the most powerful new features in GULP v1.0 is the template system:

#### Basic Template Example

Create a template file `api-request.json`:
```json
{
  "user": {
    "name": "{{.Vars.username}}",
    "email": "{{.Vars.email}}",
    "role": "{{.Vars.role}}"
  },
  "metadata": {
    "timestamp": "{{.Vars.timestamp}}",
    "environment": "{{.Vars.env}}",
    "api_version": "v1"
  },
  "settings": {
    "debug_mode": {{.Vars.debug}},
    "max_retries": {{.Vars.retries}}
  }
}
```

Use the template:
```bash
gulp --template @api-request.json \
  --var username="John Doe" \
  --var email="john@example.com" \
  --var role="admin" \
  --var timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --var env="production" \
  --var debug=false \
  --var retries=3 \
  -m POST --url https://api.example.com/users
```

#### Advanced Template with Conditionals

Create `conditional-request.json`:
```json
{
  "user": "{{.Vars.username}}",
  "environment": "{{.Vars.env}}",
  {{if eq .Vars.env "production"}}
  "logging_level": "error",
  "debug_enabled": false
  {{else}}
  "logging_level": "debug", 
  "debug_enabled": true
  {{end}},
  "features": [
    "core"{{if .Vars.feature_auth}}, "authentication"{{end}}{{if .Vars.feature_cache}}, "caching"{{end}}
  ]
}
```

### Data Input Methods

#### JSON/YAML Body Data

```bash
# Inline JSON
gulp -m POST --body '{"user": "john", "role": "admin"}' --url https://api.example.com

# From file
gulp -m POST --body @data.json --url https://api.example.com

# From stdin
echo '{"test": true}' | gulp -m POST --url https://api.example.com

# YAML (automatically converted to JSON)
gulp -m POST --body @data.yml --url https://api.example.com
```

#### Form Data and File Uploads (Enhanced in v1.0)

```bash
# Simple form fields
gulp -m POST --form name=John --form age=30 --url https://api.example.com

# File uploads
gulp -m POST --form avatar=@profile.jpg --form name=John --url https://api.example.com

# Mixed form data with JSON
gulp -m POST \
  --form name=John \
  --form email=john@example.com \
  --form metadata='{"role": "admin", "department": "IT"}' \
  --form document=@contract.pdf \
  --url https://api.example.com/users

# Form mode with stdin
echo "name=John&email=john@example.com&active=true" | \
  gulp --form-mode -m POST --url https://api.example.com
```

### Authentication

#### Basic Authentication

```bash
# Convenient format
gulp --auth-basic=username:password --url https://api.example.com

# Separate flags
gulp --basic-auth-user=username --basic-auth-pass=password --url https://api.example.com
```

#### Client Certificate Authentication

```bash
# Using certificate files
gulp --client-cert=client.pem --client-cert-key=key.pem --url https://api.example.com

# With custom CA
gulp --client-cert=client.pem --client-cert-key=key.pem --custom-ca=ca.pem --url https://api.example.com
```

### Output Options (Enhanced in v1.0)

```bash
# Default: show response body only
gulp --url https://api.example.com

# Verbose: show headers, timing, and body
gulp --output verbose --url https://api.example.com
# or
gulp -v --url https://api.example.com

# Status code only
gulp --output status --url https://api.example.com

# Disable colors
gulp --no-color --url https://api.example.com
```

### Load Testing (New in v1.0)

```bash
# Make 100 requests
gulp --repeat-times 100 --url https://api.example.com

# 100 requests with 10 concurrent connections
gulp --repeat-times 100 --repeat-concurrent 10 --url https://api.example.com

# Load test with status monitoring
gulp --repeat-times 50 --repeat-concurrent 5 --output status --url https://api.example.com | sort | uniq -c
```

## Unified Configuration Files (v1.0)

GULP v1.0 introduces a powerful unified configuration format that can contain everything needed for an API request:

### Complete Configuration Example

```yaml
# .gulp.yml - Complete v1.0 configuration
url: https://api.example.com/{{.Vars.endpoint}}
method: POST
timeout: "60s"
output: verbose

# HTTP Headers with template variables
headers:
  Authorization: "Bearer {{.Vars.api_token}}"
  Content-Type: application/json
  User-Agent: GULP/1.0
  X-Request-ID: "{{.Vars.request_id}}"
  X-Environment: "{{.Vars.environment}}"

# Authentication configuration
auth:
  basic:
    username: "{{.Vars.api_user}}"
    password: "{{.Vars.api_pass}}"
  
  certificate:
    cert: /path/to/client.pem
    key: /path/to/key.pem
    ca: /path/to/ca.pem

# Request data - multiple options
data:
  # Option 1: Inline body with template variables
  body: |
    {
      "user": {
        "name": "{{.Vars.user_name}}",
        "email": "{{.Vars.user_email}}",
        "role": "{{.Vars.user_role}}"
      },
      "metadata": {
        "created_at": "{{.timestamp}}",
        "environment": "{{.environment}}",
        "request_id": "{{.Vars.request_id}}"
      }
    }
  
  # Template variables
  variables:
    timestamp: "2024-01-15T10:30:00Z"
    environment: production
    
  # Option 2: External template file
  # template: "@templates/user-creation.json"
  
  # Option 3: Form data (alternative to JSON body)
  # form:
  #   name: "John Doe"
  #   email: "john@example.com"
  #   avatar: "@profile.jpg"

# Request settings
request:
  insecure: false
  follow_redirects: true

# Load testing configuration
repeat:
  times: 5
  concurrent: 2

# Feature flags (real booleans, not strings)
flags:
  follow_redirects: true
  use_color: true
  verify_tls: true
```

### Using the Complete Configuration

```bash
# Use complete configuration with runtime variables
gulp -c .gulp.yml \
  --var endpoint="users" \
  --var api_token="your-api-token" \
  --var request_id="$(uuidgen)" \
  --var environment="production" \
  --var user_name="Alice Johnson" \
  --var user_email="alice@company.com" \
  --var user_role="developer"

# Override configuration values
gulp -c .gulp.yml \
  --var api_token="dev-token" \
  --var environment="development" \
  --url https://api-dev.example.com/users \
  --method PATCH
```

### Environment-Specific Configurations

You can create different configuration files for different environments:

```bash
# Development
gulp -c config/development.yml --var dev_token="dev-123"

# Staging  
gulp -c config/staging.yml --var staging_token="staging-456"

# Production
gulp -c config/production.yml --var prod_token="prod-789"
```

## Command Line Reference

GULP organizes its flags into logical groups for better usability:

### Core Options
- `-m, --method` - HTTP method (GET, POST, PUT, DELETE)
- `-v, --verbose` - Show detailed request/response info
- `-c, --config` - Configuration file (.gulp.yml)

### Data Input (Enhanced in v1.0)
- `--body` - Request body (@file, @-, or inline)
- `--template` - Process file as Go template (NEW)
- `--var` - Template variable (repeat for multiple) (NEW)
- `--form` - Form field key=value or key=@file (NEW)
- `--form-mode` - Process stdin as form data (NEW)

### Authentication
- `--auth-basic` - Basic authentication user:pass
- `--basic-auth-user` - Basic auth username
- `--basic-auth-pass` - Basic auth password
- `--client-cert` - Client certificate file
- `--client-cert-key` - Client certificate key file
- `--custom-ca` - Custom CA certificate file

### Request Options
- `-H, --header` - Request header (repeat for multiple)
- `--timeout` - Request timeout in seconds (default 300)
- `--insecure` - Disable TLS certificate verification
- `--url` - Request URL (alternative to positional)

### Output & Display (Enhanced in v1.0)
- `--output` - Output mode: body, status, verbose (NEW)
- `--no-color` - Disable colored output

### Redirect Options
- `--follow-redirects` - Enable following redirects
- `--no-redirects` - Disable following redirects

### Load Testing (New in v1.0)
- `--repeat-times` - Number of requests to make (NEW)
- `--repeat-concurrent` - Number of concurrent connections (NEW)

## Advanced Examples

### API Testing Workflow with Templates

```bash
# 1. Get authentication token using a template
TOKEN=$(gulp --template @auth-request.json \
  --var username="$API_USER" \
  --var password="$API_PASS" \
  -m POST https://api.example.com/auth | jq -r '.token')

# 2. Use token for authenticated requests with configuration
gulp -c api-config.yml \
  --var token="$TOKEN" \
  --var user_id="12345" \
  --var action="update_profile"

# 3. Load test the API
gulp -c load-test-config.yml \
  --var token="$TOKEN" \
  --var concurrent=20 \
  --var requests=1000
```

### File Processing Pipeline

```bash
# Process CSV data through API
cat users.csv | while IFS=, read name email; do
  gulp --body "{\"name\":\"$name\",\"email\":\"$email\"}" \
    -m POST https://api.example.com/users
done

# Batch upload with form data
for file in *.jpg; do
  gulp -m POST --form image=@"$file" --form caption="Photo: $file" \
    https://api.example.com/photos
done
```

### Configuration-Driven Testing

```yaml
# test-environments.yml
url: https://{{.Vars.env}}.api.example.com
method: POST
output: status

auth:
  basic:
    username: "{{.Vars.username}}"
    password: "{{.Vars.password}}"

data:
  template: "@health-check.json"
  variables:
    service: api-gateway
    version: "1.0.0"

repeat:
  times: 10
  concurrent: 2
```

```bash
# Test different environments
for env in dev staging prod; do
  echo "Testing $env environment..."
  gulp -c test-environments.yml \
    --var env=$env \
    --var username=testuser \
    --var password=testpass
done
```

### Load Testing with Monitoring

```bash
# Performance test with detailed timing
gulp --repeat-times 1000 --repeat-concurrent 20 \
  --output verbose \
  https://api.example.com/health 2>&1 | \
  grep "Status:" | \
  awk '{print $4}' | \
  sort -n | \
  awk '{a[NR]=$1} END {print "Min:", a[1], "Max:", a[NR], "Median:", a[int(NR/2)]}'
```

## Integration Examples

### CI/CD Pipeline

```bash
#!/bin/bash
# api-test.sh

# Health check
if ! gulp --output status https://api.example.com/health | grep -q "200"; then
  echo "API health check failed"
  exit 1
fi

# Functional tests
gulp -c test-config.yml --var env=staging --repeat-times 5
```

### Monitoring Script

```bash
#!/bin/bash
# monitor.sh

while true; do
  STATUS=$(gulp --output status https://api.example.com/health)
  if [ "$STATUS" != "200" ]; then
    echo "$(date): API down - Status: $STATUS"
    # Send alert
  fi
  sleep 30
done
```

### Docker Integration

```dockerfile
# Dockerfile for API testing
FROM ghcr.io/thoom/gulp:latest
COPY test-config.yml /app/
COPY templates/ /app/templates/
WORKDIR /app
ENTRYPOINT ["gulp", "-c", "test-config.yml"]
```

## What's New in v1.0

### Major New Features

- **🔧 Template Processing System** - Go templates with variable substitution using `--template` and `--var`
- **📄 Unified Configuration Format** - Method, URL, headers, auth, and data all in one clean YAML file
- **⚡ Load Testing** - Built-in concurrency with `--repeat-times` and `--repeat-concurrent`
- **🎨 Enhanced Output Modes** - New `--output` flag with body/status/verbose options
- **🔄 Form Data Support** - Native form handling with `--form` and file upload support
- **📋 Better Help Organization** - Grouped flags for improved usability

### Migration from Earlier Versions

**Upgrading to v1.0?** Check out the comprehensive [**Migration Guide**](MIGRATION-V1.md) with:
- Detailed before/after examples
- Step-by-step migration instructions  
- Real-world CI/CD pipeline examples
- Breaking changes and compatibility notes

## Troubleshooting

### Common Issues

**Template variables not substituting:**
```bash
# Make sure to use .Vars prefix in templates
# Template: "user": "{{.Vars.username}}"
# Usage: --var username=john
```

**Form uploads failing:**
```bash
# Ensure file exists and is readable
gulp -m POST --form file=@/path/to/file.pdf --url https://api.example.com
```

**TLS verification errors:**
```bash
# For self-signed certificates
gulp --insecure --url https://self-signed.example.com
# Or provide custom CA
gulp --custom-ca=ca.pem --url https://api.example.com
```

### Debug Mode

Use verbose output to debug requests:
```bash
gulp -v -m POST --body '{"test": true}' --url https://api.example.com
```

## Library Dependencies

- `github.com/fatih/color` - Terminal colors
- `github.com/ghodss/yaml` - YAML processing
- `github.com/spf13/cobra` - CLI framework
- `github.com/stretchr/testify` - Testing (dev only)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Run `go test ./...` to ensure tests pass
5. Submit a pull request

## License

MIT License - see LICENSE file for details.
