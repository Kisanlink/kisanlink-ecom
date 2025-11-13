# Pre-Commit Hook Configuration

## Philosophy

Pre-commit hooks should be **fast** and **non-blocking** for developer productivity, while still ensuring code quality.

## Current Configuration

### Blocking Checks (Must Pass)
1. **Code Formatting** (`go-fmt`, `go-imports`)
   - Auto-fixes formatting issues
   - Ensures consistent code style
   - ~1-2 seconds

2. **Linting** (`golangci-lint`)
   - Only checks changed files (`--new-from-rev=HEAD`)
   - Catches common bugs and anti-patterns
   - ~3-5 seconds

3. **Build Check** (`go-build-check`)
   - Ensures code compiles
   - Catches syntax errors
   - ~5-10 seconds

4. **File Hygiene**
   - Trailing whitespace removal
   - End-of-file fixes
   - JSON/YAML validation
   - <1 second

### Non-Blocking Checks (Warnings Only)
1. **Unit Tests** (`go-unit-tests-only`)
   - Runs on manual stage only: `pre-commit run --hook-stage manual`
   - Shows warning if tests fail but doesn't block commit
   - Allows WIP commits while tests are being fixed
   - ~20-60 seconds (full test suite)

## Usage

### Normal Commit (Fast Path)
```bash
git commit -m "your message"
```
Runs: Formatting, Linting, Build check (~10-15 seconds)

### With Full Tests (Before PR)
```bash
pre-commit run --hook-stage manual --all-files
git commit -m "your message"
```
Runs: All checks including unit tests

### Skip Hooks (Emergency Only)
```bash
git commit --no-verify -m "your message"
```
⚠️ Only use for emergency hotfixes

## CI/CD Integration

The full test suite runs in CI/CD pipelines:
- **PR Pipeline**: `make ci-pr` - includes all tests + race detection
- **Push to Main**: Full integration tests
- **Pre-deployment**: Coverage checks (90% minimum)

## Rationale

### Why Tests Are Non-Blocking

1. **Speed**: Developers shouldn't wait 60+ seconds for every commit
2. **WIP Commits**: Allow incremental progress even with failing tests
3. **CI Catches Issues**: GitHub Actions runs full test suite on every PR
4. **Local Testing**: Developers run `make test-unit` manually when needed

### Why Formatting/Linting Are Blocking

1. **Fast**: <10 seconds total
2. **Auto-fix**: Most issues are automatically corrected
3. **Prevents Noise**: Avoids formatting-only commits later
4. **Code Quality**: Catches bugs early (linter)

## Customization

To add new hooks, edit `.pre-commit-config.yaml`:

```yaml
- repo: local
  hooks:
    - id: my-check
      name: My Custom Check
      entry: ./scripts/my-check.sh
      language: system
      pass_filenames: false
      types: [go]
      stages: [commit]  # or [manual] for non-blocking
```

## Troubleshooting

### Hook Running Too Slow
- Check if golangci-lint cache is working
- Consider moving slow checks to `stages: [manual]`

### Hook Failing on Unrelated Files
- Ensure `--new-from-rev=HEAD` is used for incremental checks
- Add file exclusions if needed:
  ```yaml
  exclude: ^(vendor/|.pb.go$)
  ```

### Need to Bypass Hook Temporarily
```bash
git commit --no-verify -m "message"
```
Then fix issues in next commit.

## Best Practices

1. **Run tests locally before pushing**: `make test-unit`
2. **Use manual hook before creating PR**: `pre-commit run --hook-stage manual --all-files`
3. **Keep hooks fast**: <15 seconds for normal commits
4. **Trust CI for comprehensive checks**: Let CI run the full suite
5. **Update hooks when adding new tools**: Keep `.pre-commit-config.yaml` in sync

## Maintenance

- Review hook performance monthly
- Update pre-commit tool versions quarterly
- Solicit developer feedback on hook speed/value
