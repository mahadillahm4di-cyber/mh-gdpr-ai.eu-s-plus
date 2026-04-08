# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in mh-gdpr-ai.eu S+, please report it responsibly.

**DO NOT open a public GitHub issue for security vulnerabilities.**

Instead, email: **security@mh-gdpr-ai.eu**

We will respond within 48 hours and work with you to resolve the issue.

## Security Principles

This project follows these security principles:

1. **Zero Trust** — All inputs are validated, all outputs are sanitized
2. **Encryption at Rest** — All memory data is encrypted with AES-256
3. **Encryption in Transit** — All API communication uses TLS/HTTPS
4. **Least Privilege** — Each component has minimal permissions
5. **No Secrets in Code** — All credentials via environment variables
6. **Audit Trail** — All actions are logged (without sensitive data)
7. **CORS Restriction** — Explicit origins only, never wildcard
8. **Rate Limiting** — All public endpoints are rate-limited
9. **Security Headers** — CSP, X-Frame-Options, HSTS on all responses
10. **Dependency Scanning** — Automated vulnerability checks in CI

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest  | Yes       |
