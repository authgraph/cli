# Authgraph CLI

Command-line interface for the [Authgraph](https://authgraph.dev) permission engine.

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap authgraph/tap
brew install authgraph
```

### Go Install

```bash
go install github.com/authgraph/cli@latest
```

### Script (macOS/Linux)

```bash
curl -fsSL https://get.authgraph.dev | sh
```

### From Source

```bash
git clone https://github.com/authgraph/cli.git
cd cli
go build -o authgraph .
```

## Quick Start

```bash
# Login with your API key
authgraph login

# Check a permission
authgraph check user:alice read document:readme

# Grant a permission
authgraph grant user:alice editor document:readme

# Revoke a permission
authgraph revoke user:alice editor document:readme

# List what a user can access
authgraph list user:alice read document

# See who has access to a resource
authgraph expand document:readme read

# Push a schema
authgraph schema push --file permissions.yaml

# Validate a schema
authgraph schema validate --file permissions.yaml

# Run permission tests
authgraph test --file permission-tests.yaml
```

## Permission Tests

Create a YAML file with permission assertions:

```yaml
tests:
  - name: "alice can read the readme"
    subject: "user:alice"
    permission: "read"
    resource: "document:readme"
    expected: allowed

  - name: "bob cannot delete the project"
    subject: "user:bob"
    permission: "delete"
    resource: "project:main"
    expected: denied
```

Run with:

```bash
authgraph test --file permission-tests.yaml
```

Use in CI/CD to prevent permission regressions:

```yaml
# .github/workflows/test.yaml
- run: authgraph test --file permission-tests.yaml
  env:
    AUTHGRAPH_API_KEY: ${{ secrets.AUTHGRAPH_API_KEY }}
```
