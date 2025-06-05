## GULP v1.0 Examples

This directory contains practical examples showcasing GULP v1.0's new features and capabilities.

## Quick Navigation

- [Template Examples](#template-examples)
- [Configuration Examples](#configuration-examples)
- [Load Testing Examples](#load-testing-examples)
- [Real-World Use Cases](#real-world-use-cases)
- [Migration Examples](#migration-examples)

## Template Examples

### Basic User Creation Template

**File**: `user-creation.json`
```json
{
  "user": {
    "name": "{{.Vars.name}}",
    "email": "{{.Vars.email}}",
    "role": "{{.Vars.role}}",
    "department": "{{.Vars.department}}"
  },
  "metadata": {
    "created_at": "{{.Vars.timestamp}}",
    "created_by": "{{.Vars.admin_user}}",
    "environment": "{{.Vars.env}}",
    "source": "gulp-v1"
  },
  "settings": {
    "send_welcome_email": {{.Vars.send_email}},
    "require_password_reset": {{.Vars.require_reset}},
    "access_level": {{.Vars.access_level}}
  }
}
```

**Usage**:
```bash
gulp --template @examples/user-creation.json \
  --var name="Alice Johnson" \
  --var email="alice@company.com" \
  --var role="developer" \
  --var department="engineering" \
  --var timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --var admin_user="admin" \
  --var env="production" \
  --var send_email=true \
  --var require_reset=false \
  --var access_level=3 \
  -m POST --url https://api.example.com/users
```

### Conditional Environment Template

**File**: `environment-config.json`
```json
{
  "application": "{{.Vars.app_name}}",
  "environment": "{{.Vars.env}}",
  "config": {
    {{if eq .Vars.env "production"}}
    "log_level": "error",
    "debug_mode": false,
    "cache_ttl": 3600,
    "max_connections": 100
    {{else if eq .Vars.env "staging"}}
    "log_level": "warn", 
    "debug_mode": false,
    "cache_ttl": 1800,
    "max_connections": 50
    {{else}}
    "log_level": "debug",
    "debug_mode": true,
    "cache_ttl": 300,
    "max_connections": 10
    {{end}}
  },
  "features": [
    "core"
    {{if .Vars.feature_auth}}, "authentication"{{end}}
    {{if .Vars.feature_monitoring}}, "monitoring"{{end}}
    {{if .Vars.feature_cache}}, "caching"{{end}}
  ],
  "deployed_at": "{{.Vars.timestamp}}"
}
```

## Configuration Examples

### Complete API Configuration

**File**: `api-config.yml`
```yaml
# Complete API configuration showcasing all v1.0 features
url: https://api.{{.Vars.environment}}.company.com/{{.Vars.endpoint}}
method: "{{.Vars.http_method}}"
timeout: "{{.Vars.timeout}}s"
output: "{{.Vars.output_mode}}"

# Dynamic headers with environment-specific tokens
headers:
  Authorization: "Bearer {{.Vars.api_token}}"
  Content-Type: application/json
  User-Agent: "GULP/1.0 ({{.Vars.environment}})"
  X-Request-ID: "{{.Vars.request_id}}"
  X-Client-Version: "{{.Vars.client_version}}"
  X-Environment: "{{.Vars.environment}}"

# Authentication with environment variables
auth:
  basic:
    username: "{{.Vars.api_username}}"
    password: "{{.Vars.api_password}}"

# Template-driven request data
data:
  template: "@examples/api-request.json"
  variables:
    service_name: "user-service"
    api_version: "v1"
    region: "us-east-1"

# Environment-specific settings
request:
  insecure: false
  follow_redirects: true

# Load testing configuration
repeat:
  times: "{{.Vars.repeat_count}}"
  concurrent: "{{.Vars.concurrent_count}}"

# Feature flags
flags:
  follow_redirects: true
  use_color: true
  verify_tls: true
```

**Usage**:
```bash
# Development environment
gulp -c examples/api-config.yml \
  --var environment="dev" \
  --var endpoint="users" \
  --var http_method="GET" \
  --var timeout="30" \
  --var output_mode="verbose" \
  --var api_token="$DEV_TOKEN" \
  --var request_id="$(uuidgen)" \
  --var client_version="1.0.0" \
  --var api_username="dev-user" \
  --var api_password="dev-pass" \
  --var repeat_count="1" \
  --var concurrent_count="1"

# Production load test
gulp -c examples/api-config.yml \
  --var environment="prod" \
  --var endpoint="health" \
  --var http_method="GET" \
  --var timeout="10" \
  --var output_mode="status" \
  --var api_token="$PROD_TOKEN" \
  --var request_id="load-test-$(date +%s)" \
  --var client_version="1.0.0" \
  --var repeat_count="100" \
  --var concurrent_count="10"
```

### Multi-Environment Configuration

**File**: `multi-env-config.yml`
```yaml
# Base configuration that works across environments
url: https://{{.Vars.env}}-api.company.com/{{.Vars.service}}/{{.Vars.endpoint}}
method: POST
output: body

headers:
  Authorization: "{{.Vars.auth_header}}"
  Content-Type: application/json
  X-Environment: "{{.Vars.env}}"
  X-Service: "{{.Vars.service}}"

data:
  body: |
    {
      "environment": "{{.Vars.env}}",
      "service": "{{.Vars.service}}",
      "action": "{{.Vars.action}}",
      "payload": {{.Vars.payload}},
      "metadata": {
        "timestamp": "{{.Vars.timestamp}}",
        "source": "gulp-automation",
        "version": "{{.Vars.version}}"
      }
    }
  
  variables:
    version: "1.0.0"
    source: "automated-test"

request:
  follow_redirects: true
```

## Load Testing Examples

### Performance Testing Configuration

**File**: `load-test.yml`
```yaml
# Comprehensive load testing configuration
url: https://api.example.com/{{.Vars.endpoint}}
method: "{{.Vars.method}}"
output: status

headers:
  Authorization: "Bearer {{.Vars.token}}"
  Content-Type: application/json
  X-Load-Test: "true"
  X-Test-ID: "{{.Vars.test_id}}"

data:
  body: |
    {
      "test_data": {
        "id": "{{.Vars.test_id}}",
        "timestamp": "{{.Vars.timestamp}}",
        "iteration": "{{.Vars.iteration}}",
        "payload_size": "{{.Vars.payload_size}}"
      },
      "metadata": {
        "test_type": "load_test",
        "concurrent_users": {{.Vars.concurrent}},
        "total_requests": {{.Vars.total}}
      }
    }

repeat:
  times: "{{.Vars.total}}"
  concurrent: "{{.Vars.concurrent}}"

request:
  follow_redirects: true
```

**Load Testing Script**: `run-load-test.sh`
```bash
#!/bin/bash

# Load testing with increasing concurrency
echo "Starting load tests..."

TEST_ID="load-test-$(date +%s)"
API_TOKEN="${API_TOKEN:-test-token}"

for concurrent in 5 10 20 50; do
  echo "Testing with $concurrent concurrent connections..."
  
  start_time=$(date +%s)
  
  gulp -c examples/load-test.yml \
    --var endpoint="api/v1/test" \
    --var method="POST" \
    --var token="$API_TOKEN" \
    --var test_id="$TEST_ID-$concurrent" \
    --var timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    --var iteration="1" \
    --var payload_size="medium" \
    --var concurrent="$concurrent" \
    --var total="100" | \
    sort | uniq -c > "load-test-$concurrent.results"
  
  end_time=$(date +%s)
  duration=$((end_time - start_time))
  
  echo "Results for $concurrent concurrent connections:"
  cat "load-test-$concurrent.results"
  echo "Duration: ${duration}s"
  echo "---"
done

echo "Load testing complete. Check *.results files for detailed data."
```

## Real-World Use Cases

### CI/CD Pipeline Integration

**File**: `ci-pipeline.yml`
```yaml
# CI/CD pipeline configuration
url: https://ci-api.company.com/{{.Vars.pipeline}}/{{.Vars.action}}
method: POST
output: verbose

headers:
  Authorization: "Bearer {{.Vars.ci_token}}"
  Content-Type: application/json
  X-CI-System: "{{.Vars.ci_system}}"
  X-Build-ID: "{{.Vars.build_id}}"

data:
  template: "@examples/ci-payload.json"
  variables:
    ci_system: "github-actions"
    pipeline_version: "v2"

repeat:
  times: 1
  concurrent: 1

flags:
  follow_redirects: true
  verify_tls: true
```

**CI Payload Template**: `ci-payload.json`
```json
{
  "build": {
    "id": "{{.Vars.build_id}}",
    "number": "{{.Vars.build_number}}",
    "branch": "{{.Vars.branch}}",
    "commit": "{{.Vars.commit_sha}}",
    "author": "{{.Vars.author}}"
  },
  "pipeline": {
    "name": "{{.Vars.pipeline}}",
    "stage": "{{.Vars.stage}}",
    "action": "{{.Vars.action}}",
    "environment": "{{.Vars.target_env}}"
  },
  "metadata": {
    "ci_system": "{{.ci_system}}",
    "pipeline_version": "{{.pipeline_version}}",
    "timestamp": "{{.Vars.timestamp}}",
    "webhook_url": "{{.Vars.webhook_url}}"
  },
  "artifacts": [
    {{range $i, $artifact := .Vars.artifacts}}
    {{if $i}},{{end}}
    {
      "name": "{{$artifact.name}}",
      "url": "{{$artifact.url}}",
      "type": "{{$artifact.type}}"
    }
    {{end}}
  ]
}
```

### API Monitoring Script

**File**: `monitoring.yml`
```yaml
# API monitoring configuration
url: https://{{.Vars.service}}.{{.Vars.domain}}/{{.Vars.endpoint}}
method: GET
output: status
timeout: "{{.Vars.timeout}}s"

headers:
  User-Agent: "GULP-Monitor/1.0"
  X-Monitor-Check: "health"
  X-Check-ID: "{{.Vars.check_id}}"

repeat:
  times: "{{.Vars.checks}}"
  concurrent: 1

request:
  follow_redirects: true
```

**Monitoring Script**: `monitor.sh`
```bash
#!/bin/bash

# API monitoring with GULP v1.0
SERVICES=("auth" "users" "orders" "payments")
DOMAIN="api.company.com"
TIMEOUT=10
CHECKS=3

for service in "${SERVICES[@]}"; do
  echo "Checking $service service..."
  
  CHECK_ID="monitor-$(date +%s)-$service"
  
  gulp -c examples/monitoring.yml \
    --var service="$service" \
    --var domain="$DOMAIN" \
    --var endpoint="health" \
    --var timeout="$TIMEOUT" \
    --var check_id="$CHECK_ID" \
    --var checks="$CHECKS" > /tmp/$service-health.txt
  
  if grep -q "200" /tmp/$service-health.txt; then
    echo "✅ $service: OK"
  else
    echo "❌ $service: FAILED"
    cat /tmp/$service-health.txt
  fi
done
```

### Data Migration Example

**File**: `data-migration.yml`
```yaml
# Data migration configuration
url: https://migration-api.company.com/{{.Vars.operation}}
method: POST
output: body
timeout: "300s"

headers:
  Authorization: "Bearer {{.Vars.migration_token}}"
  Content-Type: application/json
  X-Migration-ID: "{{.Vars.migration_id}}"
  X-Source-System: "{{.Vars.source_system}}"
  X-Target-System: "{{.Vars.target_system}}"

data:
  template: "@examples/migration-payload.json"
  variables:
    migration_type: "user_data"
    batch_size: 1000
    validation_level: "strict"

repeat:
  times: 1
  concurrent: 1

request:
  follow_redirects: false
```

## v1.0 Usage Examples

### Simple Request with Templates

**Using templates for dynamic requests**:
```bash
gulp --template @examples/user-creation.json \
  --var name="John" \
  --var email="john@example.com" \
  --var role="user" \
  --var department="sales" \
  --var timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --var admin_user="system" \
  --var env="production" \
  --var send_email=true \
  --var require_reset=false \
  --var access_level=1 \
  -m POST --url https://api.example.com/users
```

**Using configuration files**:
```bash
gulp -c examples/api-config.yml \
  --var environment="prod" \
  --var endpoint="users" \
  --var http_method="POST" \
  --var api_token="$TOKEN"
```

## Running the Examples

### Prerequisites

1. Install GULP v1.0
2. Set environment variables:
   ```bash
   export API_TOKEN="your-api-token"
   export DEV_TOKEN="dev-token"
   export PROD_TOKEN="prod-token"
   ```

### Example Commands

```bash
# Basic template usage
gulp --template @examples/user-creation.json \
  --var name="Test User" \
  --var email="test@example.com" \
  --var role="tester" \
  --var department="qa" \
  --var timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --var admin_user="admin" \
  --var env="staging" \
  --var send_email=false \
  --var require_reset=true \
  --var access_level=2 \
  -m POST --url https://httpbin.org/post

# Configuration-driven request
gulp -c examples/api-config.yml \
  --var environment="staging" \
  --var endpoint="test" \
  --var http_method="GET" \
  --var timeout="30" \
  --var output_mode="verbose" \
  --var api_token="test-token" \
  --var request_id="$(uuidgen)" \
  --var client_version="1.0.0"

# Load testing
gulp -c examples/load-test.yml \
  --var endpoint="post" \
  --var method="POST" \
  --var token="test-token" \
  --var test_id="example-$(date +%s)" \
  --var timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --var concurrent="5" \
  --var total="20"
```

## Tips and Best Practices

1. **Use version control** for your templates and configurations
2. **Environment variables** for sensitive data like tokens
3. **Template validation** - test templates with sample data first
4. **Modular templates** - break complex requests into smaller, reusable templates
5. **Documentation** - comment your YAML configurations for team collaboration
6. **Load testing** - start with small concurrent values and increase gradually
7. **Monitoring** - use status output mode for automated monitoring scripts

For more examples and advanced use cases, see the [Migration Guide](../MIGRATION-V1.md). 