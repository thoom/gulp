# GULP v1.0 Feature Showcase

## 🎉 Major New Features

### 1. Template Processing System
**Go templates with variable substitution**

```bash
# Create dynamic requests with templates
gulp --template user-template.json \
  --var name="John Doe" \
  --var env="production" \
  --var timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -m POST https://api.example.com/users
```

Template example (`user-template.json`):
```json
{
  "user": "{{.Vars.name}}",
  "environment": "{{.Vars.env}}",
  "created_at": "{{.Vars.timestamp}}"
}
```

### 2. Unified Configuration Format
**Everything in one clean YAML file**

```yaml
# Complete configuration with method, URL, auth, and data
url: https://api.{{.Vars.env}}.company.com/users
method: POST
timeout: "30s"
output: verbose

headers:
  Authorization: "Bearer {{.Vars.token}}"
  Content-Type: application/json

auth:
  basic:
    username: "{{.Vars.api_user}}"
    password: "{{.Vars.api_pass}}"

data:
  template: "@request-template.json"
  variables:
    service: "user-api"
    version: "v1"

repeat:
  times: 5
  concurrent: 2
```

### 3. Load Testing
**Built-in concurrency and repeat functionality**

```bash
# Performance test with 100 requests, 10 concurrent
gulp --repeat-times 100 --repeat-concurrent 10 \
  --output status https://api.example.com/health | sort | uniq -c

# Output:
#   95 200
#    3 503
#    2 502
```

### 4. Enhanced Output Modes
**Clean, organized output options**

```bash
# Body only (default)
gulp https://api.example.com/users

# Status code only  
gulp --output status https://api.example.com/health

# Verbose with full details
gulp --output verbose https://api.example.com/users
```

### 5. Form Data & File Uploads
**Native form handling with file support**

```bash
# Mixed form data with file uploads
gulp -m POST \
  --form name="John Doe" \
  --form avatar=@profile.jpg \
  --form metadata='{"role": "admin"}' \
  https://api.example.com/users
```

## 🏆 Quality Improvements

- **89% Test Coverage** - Comprehensive unit tests
- **Better Error Messages** - More helpful error reporting  
- **Enhanced Documentation** - Detailed examples and migration guide
- **Organized Help System** - Grouped flags for better usability

## 📚 Documentation

- **[Migration Guide](MIGRATION-V1.md)** - Comprehensive upgrade instructions with before/after examples
- **[Examples Directory](examples/)** - Practical, real-world use cases and templates
- **[Main README](README.md)** - Complete feature documentation and usage guide

## 🚀 Quick Start

```bash
# Simple template usage
gulp --template examples/user-creation.json \
  --var name="Alice" \
  --var email="alice@company.com" \
  --var role="developer" \
  --var department="engineering" \
  --var timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --var admin_user="admin" \
  --var env="production" \
  --var send_email=true \
  --var require_reset=false \
  --var access_level=3 \
  -m POST https://api.example.com/users

# Configuration-driven approach
gulp -c examples/api-config.yml \
  --var environment="staging" \
  --var endpoint="users" \
  --var http_method="GET" \
  --var api_token="your-token"

# Load testing
gulp -c examples/load-test.yml \
  --var endpoint="health" \
  --var method="GET" \
  --var token="test-token" \
  --var concurrent="10" \
  --var total="100"
```

## 🔄 Migration from Earlier Versions

### Before (pre-v1.0):
```bash
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com"}' \
  https://api.example.com/users
```

### After (v1.0):
```bash
gulp -c user-config.yml \
  --var token="$TOKEN" \
  --var name="John" \
  --var email="john@example.com"
```

Where `user-config.yml` contains everything needed:
```yaml
url: https://api.example.com/users
method: POST
headers:
  Authorization: "Bearer {{.Vars.token}}"
data:
  body: |
    {
      "name": "{{.Vars.name}}",
      "email": "{{.Vars.email}}"
    }
```

---

**Ready to upgrade?** Check the [Migration Guide](MIGRATION-V1.md) for step-by-step instructions and comprehensive examples. 

## Try the Examples

```bash
# Test template processing with variables  
gulp --template examples/user-creation.json \
     --var name="Jane Smith" \
     --var email="jane@example.com" \
     --var department="Engineering" \
     https://httpbin.org/post

# Test unified configuration
gulp --config examples/api-config.yml

# Test load testing
gulp --config examples/load-test.yml
``` 