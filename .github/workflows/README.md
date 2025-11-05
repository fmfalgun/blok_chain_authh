# CI/CD Workflows

📍 **Location**: `.github/workflows/`
🔗 **Parent Guide**: [Back to Main README](../../README.md)
📚 **Related**: [Developer Guide](../../DEVELOPER_GUIDE.md)

---

## 📋 Overview

This directory contains GitHub Actions workflows that automate testing, security scanning, and deployment of the blockchain authentication framework. These workflows run automatically on every push and pull request to ensure code quality and security.

## 🎯 Purpose

**Why CI/CD Pipelines?**
- ✅ **Automated Testing**: Run tests on every code change
- ✅ **Early Bug Detection**: Catch issues before they reach production
- ✅ **Security Scanning**: Identify vulnerabilities automatically
- ✅ **Consistent Builds**: Same build process for everyone
- ✅ **Deployment Automation**: Deploy to staging/production automatically

## 📁 Files in This Directory

### 1. `test.yml` - Test Pipeline

**Purpose**: Runs automated tests on every push/PR.

**What it does**:
```yaml
Triggers: Push to main/develop, Pull Requests
Jobs:
  1. Lint Code           → Checks Go code style and quality
  2. Build Chaincodes    → Compiles all 3 chaincodes
  3. Security Scan       → Runs basic Go vet checks
```

**When it runs**:
- On every push to `main`, `develop`, or feature branches
- On every pull request to `main` or `develop`

**Key Features**:
- **Parallel Execution**: Builds all 3 chaincodes simultaneously
- **Go 1.21**: Uses Go 1.21 for consistency
- **Non-Blocking**: Uses `continue-on-error` for linting (won't fail PR)
- **Timeout Protection**: 5-minute timeout for linting

**Technologies Used**:
- **GitHub Actions**: CI/CD platform
- **Go**: Programming language
- **golangci-lint**: Go linting tool
- **go vet**: Go static analysis tool

**Example Output**:
```
✅ Lint Code - Passed (with warnings)
✅ Build Chaincodes (as-chaincode) - Passed
✅ Build Chaincodes (tgs-chaincode) - Passed
✅ Build Chaincodes (isv-chaincode) - Passed
✅ Security Scan - Passed
```

**Configuration Details**:
```yaml
# Linting with timeout
- name: Run golangci-lint
  continue-on-error: true
  uses: golangci/golangci-lint-action@v3
  with:
    version: latest
    args: --timeout=5m --skip-dirs=tests
    skip-pkg-cache: true
    skip-build-cache: true

# Building each chaincode
- name: Build chaincode
  run: |
    cd chaincodes/${{ matrix.chaincode }}
    go mod tidy
    go mod download
    go build -v ./...
```

---

### 2. `security.yml` - Security Pipeline

**Purpose**: Performs security scanning and vulnerability detection.

**What it does**:
```yaml
Triggers: Push to main/develop, Pull Requests, Weekly (Sundays)
Jobs:
  1. Secret Scan         → Detects hardcoded secrets
  2. Dependency Check    → Checks for vulnerable dependencies
  3. Code Analysis       → Static code analysis
  4. License Check       → Verifies license compliance
  5. SAST (CodeQL)       → Advanced security analysis
```

**When it runs**:
- On pushes to `main` or `develop`
- On pull requests
- Weekly on Sundays (scheduled scan)

**Security Tools**:

1. **TruffleHog** - Secret Detection
   ```yaml
   Purpose: Find accidentally committed secrets (API keys, passwords)
   How: Scans entire git history
   Action: Flags suspicious patterns
   ```

2. **Go Dependency Check**
   ```yaml
   Purpose: Check for known vulnerabilities in dependencies
   How: Downloads and validates all Go modules
   Action: Reports vulnerable packages
   ```

3. **Go Vet** - Static Analysis
   ```yaml
   Purpose: Find potential bugs and suspicious code
   How: Analyzes Go code without running it
   Action: Reports suspicious constructs
   ```

4. **CodeQL** - Advanced SAST
   ```yaml
   Purpose: Deep semantic code analysis
   How: Builds code database and queries for patterns
   Action: Finds security vulnerabilities
   ```

**Why Non-Blocking?**
All security jobs use `continue-on-error: true` because:
- New dependencies may have unfixed CVEs
- Allows PRs to merge while tracking issues
- Security findings reviewed separately
- Prevents workflow paralysis

**Example Security Finding**:
```
⚠️ TruffleHog: Potential secret detected
   File: config/test.yaml
   Line: 42
   Pattern: Possible AWS Access Key

Action: Review and rotate if necessary
```

---

### 3. `deploy.yml` - Deployment Pipeline

**Purpose**: Automates deployment to staging and production environments.

**What it does**:
```yaml
Triggers: Version tags (v*.*.* format), Manual workflow dispatch
Jobs:
  1. Build Release Packages    → Compile and package chaincodes
  2. Create GitHub Release     → Create release with artifacts
  3. Deploy to Staging         → Deploy to staging environment
  4. Deploy to Production      → Deploy to production environment
```

**Deployment Triggers**:

1. **Git Tags** (Automated):
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   # Triggers: Build → Create Release → Deploy to Production
   ```

2. **Manual Trigger** (Controlled):
   ```bash
   # Go to GitHub Actions → Deploy workflow → Run workflow
   # Select: staging or production
   # Triggers: Build → Deploy to chosen environment
   ```

**Release Process**:
```
1. Developer creates git tag (v1.2.3)
2. GitHub Actions triggered automatically
3. Builds all chaincodes
4. Packages chaincodes as .tar.gz files
5. Generates SHA256 checksums
6. Creates GitHub Release with artifacts
7. Deploys to staging (if beta tag) or production
8. Verifies deployment
9. Sends notifications
```

**Deployment Environments**:

| Environment | URL | When Deployed |
|-------------|-----|---------------|
| Staging | https://staging.example.com | Beta releases (v1.0.0-beta) |
| Production | https://production.example.com | Stable releases (v1.0.0) |

**Release Artifacts**:
- `as-chaincode-v1.0.0.tar.gz` - AS chaincode package
- `tgs-chaincode-v1.0.0.tar.gz` - TGS chaincode package
- `isv-chaincode-v1.0.0.tar.gz` - ISV chaincode package
- `checksums.txt` - SHA256 checksums for verification

---

## 🛠️ Technologies Used

### GitHub Actions
**What**: CI/CD platform built into GitHub
**Why**:
- Free for public repos
- Integrated with GitHub
- Large ecosystem of actions
- Easy YAML configuration

**How it works**:
```yaml
name: Workflow Name
on: [push, pull_request]
jobs:
  job-name:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: echo "Hello World"
```

### Ubuntu Latest Runner
**What**: Virtual machine that runs workflows
**Specs**:
- 2-core CPU
- 7 GB RAM
- 14 GB SSD
- Ubuntu 22.04 LTS

**Why**: Standard, reliable, free tier available

### Go 1.21
**What**: Programming language for chaincodes
**Why**:
- Required by Hyperledger Fabric
- Fast compilation
- Strong typing
- Excellent tooling

---

## 🚀 How to Use

### Running Workflows Locally (Using Act)

```bash
# Install act (GitHub Actions local runner)
brew install act  # macOS
# or
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Run test workflow locally
act -j build-chaincodes

# Run with specific event
act push

# Run specific workflow
act -W .github/workflows/test.yml
```

### Triggering Manual Deployment

1. Go to your GitHub repository
2. Click "Actions" tab
3. Select "Deploy" workflow
4. Click "Run workflow"
5. Choose environment (staging/production)
6. Click "Run workflow" button

### Viewing Workflow Results

```bash
# Using GitHub CLI
gh run list
gh run view <run-id>
gh run watch <run-id>

# Check specific workflow
gh run list --workflow=test.yml
```

---

## 📊 Workflow Status Badges

Add to your README.md:

```markdown
![Test](https://github.com/your-org/blok_chain_authh/actions/workflows/test.yml/badge.svg)
![Security](https://github.com/your-org/blok_chain_authh/actions/workflows/security.yml/badge.svg)
![Deploy](https://github.com/your-org/blok_chain_authh/actions/workflows/deploy.yml/badge.svg)
```

---

## 🔧 Customization

### Adding New Jobs

Edit any workflow file:

```yaml
jobs:
  my-new-job:
    name: My New Job
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run custom script
        run: ./scripts/my-script.sh
```

### Changing Triggers

```yaml
# Run on specific branches only
on:
  push:
    branches: [ main, develop ]

# Run on schedule
on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly on Sundays
```

### Adding Secrets

```bash
# GitHub Settings → Secrets → New repository secret
Name: AWS_ACCESS_KEY_ID
Value: your-secret-value

# Use in workflow
env:
  AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
```

---

## 🐛 Troubleshooting

### Workflow Fails to Start
**Problem**: Workflow doesn't trigger
**Solution**: Check triggers in workflow file match your action (push, PR, etc.)

### Build Timeouts
**Problem**: Job exceeds timeout
**Solution**: Increase timeout in workflow:
```yaml
jobs:
  my-job:
    timeout-minutes: 30  # Default is 360
```

### Dependency Download Fails
**Problem**: `go mod download` fails
**Solution**:
```bash
# Locally verify dependencies
cd chaincodes/as-chaincode
go mod verify
go mod tidy
```

---

## 📚 Learn More

### GitHub Actions Documentation
- [Workflow Syntax](https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions)
- [Available Actions](https://github.com/marketplace?type=actions)
- [Self-Hosted Runners](https://docs.github.com/en/actions/hosting-your-own-runners)

### Related Documentation
- **Test Writing**: See [tests/README.md](../../tests/README.md)
- **Deployment**: See [docs/deployment/PRODUCTION_DEPLOYMENT.md](../../docs/deployment/PRODUCTION_DEPLOYMENT.md)
- **Security**: See [docs/security/](../../docs/security/)

---

## 🎯 Best Practices

1. ✅ **Always use specific action versions** (v3, not @main)
2. ✅ **Set timeouts** to prevent hanging workflows
3. ✅ **Use secrets** for sensitive data, never hardcode
4. ✅ **Cache dependencies** to speed up builds
5. ✅ **Use matrix builds** for testing multiple versions
6. ✅ **Add status badges** to README for visibility

---

## 🔄 Next Steps

After understanding CI/CD workflows:
- 📖 **Learn about chaincodes**: [chaincodes/README.md](../../chaincodes/README.md)
- 🧪 **Understand testing**: [tests/README.md](../../tests/README.md)
- 🚀 **Deploy the system**: [docs/deployment/](../../docs/deployment/)

---

📍 **Navigation**: [Main README](../../README.md) | [Developer Guide](../../DEVELOPER_GUIDE.md) | [Chaincodes →](../../chaincodes/README.md)
