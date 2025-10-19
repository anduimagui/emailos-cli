# NPM CLI Publishing Best Practices Guide

**Tags**: `npm`, `cli`, `packaging`, `distribution`, `best-practices`, `golang`, `binary-distribution`  
**Date**: 2025-10-19  
**Author**: Development Team  
**Status**: Active Guide

## Executive Summary

This guide outlines the best practices for publishing CLI tools to NPM, specifically addressing the challenges of distributing compiled binaries across multiple platforms. We examine three primary distribution strategies: bundled binaries, on-demand downloads, and source compilation, with recommendations for each approach based on project requirements.

## Understanding CLI Distribution on NPM

### Core Concepts

**Pre-compiled Binaries**: Platform-specific executable files built during CI/CD that are ready to run without compilation on the target system.

**Package Size Considerations**: NPM packages should balance functionality with download size. Large packages impact installation speed and storage requirements.

**Platform Detection**: Runtime identification of the user's operating system and architecture to select appropriate binaries.

## Distribution Strategies

### Strategy 1: Bundle All Platform Binaries (Recommended)

**Approach**: Include all platform-specific binaries directly in the NPM package.

**Advantages**:
- No network dependencies during installation
- Guaranteed availability across all platforms
- Immediate execution after installation
- Works in offline environments
- Simplest installation experience

**Disadvantages**:
- Larger package size (5x typical size)
- Downloads unused binaries for other platforms
- Higher bandwidth usage for initial download

**Implementation**:
```javascript
// In install.js
const platform = process.platform;
const arch = process.arch;
const binaryPath = path.join(__dirname, '..', 'bin', `mailos-${platform}-${arch}`);
```

**Package Structure**:
```
npm-package/
├── bin/
│   ├── mailos-darwin-amd64
│   ├── mailos-darwin-arm64
│   ├── mailos-linux-amd64
│   ├── mailos-linux-arm64
│   └── mailos-windows-amd64.exe
├── scripts/
│   └── install.js
└── package.json
```

### Strategy 2: On-Demand Binary Download

**Approach**: Download platform-specific binary during npm install from external source.

**Advantages**:
- Smaller initial package size
- Only downloads required binary
- Efficient bandwidth usage

**Disadvantages**:
- Requires internet connection during installation
- Dependency on external service availability
- Complex error handling for network failures
- Authentication issues with private repositories
- Potential security concerns with remote downloads

**Current Issues in Our Implementation**:
- Private GitHub repository blocks public access to release assets
- API rate limiting for unauthenticated requests
- Complex fallback mechanisms required

### Strategy 3: Source Code Compilation

**Approach**: Include Go source code and compile during installation.

**Advantages**:
- Smallest package size
- Always up-to-date with platform optimizations
- No pre-compilation required

**Disadvantages**:
- Requires Go toolchain on user system
- Significantly slower installation
- Compilation can fail on various systems
- Complex dependency management
- Poor user experience for non-developers

## Industry Best Practices

### Large CLI Tools (>50MB)

**Examples**: Docker CLI, Kubernetes kubectl, Terraform

**Pattern**: Platform-specific NPM packages
```
@mailos/cli-darwin-x64
@mailos/cli-linux-x64
@mailos/cli-win32-x64
```

**Main Package**: Meta-package that installs correct platform package
```javascript
// In main package's install script
const platformPackage = `@mailos/cli-${process.platform}-${process.arch}`;
execSync(`npm install ${platformPackage}`);
```

### Medium CLI Tools (10-50MB)

**Examples**: ESLint, Prettier, TypeScript compiler

**Pattern**: Bundle all binaries with smart extraction
- Include all platform binaries
- Extract only the required binary during postinstall
- Remove unused binaries after extraction

### Small CLI Tools (<10MB)

**Examples**: rimraf, concurrently, nodemon

**Pattern**: Bundle everything
- Include all platform binaries directly
- Simple platform detection in executable wrapper
- Minimal overhead acceptable

## Recommended Implementation for MailOS

### Phase 1: Immediate Fix (Bundle All Binaries)

**Why This Approach**:
- MailOS binaries are ~25MB total for all platforms
- Acceptable size for modern package managers
- Eliminates all network dependencies
- Provides reliable installation experience

**Implementation Steps**:

1. **Modify CI/CD Workflow**:
```yaml
- name: Prepare NPM Package
  run: |
    cd npm
    mkdir -p bin
    # Extract all platform binaries to bin/
    for platform in darwin-amd64 darwin-arm64 linux-amd64 linux-arm64 windows-amd64; do
      tar -xzf "../artifacts/mailos-${platform}.tar.gz" -C bin/
      mv bin/mailos bin/mailos-${platform}
    done
```

2. **Simplify Install Script**:
```javascript
const platformKey = `${process.platform}-${process.arch}`;
const binaryName = process.platform === 'win32' ? 'mailos.exe' : 'mailos';
const sourceBinary = path.join(__dirname, '..', 'bin', `mailos-${platformKey}`);
const targetBinary = path.join(__dirname, '..', 'bin', binaryName);

// Copy and make executable
fs.copyFileSync(sourceBinary, targetBinary);
if (process.platform !== 'win32') {
  fs.chmodSync(targetBinary, '755');
}
```

3. **Update Package.json**:
```json
{
  "bin": {
    "mailos": "./bin/mailos"
  },
  "files": [
    "bin/",
    "scripts/",
    "index.js"
  ]
}
```

### Phase 2: Future Optimization (Platform-Specific Packages)

**When to Consider**:
- Package size exceeds 100MB
- Multiple CLI tools in ecosystem
- Need for granular update control

**Structure**:
```
@mailos/core           # Shared libraries and utilities
@mailos/cli-darwin     # macOS-specific package
@mailos/cli-linux      # Linux-specific package  
@mailos/cli-windows    # Windows-specific package
mailos                 # Meta-package that installs correct platform package
```

## Security Considerations

### Binary Verification
- Include SHA256 checksums for all binaries
- Verify integrity during installation
- Sign binaries with code signing certificates

### Supply Chain Security
- Pin all dependencies to specific versions
- Use npm audit for vulnerability scanning
- Implement provenance attestation for binaries

### Access Control
- Use NPM automation tokens with minimal scope
- Implement proper secret management
- Regular rotation of publishing credentials

## Performance Optimization

### Package Size Management
- Compress binaries with UPX (Universal Packer for eXecutables)
- Remove debug symbols from release binaries
- Use Go build flags: `-ldflags="-s -w"`

### Installation Speed
- Minimize postinstall script complexity
- Avoid unnecessary file operations
- Use efficient binary copying methods

### Runtime Performance
- Place binaries in predictable locations
- Minimize startup time with proper binary structure
- Implement lazy loading for large dependencies

## Monitoring and Analytics

### Installation Metrics
- Track installation success/failure rates by platform
- Monitor download sizes and times
- Collect error reports from postinstall scripts

### Usage Analytics
- Platform distribution of users
- Version adoption rates
- Geographic distribution patterns

## Conclusion

For MailOS CLI, the recommended approach is to bundle all platform binaries directly in the NPM package. This provides the most reliable user experience while maintaining acceptable package sizes. The current on-demand download approach introduces unnecessary complexity and failure points that can be eliminated with a simpler bundled binary strategy.

The bundled binary approach aligns with industry standards for CLI tools of similar size and complexity, ensuring users can install and use MailOS reliably across all supported platforms without additional system requirements or network dependencies.