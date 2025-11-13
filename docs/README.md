# Documentation

This directory contains comprehensive documentation for the KisanLink E-commerce API project.

## 📚 Available Documentation

### [DEVELOPMENT.md](DEVELOPMENT.md)

Complete guide for developers including:

- Development tools setup
- Code quality standards
- Pre-commit hooks configuration
- Linting and security scanning
- Testing strategies
- CI/CD integration

## 🚀 Quick Start for Developers

1. **Setup your development environment:**

   ```bash
   make setup-hooks
   ```

2. **Run quality checks:**

   ```bash
   make quick-check
   ```

3. **Make changes and commit:**
   ```bash
   # Make your changes
   make check        # Run all quality checks
   git add .
   git commit -m "your changes"  # Pre-commit hook runs automatically
   ```

## 📖 Additional Documentation

- [API Documentation](../README.md) - Main project README
- [Scripts Documentation](../scripts/README.md) - Development scripts
- [GitHub Templates](../.github/) - Issue and PR templates

## 🆘 Need Help?

- Check the troubleshooting section in [DEVELOPMENT.md](DEVELOPMENT.md)
- Create an issue using the templates in `.github/ISSUE_TEMPLATE/`
- Review the project's [main README](../README.md)
