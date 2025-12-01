# Contributing to ClamAV REST

Thank you for your interest in contributing to ClamAV REST! This document provides guidelines and instructions for contributing.

## 🔄 Development Workflow

We use a semantic versioning workflow with automated releases based on your commit messages. To keep our project maintainable and high quality, please follow these best practices:

### 1. Fork and Clone

Fork the repository and work on a feature branch. Make sure your branch is up-to-date with the latest changes.

```bash
git clone https://github.com/YOUR_USERNAME/clamav-rest.git
cd clamav-rest
git remote add upstream https://github.com/ajilach/clamav-rest.git
```

### 2. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### 3. Make Your Changes

Write your code, following the existing code style and conventions. Adhere to standard Go conventions and ensure your code is clean, well-documented, and tested.

### 4. Commit Using Conventional Commits

We use [Conventional Commits](https://www.conventionalcommits.org/) to automatically determine version numbers and generate changelogs. Write clear and concise commit messages explaining your changes.

**Commit Message Format:**

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**

- `feat:` - A new feature (triggers a **minor** version bump, e.g., 1.0.0 → 1.1.0)
- `fix:` - A bug fix (triggers a **patch** version bump, e.g., 1.0.0 → 1.0.1)
- `docs:` - Documentation changes only (no version bump)
- `style:` - Code style changes (formatting, semicolons, etc.) (no version bump)
- `refactor:` - Code changes that neither fix bugs nor add features (no version bump)
- `perf:` - Performance improvements (triggers a **patch** version bump)
- `test:` - Adding or updating tests (no version bump)
- `chore:` - Maintenance tasks, dependencies updates (no version bump)
- `ci:` - CI/CD pipeline changes (no version bump)

**Breaking Changes:**

For breaking changes that require a **major** version bump (e.g., 1.0.0 → 2.0.0), add `BREAKING CHANGE:` in the commit footer:

```
feat: change API response format

BREAKING CHANGE: The scan endpoint now returns a different JSON structure.
Old format: { "status": "OK" }
New format: { "result": { "status": "OK" } }
```

**Examples:**

```bash
# New feature (minor bump)
git commit -m "feat: add support for scanning ZIP archives"

# Bug fix (patch bump)
git commit -m "fix: resolve memory leak in scan handler"

# Bug fix with more details (patch bump)
git commit -m "fix: prevent race condition in concurrent scans

Added mutex to protect shared scan state when multiple
requests are processed simultaneously."

# Breaking change (major bump)
git commit -m "feat: redesign REST API endpoints

BREAKING CHANGE: All endpoints now use /api/v3 prefix instead of /api/v2"

# Documentation update (no bump)
git commit -m "docs: update installation instructions"

# Dependency update (no bump)
git commit -m "chore: update clamd dependency to v2.1.0"
```

### 5. Push and Create a Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub targeting the `master` branch with a clear description of your changes and reference any related issues. Our maintainers will review and provide feedback.

**Before submitting your PR:**

- Ensure your code follows our coding standards
- Run tests locally to verify everything works
- Update relevant documentation and tests as needed with your changes
- Check that your branch is up-to-date with master

## 🚀 Release Process

The release process is fully automated:

1. **Pull Request** - When you open a PR:

   - CI workflow runs automatically
   - Builds Docker images for all platforms
   - Runs tests
   - Validates the build

2. **Merge to Master** - When your PR is merged:
   - Release workflow analyzes commit messages
   - Calculates the next semantic version
   - Builds and pushes multi-platform Docker images
   - Creates a GitHub release with changelog
   - Tags images with semantic versions

**Version Tags:**

After release, images are available as:

- `ajilaag/clamav-rest:1.2.3` - Specific version
- `ajilaag/clamav-rest:1.2` - Minor version (auto-updated)
- `ajilaag/clamav-rest:1` - Major version (auto-updated)
- `ajilaag/clamav-rest:latest` - Latest stable release

## 🧪 Testing

Before submitting your PR, please test your changes:

```bash
# Run the test suite
./run-tests

# Build and test locally
docker build -t clamav-rest:test .
docker run --rm clamav-rest:test
```

## 📝 Code Style

- Follow Go best practices and conventions
- Use `gofmt` to format your code
- Write clear, descriptive variable and function names
- Add comments for complex logic
- Keep functions focused and concise

## 🐛 Reporting Issues

If you encounter a bug or have a feature suggestion, please open an issue before starting work to discuss your idea.

When reporting issues, please include:

- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, Docker version, etc.)
- Relevant logs or error messages

## 📄 License

By contributing, you agree that your contributions will be licensed under the same license as the project (see LICENSE.md).

## 💬 Questions?

If you have questions about contributing, feel free to open an issue for discussion.

Thank you for contributing! 🎉
