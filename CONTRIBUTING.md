# Contributing to ZLADXHD Installer

Thank you for your interest in contributing to ZLADXHD Installer! This document provides guidelines and information for contributors.

## Code of Conduct

Please be respectful and considerate in all interactions. We welcome contributors of all backgrounds and experience levels.

## Getting Started

### Prerequisites

- Go 1.25 or later
- Git
- Linux (for full testing, as this is a Linux-specific tool)
- Steam installed (for integration testing)

### Setting Up the Development Environment

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/zladxhd-installer.git
   cd zladxhd-installer
   ```
3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/jslay88/zladxhd-installer.git
   ```
4. Install dependencies:
   ```bash
   go mod download
   ```

### Running Tests

```bash
# Run all tests
make test

# Run tests with verbose output
make test-verbose

# Run tests with coverage
make cover

# Generate HTML coverage report
make cover-html
```

### Building

```bash
# Build the binary
make build

# Build with debug symbols
make build-debug

# Install to $GOPATH/bin
make install
```

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in [Issues](https://github.com/jslay88/zladxhd-installer/issues)
2. If not, create a new issue with:
   - A clear, descriptive title
   - Steps to reproduce the bug
   - Expected behavior
   - Actual behavior
   - Go version and Linux distribution
   - Steam and Proton versions
   - Any relevant log output

### Suggesting Features

1. Check if the feature has already been suggested in [Issues](https://github.com/jslay88/zladxhd-installer/issues)
2. If not, create a new issue with:
   - A clear description of the feature
   - Use cases and benefits
   - Any implementation ideas you have

### Submitting Changes

1. Create a new branch from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following the code style guidelines

3. Add or update tests as needed

4. Ensure all tests pass:
   ```bash
   make test
   ```

5. Ensure code is properly formatted:
   ```bash
   make fmt
   ```

6. Run the linter:
   ```bash
   make lint
   ```

7. Commit your changes with a clear message:
   ```bash
   git commit -m "Add feature: description of your change"
   ```

8. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

9. Create a Pull Request on GitHub

## Project Structure

```
zladxhd-installer/
├── cmd/
│   └── zladxhd-installer/
│       └── main.go           # CLI entry point
├── internal/
│   ├── archive/              # Archive download, extraction, verification
│   ├── backup/               # Steam backup functionality
│   ├── cli/                  # Cobra CLI commands
│   ├── patcher/              # HD patcher execution
│   ├── proton/               # Proton detection and configuration
│   ├── protontricks/         # protontricks integration
│   ├── state/                # Installation state management
│   └── steam/                # Steam path detection, shortcuts, users
├── build/                    # Build output directory
├── Makefile                  # Build automation
└── README.md
```

## Code Style Guidelines

### General

- Follow standard Go conventions and idioms
- Use `go fmt` to format code
- Use `go vet` to catch common issues
- Keep functions focused and reasonably sized

### Naming

- Use clear, descriptive names
- Follow Go naming conventions (CamelCase for exported, camelCase for unexported)
- Avoid abbreviations unless widely understood

### Documentation

- Add godoc comments for all exported types, functions, and constants
- Include examples in documentation where helpful
- Keep comments up-to-date with code changes

### Testing

- Write tests for all new functionality
- Use Ginkgo/Gomega for BDD-style tests
- Test edge cases and error conditions
- Maintain or improve code coverage

### Error Handling

- Return errors rather than panicking
- Use sentinel errors for common conditions
- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Provide helpful error messages for user-facing errors

## Pull Request Guidelines

### Before Submitting

- [ ] All tests pass (`make test`)
- [ ] Code is formatted with `go fmt` (`make fmt`)
- [ ] No issues from `go vet` (`make vet`)
- [ ] Linter passes (`make lint`)
- [ ] New code has appropriate test coverage
- [ ] Documentation is updated if needed
- [ ] Commit messages are clear and descriptive

### PR Description

Include in your PR description:
- What changes are being made
- Why these changes are needed
- Any breaking changes
- Related issues (use "Fixes #123" to auto-close)

### Review Process

1. Maintainers will review your PR
2. Address any feedback or requested changes
3. Once approved, your PR will be merged

## Types of Contributions

### Good First Issues

Look for issues labeled `good first issue` - these are great for newcomers!

### Documentation

- Fix typos or unclear wording
- Add examples
- Improve README or other docs

### Code

- Bug fixes
- New features
- Performance improvements
- Refactoring

### Testing

- Add missing tests
- Improve test coverage
- Add integration tests

## Testing Notes

This project involves system integration with:
- Steam and its file structures
- Proton/Wine prefixes
- protontricks
- File system operations

When testing:
- Use mock/stub approaches for external system calls where possible
- Be careful with tests that modify real Steam installations
- Consider using a test Steam installation if doing integration testing

## Questions?

If you have questions about contributing, feel free to:
- Open an issue with your question
- Start a discussion on GitHub

Thank you for contributing! 🎉
