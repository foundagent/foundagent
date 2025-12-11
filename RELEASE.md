# Release Process

This document describes the release process for Foundagent.

## Versioning

Foundagent follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for new functionality in a backwards compatible manner
- **PATCH** version for backwards compatible bug fixes

## Release Checklist

### 1. Prepare the Release

- [ ] Ensure all features for the release are merged to `main`
- [ ] Run full test suite: `make test`
- [ ] Run linters: `make lint`
- [ ] Update version in documentation if needed
- [ ] Review and update CHANGELOG.md (if maintained)

### 2. Create Release Tag

```bash
# Fetch latest changes
git checkout main
git pull origin main

# Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### 3. Automated Release Process

Once the tag is pushed, GitHub Actions will automatically:

1. **Run Tests**: Execute full test suite on all platforms
2. **Build Binaries**: Create binaries for:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64, arm64)
3. **Generate Checksums**: Create SHA256 checksums for all binaries
4. **Create Release**: Publish GitHub release with:
   - Release notes (auto-generated from commits)
   - Installation instructions
   - All binaries and checksums
5. **Build Docker Image**: Build and push multi-arch Docker image to GHCR

### 4. Verify Release

- [ ] Check [GitHub Releases](https://github.com/foundagent/foundagent/releases) page
- [ ] Download and test binaries for different platforms
- [ ] Verify Docker image: `docker pull ghcr.io/foundagent/foundagent:v1.0.0`
- [ ] Test shell completions work correctly

### 5. Announce Release

- [ ] Update documentation site (if applicable)
- [ ] Announce on relevant channels
- [ ] Update Homebrew formula (if applicable)

## Local Build (if needed)

If you need to build binaries locally for testing:

```bash
# Build all platform binaries
make release

# This creates binaries in dist/ directory
ls -l dist/
```

## Hotfix Release

For critical bug fixes:

1. Create a hotfix branch from the tag:
   ```bash
   git checkout -b hotfix/v1.0.1 v1.0.0
   ```

2. Apply fixes and commit

3. Create new tag:
   ```bash
   git tag -a v1.0.1 -m "Hotfix v1.0.1"
   git push origin v1.0.1
   ```

4. Merge hotfix back to main:
   ```bash
   git checkout main
   git merge hotfix/v1.0.1
   git push origin main
   ```

## Rollback

If a release has critical issues:

1. **Delete the tag**:
   ```bash
   git tag -d v1.0.0
   git push origin :refs/tags/v1.0.0
   ```

2. **Delete the GitHub release** via web interface

3. **Fix the issue** and create a new release

## Pre-releases

For beta or release candidate versions, use tags with `-alpha`, `-beta`, or `-rc` suffixes:

```bash
# Examples of pre-release tags
git tag -a v1.0.0-beta -m "Beta release"
git tag -a v1.0.0-rc.1 -m "Release Candidate 1"
git tag -a v1.0.0-alpha.1 -m "Alpha release 1"

# Push the tag
git push origin v1.0.0-beta

# GitHub Actions will automatically detect and mark it as a pre-release
```

## Release Schedule

- **Patch releases**: As needed for bug fixes
- **Minor releases**: Monthly or when significant features are complete
- **Major releases**: When breaking changes are introduced

## Troubleshooting

### Build fails in CI

1. Check GitHub Actions logs
2. Reproduce locally: `make release`
3. Fix issues and push new commit
4. Delete and recreate tag if necessary

### Binary doesn't work on target platform

1. Verify GOOS/GOARCH combinations in workflow
2. Test cross-compilation locally
3. Check for platform-specific dependencies

### Docker build fails

1. Test Docker build locally: `make docker-build`
2. Check Dockerfile for issues
3. Verify base image versions

## Contact

For release-related questions, contact the maintainers or open an issue.
