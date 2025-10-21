# GULP Build System v1.0+

This document describes the standardized build system for GULP v1.0 and later versions.

## Overview

Starting with v1.0, GULP uses a comprehensive Makefile-based build system that replaces the previous shell scripts. This provides better consistency, dependency management, and easier maintenance.

## Quick Start

```bash
# Build for current platform with timestamp snapshot
make build

# Build for all platforms
make build-all RELEASE_VERSION=1.2.3

# Full release pipeline
make release RELEASE_VERSION=1.2.3
```

## Build Targets

### Development Builds
- `make build` - Build for current platform with timestamp snapshot version
- `make snapshot` - Explicit snapshot build with timestamp  
- `make build-version V=1.2.3` - Build with custom version
- `make run` - Build and run the application

### Release Builds  
- `make build-all RELEASE_VERSION=1.2.3` - Build for all platforms
- `make release RELEASE_VERSION=1.2.3` - Complete release pipeline

### Docker
- `make docker-build RELEASE_VERSION=1.2.3` - Build Docker image
- `make docker-deploy RELEASE_VERSION=1.2.3` - Build and deploy to registries
- `make docker-clean` - Remove Docker images

### Development Tools
- `make test` - Run tests
- `make test-coverage` - Run tests with coverage report
- `make fmt` - Format code
- `make lint` - Run linter (requires golangci-lint)
- `make deps` - Install/update dependencies

### Utilities
- `make clean` - Remove build artifacts
- `make version` - Build and show version
- `make show-snapshot-version` - Show what snapshot version would be generated
- `make help` - Show all available targets

## Version Handling

### Snapshot Versions
Snapshot builds automatically generate timestamps in the format: `YYYYMMDD.HHMMAM/PM.TZ-SNAPSHOT`

Example: `20250602.1133AM.MDT-SNAPSHOT`

This is generated **at build time**, not runtime, ensuring consistent versions for each build.

### Release Versions
Release versions are specified via the `RELEASE_VERSION` environment variable:

```bash
make build-all RELEASE_VERSION=1.2.3
```

## Multi-Platform Builds

The `build-all` target creates binaries for:
- Linux 386 (`gulp.linux-386.tar.gz`)
- Linux AMD64 (`gulp.linux-amd64.tar.gz`) 
- Darwin AMD64 (`gulp.darwin-amd64.tar.gz`)
- Darwin ARM64 (`gulp.darwin-arm64.tar.gz`)
- Windows AMD64 (`gulp.windows.zip`)

All builds are placed in the `./build/` directory.

## Docker Deployment

Docker deployment supports both Docker Hub and GitHub Container Registry:

```bash
# Set environment variables
export DOCKER_USER=your-dockerhub-username
export DOCKER_PASS=your-dockerhub-token
export GH_USER=your-github-username  
export GH_PASS=your-github-token
export IMAGE_NAME=your-org/gulp

# Deploy
make docker-deploy RELEASE_VERSION=1.2.3
```

## CI/CD Integration

The GitHub Actions workflow (`.github/workflows/main.yml`) automatically uses the Makefile targets:

- **Build**: `make build-all`
- **Deploy**: `make docker-deploy`

## Migration from Shell Scripts

The following shell scripts have been replaced:

| Old Script | New Makefile Target |
|------------|-------------------|
| `scripts/build.sh` | `make build-all` |
| `scripts/deploy.sh` | `make docker-deploy` |

## Environment Variables

| Variable | Description | Required For |
|----------|-------------|--------------|
| `RELEASE_VERSION` | Version for releases | `build-all`, `docker-*`, `release` |
| `IMAGE_NAME` | Docker image name | `docker-deploy` |
| `DOCKER_USER` | Docker Hub username | `docker-deploy` |
| `DOCKER_PASS` | Docker Hub token | `docker-deploy` |
| `GH_USER` | GitHub username | `docker-deploy` |
| `GH_PASS` | GitHub token | `docker-deploy` |

## Example Workflows

### Local Development
```bash
# Start developing
make deps
make test
make build
make run

# Before committing
make fmt
make lint
make test-coverage
```

### Release Process
```bash
# Full release
make release RELEASE_VERSION=1.2.3

# Or step by step
make clean
make test  
make build-all RELEASE_VERSION=1.2.3
make docker-deploy RELEASE_VERSION=1.2.3
```

### Snapshot Testing
```bash
# Show what version will be built
make show-snapshot-version

# Build snapshot
make snapshot

# Check version
make version
``` 