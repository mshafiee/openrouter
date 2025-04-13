# Contributing to openrouter-go

First off, thank you for considering contributing to `openrouter-go`! Your help is appreciated. Whether it's reporting a bug, suggesting a feature, improving documentation, or writing code, contributions make this project better.

This document provides guidelines for contributing. Please feel free to propose changes to this document in a pull request.

## Table of Contents

*   [Code of Conduct](#code-of-conduct)
*   [How Can I Contribute?](#how-can-i-contribute)
    *   [Reporting Bugs](#reporting-bugs)
    *   [Suggesting Enhancements](#suggesting-enhancements)
    *   [Pull Requests](#pull-requests)
*   [Development Setup](#development-setup)
*   [Pull Request Process](#pull-request-process)
*   [Style Guides](#style-guides)
*   [License](#license)

## Code of Conduct

This project and everyone participating in it is governed by the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Bugs are tracked as [GitHub Issues](https://github.com/mshafiee/openrouter-go/issues).

Before reporting a bug, please check the existing issues to see if someone has already reported it.

When reporting a bug, please include:

1.  **A clear and descriptive title.**
2.  **Steps to reproduce the behavior.** Provide a minimal code example if possible.
3.  **Expected behavior:** What you expected to happen.
4.  **Actual behavior:** What actually happened.
5.  **Environment:** Your Go version (`go version`), operating system.
6.  **Library Version:** The version of `openrouter-go` you are using (e.g., `v0.1.0`).
7.  **Relevant logs or error messages.** Please format logs and code snippets using Markdown code blocks.

*(Optional: Consider creating Issue Templates on GitHub for bugs.)*

### Suggesting Enhancements

Enhancement suggestions are tracked as [GitHub Issues](https://github.com/mshafiee/openrouter-go/issues) or [GitHub Discussions](https://github.com/mshafiee/openrouter-go/discussions) (if enabled).

Before suggesting an enhancement, please check if a similar suggestion already exists.

When suggesting an enhancement, please include:

1.  **A clear and descriptive title.**
2.  **Detailed description:** Explain the enhancement and why it would be valuable. What problem does it solve?
3.  **Use case:** How would this feature be used?
4.  **(Optional) Proposed implementation:** Any ideas on how it could be implemented.

*(Optional: Consider creating Issue Templates on GitHub for feature requests.)*

### Pull Requests

We actively welcome your pull requests!

*   Small bug fixes or documentation improvements can be submitted directly.
*   For larger changes (new features, significant refactoring), it's best to **open an issue first** to discuss the proposed changes with the maintainers. This helps ensure your contribution aligns with the project's goals and avoids duplicated effort.

## Development Setup

To get started with developing `openrouter-go`:

1.  **Prerequisites:**
    *   Go (latest stable version recommended)
    *   Git
    *   An OpenRouter API Key (set as `OPENROUTER_API_KEY` environment variable for running tests against the live API). *Be cautious when running tests that hit the live API, as they may incur costs.*
    *   (Optional) An OpenRouter Provisioning Key (set as `OPENROUTER_PROVISIONING_KEY` environment variable) for testing API key management features.

2.  **Clone the repository:**
    ```bash
    git clone https://github.com/mshafiee/openrouter-go.git
    cd openrouter-go
    ```

3.  **Install dependencies:**
    ```bash
    go mod download
    ```

4.  **Build the code:**
    ```bash
    go build ./...
    ```

5.  **Run tests:**
    ```bash
    # Run unit tests (usually don't require API keys)
    go test ./...

    # Run integration tests (may require API keys set as environment variables)
    # You might need specific build tags or test flags depending on test setup
    # Example: go test -tags=integration ./...
    # Ensure you understand the implications (cost, resource creation) before running integration tests.
    ```
    *   Consider adding mocks for external API calls to enable more comprehensive unit testing without hitting the actual OpenRouter service. See existing tests for patterns.

6.  **Format and Lint:**
    ```bash
    go fmt ./...
    go vet ./...
    # Consider using golangci-lint if configured
    # golangci-lint run
    ```

## Pull Request Process

1.  **Fork** the repository (`https://github.com/mshafiee/openrouter-go/fork`).
2.  **Clone** your fork (`git clone https://github.com/YOUR_USERNAME/openrouter-go.git`).
3.  Create a **new branch** for your changes (`git checkout -b feature/your-feature-name` or `fix/issue-123`). Use a descriptive branch name.
4.  Make your changes. **Commit frequently** with clear, concise commit messages.
5.  **Add tests** for your changes if applicable (bug fixes should ideally have regression tests, new features need tests).
6.  Ensure all **tests pass** (`go test ./...`).
7.  Ensure your code is **formatted** (`go fmt ./...`) and **linted** (`go vet ./...` or linters).
8.  **Push** your branch to your fork (`git push origin feature/your-feature-name`).
9.  Open a **Pull Request (PR)** from your fork's branch to the `main` branch of the original `mshafiee/openrouter-go` repository.
10. **Describe your PR** clearly:
    *   What does it do?
    *   Why is this change needed?
    *   How was it tested?
    *   Link any related issues (e.g., "Fixes #123").
11. Be prepared to **respond to review comments** and make changes if requested.
12. Once approved, a maintainer will merge your PR.

## Style Guides

### Go Code

*   Follow standard Go formatting (`gofmt`).
*   Follow standard Go linting practices (`go vet`).
*   Adhere to recommendations from [Effective Go](https://go.dev/doc/effective_go) and the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) wiki where applicable.
*   Keep code clear, concise, and readable. Add comments where necessary to explain complex logic.

### Git Commit Messages

*   Use the present tense ("Add feature" not "Added feature").
*   Use the imperative mood ("Move cursor to..." not "Moves cursor to...").
*   Limit the first line to 72 characters or less.
*   Reference issues and pull requests liberally after the first line.

## License

By contributing to `openrouter-go`, you agree that your contributions will be licensed under its [MIT License](LICENSE).
