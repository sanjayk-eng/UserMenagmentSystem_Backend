# Password Update Email Test

## Steps to Test:

1. **Start the server:**
   ```bash
   go run main.go
   ```

2. **Login as ADMIN/HR/SUPERADMIN:**
   ```bash
   curl -X POST http://localhost:8082/api/login \
     -H "Content-Type: application/json" \
     -d '{
       "email": "admin@zenithive.com",
       "password": "your_password"
     }'
   ```

3. **Update an employee password:**
   ```bash
   curl -X PATCH http://localhost:8082/api/employee/{employee_id}/password \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -d '{
       "new_password": "NewPass123"
     }'
   ```

4. **Check the console logs for:**
   - "Sending password update email to: [email]"
   - "Attempting to send email to: [email] with subject: Your Password Has Been Updated"
   - "Email sent successfully to: [email]"
   - "Password update email sent successfully to: [email]"

## Common Issues:

### Email Not Sending:
1. **Check GOOGLE_SCRIPT_URL** is set in .env
2. **Check console logs** for error messages
3. **Verify employee email** exists in database
4. **Check Google Apps Script** is deployed and accessible

### Debugging:
- Look for error messages in console starting with "Failed to..."
- Verify the goroutine is executing (check for log messages)
- Test the Google Apps Script URL directly with curl

## Expected Email Content:

```
Subject: Your Password Has Been Updated

Dear [Employee Name],

Your account password has been updated by [Admin Name] ([Role]).

Your new login credentials are:
Email: [employee@zenithive.com]
Password: [NewPass123]

If you did not request this change, please contact your HR department immediately.

For security reasons, we recommend:
1. Login with your new password
2. Change your password to something memorable
3. Keep your password secure and do not share it with anyone

Login URL: [https://zenithiveapp.netlify.app]

Best regards,
Zenithive HR Team
```
