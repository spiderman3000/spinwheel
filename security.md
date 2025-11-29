## Security Considerations

### For Development
- Never commit `.env` files or secrets
- Use `.env.local` for local development
- The `X-User-Id` header is for development only

### For Production
- Replace `X-User-Id` with proper JWT authentication
- Enable HTTPS/TLS for all connections
- Use environment variables for all sensitive configuration
- Enable rate limiting (100 req/min per user)
- Rotate any exposed credentials immediately
- Review CORS settings for production domains - not implemented

### Reporting Security Issues
Please report security vulnerabilities to [your-email] or use GitHub Security Advisories.