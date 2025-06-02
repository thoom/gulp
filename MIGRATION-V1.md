# GULP v1.0 Migration Guide

This guide helps you migrate from earlier versions of GULP to v1.0. **Note: v1.0 removes legacy features entirely** - this guide shows you exactly what needs to change.

## Breaking Changes Summary

- ❌ **Legacy flags removed**: `-ro`, `-sco` (single dash flags)
- ❌ **Configuration format changes**: `display` → `output`, `client_auth` → `auth.certificate`
- ❌ **Removed flag**: `use_color` (now controlled by `--output`)
- ✅ **NEW: Template system**: `--template` with `--var` for dynamic templates
- ✅ **NEW: Basic authentication**: `auth.basic` support in config
- ✅ **NEW: Form handling**: URL-encoded and multipart form support
- ✅ **NEW: cURL-like input**: Enhanced input handling
- ✅ **NEW: Method in config**: HTTP method can be specified in configuration
- ✅ **NEW: Data section**: Request body templating with `data:` section
- ✅ **Enhanced: Double-dash flags**: Modern `--output body|status|verbose` syntax

## Step-by-Step Migration

### 1. Update Command Line Usage

#### Output Flags (BREAKING CHANGE)

**❌ Old (REMOVED in v1.0):**
```bash
# These single-dash flags no longer exist
gulp -ro https://api.example.com          # REMOVED
gulp -sco https://api.example.com         # REMOVED
```

**✅ New (v1.0):**
```bash
# Use the new double-dash --output flag
gulp --output body https://api.example.com     # Response body only
gulp --output status https://api.example.com   # Status code only  
gulp --output verbose https://api.example.com  # Full details
```

#### Template Processing (NEW)

**❌ Old (Limited):**
```bash
# Basic stdin redirection only
gulp -m POST https://api.example.com < data.json
```

**✅ New (Dynamic Templates):**
```bash
# Dynamic templates with variables
gulp --template request.json \
  --var name="John" \
  --var env="prod" \
  --var timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -m POST https://api.example.com
```

Where `request.json` contains:
```json
{
  "user": "{{.Vars.name}}",
  "environment": "{{.Vars.env}}",
  "created_at": "{{.Vars.timestamp}}"
}
```

### 2. Migrate Configuration Files

#### Configuration Structure (BREAKING CHANGES)

**❌ Old Format (.gulp.yml):**
```yaml
# Pre-1.0 .gulp.yml format
url: https://api.example.com/users
headers:
  Authorization: Bearer token
  Content-Type: application/json
display: verbose  # or status-code-only
timeout: 300
client_auth:
  cert: /path/to/cert.pem
  key: /path/to/key.pem
  ca: /path/to/ca.pem
flags:
  follow_redirects: true
  use_color: true
  verify_tls: true
```

**✅ New Format (v1.0):**
```yaml
# v1.0 unified configuration
url: https://api.example.com/users
method: POST  # NEW: method in config

headers:
  Authorization: "Bearer {{.Vars.token}}"  # NEW: template support
  Content-Type: application/json

output: verbose  # CHANGED: was "display"

auth:  # CHANGED: structure changed + NEW basic auth
  basic:  # NEW: basic auth support
    username: "{{.Vars.user}}"
    password: "{{.Vars.pass}}"
  certificate:  # CHANGED: was "client_auth"
    cert: /path/to/cert.pem
    key: /path/to/key.pem
    ca: /path/to/ca.pem

data:  # NEW: data/template support
  body: |
    {
      "name": "{{.Vars.name}}",
      "email": "{{.Vars.email}}"
    }

flags:
  follow_redirects: true
  verify_tls: true
  # REMOVED: use_color (now controlled by --output)
```

#### Key Configuration Changes

1. **Display → Output (BREAKING)**:
   - Old: `display: verbose` or `display: status-code-only`
   - New: `output: verbose`, `output: body`, or `output: status`

2. **Client Auth → Auth Structure (BREAKING)**:
   - Old: `client_auth:` with cert/key/ca (certificate auth only)
   - New: `auth.certificate:` with cert/key/ca + NEW `auth.basic:` support

3. **Basic Authentication (NEW)**:
   - Pre-1.0 had no basic auth support
   - v1.0 adds `auth.basic:` with username/password

4. **Template Support (NEW)**:
   - Headers and data now support `{{.Vars.variableName}}` templates
   - Use with `--var variableName=value`

5. **Method in Config (NEW)**:
   - Can now specify HTTP method in configuration file
   - Was command-line only before

6. **Data Section (NEW)**:
   - `data:` section for request body with template support
   - Form data support with `data.form:`

### 3. Practical Migration Examples

#### Example 1: Simple API Call

**❌ Old Script:**
```bash
#!/bin/bash
# Static data via pipe
echo '{"name":"John","email":"john@example.com"}' | \
gulp -m POST \
  -ro \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  https://api.example.com/users
```

**✅ New Script:**
```bash
#!/bin/bash
# Create template file
cat > user-request.json << 'EOF'
{
  "name": "{{.Vars.name}}",
  "email": "{{.Vars.email}}",
  "created_at": "{{.Vars.timestamp}}"
}
EOF

# Use template with variables
gulp --template user-request.json \
  --var name="John" \
  --var email="john@example.com" \
  --var timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --output body \
  -H "Authorization: Bearer $TOKEN" \
  -m POST https://api.example.com/users
```

#### Example 2: Environment-Specific Configuration

**❌ Old Multi-Environment Setup:**
```bash
# prod.sh
gulp -m GET -sco -H "Authorization: Bearer $PROD_TOKEN" https://api.example.com/health

# dev.sh  
gulp -m GET -sco -H "Authorization: Bearer $DEV_TOKEN" https://dev-api.example.com/health
```

**✅ New Multi-Environment Setup:**

Create `api-config.yml`:
```yaml
url: https://{{.Vars.env}}-api.example.com/{{.Vars.endpoint}}
method: GET
output: status

headers:
  Authorization: "Bearer {{.Vars.token}}"
```

Use with any environment:
```bash
# Production
gulp -c api-config.yml --var env="api" --var endpoint="health" --var token="$PROD_TOKEN"

# Development  
gulp -c api-config.yml --var env="dev" --var endpoint="health" --var token="$DEV_TOKEN"

# Staging
gulp -c api-config.yml --var env="staging" --var endpoint="health" --var token="$STAGING_TOKEN"
```

#### Example 3: CI/CD Pipeline Migration

**❌ Old CI Pipeline:**
```bash
#!/bin/bash
# deploy.sh
echo "Building..."
BUILD_ID=$(date +%s)

# Health check with old flags
if gulp -sco https://api.example.com/health | grep -q "200"; then
  echo "API is healthy"
  
  # Deploy notification with static data via pipe
  echo "{\"build_id\":\"$BUILD_ID\",\"status\":\"deployed\"}" | \
  gulp -m POST \
    -ro \
    -H "Authorization: Bearer $CI_TOKEN" \
    -H "Content-Type: application/json" \
    https://api.example.com/deployments
else
  echo "API health check failed"
  exit 1
fi
```

**✅ New CI Pipeline:**

Create `ci-config.yml`:
```yaml
url: https://api.example.com/{{.Vars.endpoint}}
method: "{{.Vars.method}}"
output: "{{.Vars.output_mode}}"

headers:
  Authorization: "Bearer {{.Vars.ci_token}}"
  Content-Type: application/json
  X-CI-Build: "{{.Vars.build_id}}"

data:
  template: "@ci-deployment.json"
```

Create `ci-deployment.json`:
```json
{
  "build": {
    "id": "{{.Vars.build_id}}",
    "timestamp": "{{.Vars.timestamp}}",
    "branch": "{{.Vars.branch}}",
    "commit": "{{.Vars.commit_sha}}"
  },
  "deployment": {
    "status": "{{.Vars.status}}",
    "environment": "{{.Vars.target_env}}",
    "deployer": "{{.Vars.deployer}}"
  }
}
```

New script:
```bash
#!/bin/bash
# deploy.sh
BUILD_ID=$(date +%s)
TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Health check with new output mode
if gulp -c ci-config.yml \
  --var endpoint="health" \
  --var method="GET" \
  --var output_mode="status" \
  --var ci_token="$CI_TOKEN" \
  --var build_id="$BUILD_ID" | grep -q "200"; then
  
  echo "API is healthy"
  
  # Deploy notification with dynamic template
  gulp -c ci-config.yml \
    --var endpoint="deployments" \
    --var method="POST" \
    --var output_mode="body" \
    --var ci_token="$CI_TOKEN" \
    --var build_id="$BUILD_ID" \
    --var timestamp="$TIMESTAMP" \
    --var branch="$CI_BRANCH" \
    --var commit_sha="$CI_COMMIT_SHA" \
    --var status="deployed" \
    --var target_env="production" \
    --var deployer="$CI_USER"
else
  echo "API health check failed"
  exit 1
fi
```

### 4. Load Testing Migration

**❌ Old Load Testing (Limited):**
```bash
# Manual script with loops
for i in {1..100}; do
  gulp -sco https://api.example.com/health &
done
wait
```

**✅ New Load Testing (Built-in):**

Create `load-test.yml`:
```yaml
url: https://api.example.com/{{.Vars.endpoint}}
method: GET
output: status

headers:
  X-Load-Test: "true"
  X-Test-ID: "{{.Vars.test_id}}"

repeat:
  times: "{{.Vars.requests}}"
  concurrent: "{{.Vars.concurrent}}"
```

Use built-in load testing:
```bash
# Simple load test
gulp -c load-test.yml \
  --var endpoint="health" \
  --var test_id="load-$(date +%s)" \
  --var requests="100" \
  --var concurrent="10" | sort | uniq -c

# Output:
#   95 200
#    3 503  
#    2 502
```

### 5. Form Data Support (NEW FEATURE)

**✅ New Form Data (v1.0 Feature):**
```bash
# Native form support with file uploads
gulp -m POST \
  --form name="John Doe" \
  --form email="john@example.com" \
  --form department="Engineering" \
  --form resume=@resume.pdf \
  --form avatar=@photo.jpg \
  https://api.example.com/users
```

**Note**: Form data support is entirely new in v1.0. Previous versions had no form data capabilities.

## Migration Checklist

- [ ] Replace `-ro` with `--output body`
- [ ] Replace `-sco` with `--output status`
- [ ] Convert configuration files to new format:
  - [ ] Change `display: verbose` to `output: verbose`
  - [ ] Change `display: status-code-only` to `output: status`
  - [ ] Change `client_auth:` to `auth.certificate:`
  - [ ] Remove `use_color` flag (now controlled by `--output`)
  - [ ] Add `method:` to config if using POST/PUT/etc
  - [ ] Migrate to template syntax in headers if using variables
- [ ] Update scripts to use templates with `--template` and `--var` for dynamic content
- [ ] Replace manual load testing with `--repeat-times` and `--repeat-concurrent`
- [ ] Use new `--form` flags for form data (if applicable)
- [ ] Test all configurations and scripts with v1.0

## Migration Tools

### Quick Config Converter

```bash
#!/bin/bash
# convert-config.sh - Helper to convert old configs

if [ ! -f ".gulp.yml" ]; then
  echo "No .gulp.yml found"
  exit 1
fi

echo "Converting .gulp.yml to v1.0 format..."
cp .gulp.yml .gulp.yml.backup

# This is a basic example - adapt for your specific config
cat > .gulp.yml.new << 'EOF'
# Converted to GULP v1.0 format
url: https://api.example.com
method: GET
output: body

headers:
  Authorization: "Bearer {{.Vars.token}}"
  Content-Type: application/json

data:
  body: |
    {
      "message": "{{.Vars.message}}",
      "timestamp": "{{.Vars.timestamp}}"
    }

flags:
  follow_redirects: true
  verify_tls: true
EOF

echo "New config created as .gulp.yml.new"
echo "Review and rename when ready: mv .gulp.yml.new .gulp.yml"
```

### Script Updater

```bash
#!/bin/bash
# update-scripts.sh - Helper to find old flag usage

echo "Checking for old GULP flags in scripts..."

# Find old single-dash flags
grep -r --include="*.sh" --include="*.bash" -n "\-ro\|\-sco" .

echo "Update these to use --output body or --output status"
```

## Troubleshooting Migration

### Common Issues

1. **"flag provided but not defined" errors**
   - Old flags like `-ro`, `-sco` no longer exist
   - Replace with `--output body`, `--output status`

2. **Configuration file errors**
   - String booleans like `"true"` must be real booleans: `true`
   - Update auth structure from `client_auth` to `auth.basic`

3. **Template variables not working**
   - Use `{{.Vars.variableName}}` syntax in templates
   - Pass variables with `--var variableName=value`

4. **Form data not working**
   - Use `--form key=value` instead of manual encoding
   - Use `--form file=@path` for file uploads

For additional help, see the [examples directory](examples/) for working v1.0 configurations and templates. 