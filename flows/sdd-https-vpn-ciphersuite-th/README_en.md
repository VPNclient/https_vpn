# Thai Post-Quantum Ciphersuite (TH-PQC) for HTTPS VPN

> **English Version** | [ภาษาไทย](README.md)

This project implements a Post-Quantum Cryptography (PQC) stack for the HTTPS VPN architecture in Thailand. It prepares the system for future quantum threats by integrating NIST FIPS 203, 204, and 205 standards using a hybrid approach.

## Key Features

### 1. Hybrid Key Encapsulation (Hybrid KEM)
Combines classical ECC with PQC for secure key exchange:
- **Balanced Profile (Default):** X25519 + ML-KEM-768 for general traffic.
- **High-Assurance Profile:** P-384 + ML-KEM-1024 for administrative channels and enrollment.

### 2. Digital Signatures
- **Operational Use:** Hybrid Ed25519/ECDSA + ML-DSA-65 for general certificates and control API signing.
- **Conservative Trust Anchor:** SLH-DSA (FIPS 205) for offline Root Manifests and firmware signing fallback.

### 3. Security & Resilience
- **Backup KEM:** HQC (Hamming Quasi-Cyclic) included as a fallback if lattice-based schemes are compromised.
- **Hybrid Logic:** Uses HKDF-based key combination for maximum cryptographic strength.

## SDD Documentation

- [01-Requirements](01-requirements.md)
- [02-Specifications](02-specifications.md)
- [03-Plan](03-plan.md)
- [04-Implementation Log](04-implementation-log.md)

## Project Status

✅ **COMPLETED** - Core infrastructure implemented and verified with unit tests.

---
Developed by Gemini CLI for the HTTPS VPN project.
