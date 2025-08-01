# Scripts Directory

This directory contains utility scripts for development, testing, and deployment.

## Scripts

### `install-hooks.sh`
Sets up the development environment including Git hooks and required tools.

**Usage:**
```bash
./scripts/install-hooks.sh
# or
make setup-hooks
```

**What it does:**
- Installs Git pre-commit hook
- Installs required Go development tools
- Sets up pre-commit framework (if available)
- Updates .gitignore with development entries

### `pre-commit.sh`
The actual pre-commit hook that runs code quality checks before each commit.

**Usage:**
This script runs automatically when you commit, but you can also run it manually:
```bash
./scripts/pre-commit.sh
# or
make pre-commit
```

**Checks performed:**
- `go mod tidy` - Ensures dependencies are clean
- `gofmt` - Code formatting
- `goimports` - Import organization
- `go vet` - Static analysis
- `golangci-lint` - Comprehensive linting
- `go test` - Unit tests
- `gosec` - Security analysis (if installed)
- TODO/FIXME comment detection

## Installation

Run the setup script to install everything:

```bash
make setup-hooks
```

This will:
1. Install the Git pre-commit hook
2. Install required Go tools
3. Set up the development environment

## Manual Tool Installation

If you prefer to install tools manually:

```bash
# Required tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Optional but recommended
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
go install mvdan.cc/gofumpt@latest

# For pre-commit framework (optional)
pip install pre-commit
# or
brew install pre-commit
```

## Troubleshooting

### Pre-commit hook is too slow
If the pre-commit hook takes too long, you can:
- Skip tests: Set `SKIP_TESTS=1` environment variable
- Skip linting: Use `git commit --no-verify` (not recommended)

### Tools not found
Run `make install-tools` to install all required tools.

### Permission denied
Make sure scripts are executable:
```bash
chmod +x scripts/*.sh
```

## Configuration

### Customizing the pre-commit hook
Edit `scripts/pre-commit.sh` to modify which checks run.

### Customizing linting rules
Edit `.golangci.yml` to modify linting configuration.

### Pre-commit framework
If you have the pre-commit framework installed, you can use:
- `.pre-commit-config.yaml` for configuration
- `pre-commit run --all-files` to run all checks
- `pre-commit autoupdate` to update hooks 