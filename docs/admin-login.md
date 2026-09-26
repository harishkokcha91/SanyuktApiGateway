# Admin Panel Login Guide

## Default Admin Credentials
- **Email:** admin@sanyuktnamdev.com
- **Password:** Admin@123

> ⚠️ Change these credentials immediately in production.

## Login Steps
1. Navigate to the Admin Panel URL (e.g., http://localhost:3000/admin).
2. Enter your admin email and password.
3. On successful login, you will be redirected to the **Dashboard**.
4. Use the sidebar to access:
   - **Analytics** → View system stats
   - **Pending Items** → Approve/Reject content
   - **Notifications** → System alerts

## Token Handling
- JWT tokens are issued on login.
- Tokens expire after 1 hour.
- Refresh tokens are automatically handled by the client.

## Security Notes
- Always use HTTPS in production.
- Rotate admin passwords regularly.
- Enable 2FA for admin accounts (future feature).
