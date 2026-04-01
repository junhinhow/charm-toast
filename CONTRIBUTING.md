# Contributing to charm-toast

Thank you for your interest in contributing!

## How to Contribute

1. **Fork** the repository
2. **Clone** your fork locally
3. **Create a branch** for your feature or fix: `git checkout -b feat/my-feature`
4. **Make your changes** and add tests
5. **Run tests**: `go test ./...`
6. **Commit** using conventional commits: `feat:`, `fix:`, `refactor:`, `docs:`, `chore:`
7. **Push** your branch and open a **Pull Request**

## Development

```bash
# Run tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Check formatting
gofmt -d .
```

## Guidelines

- Write clear, descriptive commit messages
- Add tests for new functionality
- Keep code simple and well-documented
- Follow Go conventions and idioms

## Reporting Issues

Open an issue describing:
- What you expected to happen
- What actually happened
- Steps to reproduce
- Go version and OS

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
