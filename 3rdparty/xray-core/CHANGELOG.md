# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive samples directory with organized examples
- Certificate generation scripts for GOST, SM2, and standard algorithms
- Protocol configuration examples for VMess and VLESS
- Server startup scripts for different certificate types
- Detailed README.md with usage examples and troubleshooting

### Changed
- Simplified samples directory structure to single level
- Consolidated all files into main samples directory
- Removed excessive subdirectories and redundant README files
- Streamlined documentation with single comprehensive README.md

### Removed
- Complex nested directory structure
- Multiple redundant README files
- Empty subdirectories
- Over-engineered organization

## File Structure

### Simplified Organization
```
samples/
├── README.md                    # Comprehensive documentation
├── make_cert_*.sh              # Certificate generation scripts
├── vmess_*.json                # VMess protocol configurations
├── vless_*.json                # VLESS protocol configurations
└── run_*.sh                    # Server startup scripts
```

### Features
- **Certificate Generation**: Scripts for SM2, GOST2012_256, and GOST2012_512 certificates
- **Protocol Examples**: Complete VMess and VLESS configurations with various certificate types
- **Server Scripts**: Ready-to-use startup scripts for different configurations
- **Documentation**: Single comprehensive README with examples and troubleshooting

### Security Improvements
- Proper file permissions for certificate scripts
- Secure certificate generation practices
- Best practices documentation for production use
- Clear security warnings and guidelines 