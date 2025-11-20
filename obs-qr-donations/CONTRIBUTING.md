# Contributing to OBS QR Donations

Thank you for your interest in contributing to the OBS QR Donations plugin! We welcome contributions from the community to help improve this project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Submitting a Pull Request](#submitting-a-pull-request)
- [Reporting Issues](#reporting-issues)
- [Feature Requests](#feature-requests)
- [Code Style](#code-style)
- [Testing](#testing)
- [Documentation](#documentation)
- [License](#license)

## Code of Conduct

This project adheres to the [Contributor Covenant](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
   ```bash
   git clone https://github.com/your-username/obs-qr-donations.git
   cd obs-qr-donations
   ```
3. Set up the development environment (see below)
4. Create a new branch for your changes
   ```bash
   git checkout -b feature/your-feature-name
   ```
5. Make your changes and commit them
6. Push your changes to your fork
7. Open a pull request

## Development Setup

### Prerequisites

- OBS Studio 28.0 or later
- CMake 3.16 or later
- C++17 compatible compiler
- Qt 6.x
- Python 3.8+ (for development tools)

### Building the Plugin

1. Create a build directory and configure the project:
   ```bash
   mkdir build
   cd build
   cmake ..
   ```

2. Build the plugin:
   ```bash
   cmake --build . --config Release
   ```

3. Install the plugin:
   ```bash
   cmake --install . --config Release
   ```

## Making Changes

### Code Style

- Follow the existing code style in the project
- Use meaningful variable and function names
- Add comments to explain complex logic
- Keep lines under 100 characters
- Use 4 spaces for indentation

### Git Commit Messages

- Use the present tense ("Add feature" not "Added feature")
- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally
- Consider starting the commit message with an applicable emoji:
  - ✨ `:sparkles:` when adding a new feature
  - 🐛 `:bug:` when fixing a bug
  - ♻️ `:recycle:` when refactoring code
  - 📚 `:books:` when updating documentation
  - 🚀 `:rocket:` when improving performance
  - 🧪 `:test_tube:` when adding tests

## Submitting a Pull Request

1. Ensure your fork is up to date with the main repository
2. Create a new branch for your changes
3. Make your changes and commit them with descriptive messages
4. Push your changes to your fork
5. Open a pull request against the main branch
6. Fill out the pull request template with all relevant information
7. Ensure all tests pass and the code is properly documented

## Reporting Issues

When reporting issues, please include:

- A clear and descriptive title
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Screenshots if applicable
- OBS Studio version and operating system
- Any error messages

## Feature Requests

We welcome feature requests! Please:

1. Check if the feature has already been requested
2. Explain why this feature would be useful
3. Provide as much detail as possible
4. Include any relevant links or references

## Testing

Please ensure your changes are properly tested. We use:

- Unit tests for core functionality
- Integration tests for OBS Studio interaction
- Manual testing for UI changes

To run the tests:

```bash
cd build
ctest -V
```

## Documentation

Good documentation is crucial. Please ensure:

- All new features are documented
- Existing documentation is updated if your changes affect it
- Code is properly commented
- README.md is updated if necessary

## License

By contributing to this project, you agree that your contributions will be licensed under the project's [MIT License](LICENSE).

## Thank You!

Your contributions help make this project better for everyone. Thank you for taking the time to contribute!
