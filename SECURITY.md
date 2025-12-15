# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of Go API Starter seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### How to Report

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to: **<security@example.com>** (replace with actual email)

You should receive a response within 48 hours. If for some reason you do not, please follow up via email to ensure we received your original message.

Please include the following information in your report:

- Type of vulnerability (e.g., SQL injection, XSS, authentication bypass)
- Full paths of source file(s) related to the vulnerability
- Location of the affected source code (tag/branch/commit or direct URL)
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit it

### Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 5 business days
- **Resolution Target**: Within 90 days (critical vulnerabilities may be expedited)

## Security Best Practices

This project implements the following security measures:

### Authentication & Authorization

- **JWT Authentication**: Tokens are signed using HMAC-SHA256
- **Token Expiration**: Configurable token lifetime (default: 24 hours)
- **Password Hashing**: bcrypt with configurable cost factor
- **Rate Limiting**: Prevents brute force attacks

### Input Validation

- **Request Validation**: All inputs are validated using go-playground/validator
- **SQL Injection Prevention**: Parameterized queries via database/sql
- **XSS Prevention**: Response encoding and Content-Type headers

### Configuration Security

- **Environment Variables**: Sensitive data stored in environment variables
- **No Hardcoded Secrets**: All secrets are externalized
- **Configuration Validation**: Startup validation of required configuration

### HTTP Security Headers

The API includes security middleware that sets:

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Content-Security-Policy: default-src 'self'
Strict-Transport-Security: max-age=31536000; includeSubDomains (HTTPS only)
```

### Logging & Monitoring

- **Request Logging**: All requests are logged with correlation IDs
- **Audit Trail**: Authentication events are logged
- **Error Handling**: Errors are logged but not exposed to clients
- **Metrics**: Prometheus metrics for security monitoring

## Secure Development Guidelines

### For Contributors

1. **Never commit secrets**: Use environment variables
2. **Validate all inputs**: Never trust user input
3. **Use parameterized queries**: Prevent SQL injection
4. **Handle errors securely**: Don't expose internal details
5. **Keep dependencies updated**: Run `go get -u` and audit regularly
6. **Run security linters**: golangci-lint includes security checks

### Dependency Security

We use the following tools to monitor dependencies:

- **go mod tidy**: Keep dependencies clean
- **govulncheck**: Go vulnerability scanner
- **dependabot**: Automated dependency updates

To check for vulnerabilities:

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

### Code Review Checklist

- [ ] No hardcoded credentials or secrets
- [ ] All inputs validated and sanitized
- [ ] Errors handled without exposing internals
- [ ] Authentication/authorization properly implemented
- [ ] SQL queries use parameterized statements
- [ ] Sensitive data not logged
- [ ] Rate limiting applied where appropriate

## Known Security Considerations

### Development vs Production

- **Development**: Uses local environment variables, may have relaxed CORS
- **Production**: Should use secure secrets management, strict CORS

### API Keys and Tokens

- JWT secret must be at least 32 characters
- Rotate secrets periodically
- Use different secrets per environment

### Database Security

- Use TLS for database connections in production
- Implement connection pooling limits
- Use least-privilege database users

## Vulnerability Disclosure Timeline

1. **Day 0**: Vulnerability reported
2. **Day 1-2**: Initial assessment and acknowledgment
3. **Day 3-14**: Investigation and fix development
4. **Day 15-30**: Testing and validation
5. **Day 31-90**: Release and disclosure coordination

## Contact

For security-related questions that don't involve a vulnerability report, please open a GitHub issue with the `security` label.

Thank you for helping keep Go API Starter and its users safe!
