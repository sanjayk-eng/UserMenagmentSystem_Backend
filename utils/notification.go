package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// SendEmail sends an email using Google Apps Script
func SendEmail(to, subject, body string) error {
	// Get GOOGLE_SCRIPT_URL from environment at runtime
	googleScriptURL := os.Getenv("GOOGLE_SCRIPT_URL")

	// Check if URL is set
	if googleScriptURL == "" {
		return fmt.Errorf("GOOGLE_SCRIPT_URL environment variable is not set")
	}

	fmt.Printf("Attempting to send email to: %s with subject: %s\n", to, subject)

	emailReq := EmailRequest{
		To:      to,
		Subject: subject,
		Body:    body,
	}

	jsonData, err := json.Marshal(emailReq)
	if err != nil {
		return fmt.Errorf("failed to marshal email request: %v", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second, // Increased timeout for Google Apps Script
	}

	resp, err := client.Post(googleScriptURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("email service returned status: %d", resp.StatusCode)
	}

	fmt.Printf("Email sent successfully to: %s\n", to)
	return nil
}

// SendEmployeeCreationEmail sends notification to newly created employee
func SendEmployeeCreationEmail(employeeEmail, employeeName, password string) error {
	subject := "Welcome to Zenithive - Your Account Has Been Created"
	body := fmt.Sprintf(`
Dear %s,

Welcome to Zenithive!

Your employee account has been successfully created. Below are your login credentials:

Email: %s
Password: %s

Please login to the system and change your password at your earliest convenience.

Login URL: [https://zenithiveapp.netlify.app]

If you have any questions, please contact your HR department.

Best regards,
Zenithive HR Team
`, employeeName, employeeEmail, password)

	return SendEmail(employeeEmail, subject, body)
}

// SendLeaveApplicationEmail sends notification to manager, admin, and superadmin
func SendLeaveApplicationEmail(recipients []string, employeeName, leaveType, startDate, endDate string, days float64, reason string) error {
	subject := fmt.Sprintf("Leave Application - %s", employeeName)
	body := fmt.Sprintf(`
Dear Manager/Admin,

A new leave application has been submitted and requires your review.

Employee: %s
Leave Type: %s
Start Date: %s
End Date: %s
Duration: %.1f days
Reason: %s
Status: Pending Approval

Please login to the system to approve or reject this leave request.

Best regards,
Zenithive Leave Management System
`, employeeName, leaveType, startDate, endDate, days, reason)

	for _, recipient := range recipients {
		if err := SendEmail(recipient, subject, body); err != nil {
			// Log error but continue sending to other recipients
			fmt.Printf("Failed to send email to %s: %v\n", recipient, err)
		}
	}

	return nil
}

func SendLeaveManagerRejectionEmail(
	AdminEmail []string,
	empEmail string,
	employeeName, leaveType, startDate, endDate string,
	days float64, rejectedBy string,
) error {

	subject := "Leave Request - Manager Rejection (Pending Final Decision)"

	// ------------------------
	// EMPLOYEE EMAIL (Step 1)
	// ------------------------
	empBody := fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; line-height: 1.6; font-size: 15px;">
<p><strong>Dear %s,</strong></p>

<p>Your leave application has been <strong style="color:#d9534f;">REJECTED</strong> by your manager <strong>%s</strong>.</p>

<p>This is the first-level rejection.  
The request is now forwarded to <strong>Admin/SuperAdmin</strong> for final review.</p>

<p>
<strong>Leave Type:</strong> %s<br>
<strong>Start Date:</strong> %s<br>
<strong>End Date:</strong> %s<br>
<strong>Duration:</strong> %.1f days<br>
<strong>Status:</strong> <span style="color:#d9534f;">MANAGER_REJECTED</span>
</p>

<p>For more information, please contact your manager.</p>

<p>Best regards,<br>
<strong>Zenithive Leave Management System</strong></p>
</div>
`, employeeName, rejectedBy, leaveType, startDate, endDate, days)

	if err := SendEmail(empEmail, subject, empBody); err != nil {
		return err
	}

	// ------------------------
	// ADMIN EMAIL (Step 1)
	// ------------------------
	adminBody := fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; line-height: 1.6; font-size: 15px;">
<p><strong>Dear Admin,</strong></p>

<p>A leave request has been <strong style="color:#d9534f;">REJECTED</strong> at manager level by <strong>%s</strong>.</p>

<p>This leave now requires <strong>final rejection approval</strong> from Admin/SuperAdmin.</p>

<p>
<strong>Employee:</strong> %s<br>
<strong>Leave Type:</strong> %s<br>
<strong>Start Date:</strong> %s<br>
<strong>End Date:</strong> %s<br>
<strong>Duration:</strong> %.1f days<br>
<strong>Status:</strong> <span style="color:#d9534f;">MANAGER_REJECTED</span>
</p>

<p>Please log in to the admin panel to complete the final review.</p>

<p>Best regards,<br>
<strong>Zenithive Leave Management System</strong></p>
</div>
`, rejectedBy, employeeName, leaveType, startDate, endDate, days)

	for _, email := range AdminEmail {
		if err := SendEmail(email, subject, adminBody); err != nil {
			return err
		}
	}

	return nil
}

// SendLeaveManagerApprovalEmail sends notification for manager-level approval (first step)
func SendLeaveManagerApprovalEmail(
	AdminEmail []string,
	employeeEmail, employeeName, leaveType, startDate, endDate string,
	days float64, approvedBy string,
) error {

	subject := "Leave Approved by Manager"

	// ------------------------
	// 1) EMPLOYEE EMAIL
	// ------------------------
	empBody := fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; line-height: 1.6; font-size: 15px;">
<p><strong>Dear %s,</strong></p>

<p>Your leave application has been <strong style="color:#5cb85c;">APPROVED</strong> by your manager <strong>%s</strong>.</p>

<p>
<strong>Leave Type:</strong> %s<br>
<strong>Start Date:</strong> %s<br>
<strong>End Date:</strong> %s<br>
<strong>Duration:</strong> %.1f days<br>
<strong>Status:</strong> <span style="color:#5cb85c;">MANAGER APPROVED</span>
</p>

<p>Note: Your leave is pending final approval from ADMIN/SUPERADMIN.</p>

<p>Best regards,<br>
<strong>Zenithive Leave Management System</strong></p>
</div>
`, employeeName, approvedBy, leaveType, startDate, endDate, days)

	if err := SendEmail(employeeEmail, subject, empBody); err != nil {
		return err
	}

	// ------------------------
	// 2) ADMIN EMAIL TEMPLATE
	// ------------------------
	adminBody := fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; line-height: 1.6; font-size: 15px;">
<p><strong>Dear Admin,</strong></p>

<p>A leave request has been <strong style="color:#5cb85c;">APPROVED</strong> by manager <strong>%s</strong>.</p>

<p>
<strong>Employee:</strong> %s<br>
<strong>Leave Type:</strong> %s<br>
<strong>Start Date:</strong> %s<br>
<strong>End Date:</strong> %s<br>
<strong>Duration:</strong> %.1f days<br>
<strong>Status:</strong> <span style="color:#5cb85c;">MANAGER APPROVED</span>
</p>

<p>Please review and take final action.</p>

<p>Best regards,<br>
<strong>Zenithive Leave Management System</strong></p>
</div>
`, approvedBy, employeeName, leaveType, startDate, endDate, days)

	for _, email := range AdminEmail {
		if err := SendEmail(email, subject, adminBody); err != nil {
			return err
		}
	}

	return nil
}

// SendLeaveApprovalEmail sends notification to employee when leave is approved
func SendLeaveFinalApprovalEmail(
	AdminEmail []string,
	employeeEmail, employeeName, leaveType, startDate, endDate string,
	days float64, approvedBy string,
) error {

	subject := "Leave Approved"

	// ------------------------
	// 1) EMPLOYEE EMAIL
	// ------------------------
	empBody := fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; line-height: 1.6; font-size: 15px;">
<p><strong>Dear %s,</strong></p>

<p>Your leave application has been <strong style="color:#5cb85c;">APPROVED</strong> by <strong>%s</strong>.</p>

<p>
<strong>Leave Type:</strong> %s<br>
<strong>Start Date:</strong> %s<br>
<strong>End Date:</strong> %s<br>
<strong>Duration:</strong> %.1f days<br>
<strong>Status:</strong> <span style="color:#5cb85c;">APPROVED</span>
</p>

<p>Enjoy your time off!</p>

<p>Best regards,<br>
<strong>Zenithive Leave Management System</strong></p>
</div>
`, employeeName, approvedBy, leaveType, startDate, endDate, days)

	if err := SendEmail(employeeEmail, subject, empBody); err != nil {
		return err
	}

	// ------------------------
	// 2) ADMIN EMAIL TEMPLATE
	// ------------------------
	adminBody := fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; line-height: 1.6; font-size: 15px;">
<p><strong>Dear Admin,</strong></p>

<p>The leave request of employee <strong>%s</strong> has been <strong style="color:#5cb85c;">APPROVED</strong> by <strong>%s</strong>.</p>

<p>
<strong>Leave Type:</strong> %s<br>
<strong>Start Date:</strong> %s<br>
<strong>End Date:</strong> %s<br>
<strong>Duration:</strong> %.1f days<br>
<strong>Status:</strong> <span style="color:#5cb85c;">APPROVED</span>
</p>

<p>Best regards,<br>
<strong>Zenithive Leave Management System</strong></p>
</div>
`, employeeName, approvedBy, leaveType, startDate, endDate, days)

	for _, email := range AdminEmail {
		if err := SendEmail(email, subject, adminBody); err != nil {
			return err
		}
	}

	return nil
}

// SendLeaveRejectionEmail sends notification to employee when leave is rejected
func SendLeaveRejectionEmail(
	AdminEmail []string,
	empEmail string,
	employeeName, leaveType, startDate, endDate string,
	days float64, rejectedBy string,
) error {

	subject := "Leave Request Rejected"

	// ------------------------
	// 1) EMPLOYEE EMAIL
	// ------------------------
	empBody := fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; line-height: 1.6; font-size: 15px;">
<p><strong>Dear %s,</strong></p>

<p>We regret to inform you that your leave application has been <strong style="color:#d9534f;">REJECTED</strong> by <strong>%s</strong>.</p>

<p>
<strong>Leave Type:</strong> %s<br>
<strong>Start Date:</strong> %s<br>
<strong>End Date:</strong> %s<br>
<strong>Duration:</strong> %.1f days<br>
<strong>Status:</strong> <span style="color:#d9534f;">REJECTED</span>
</p>

<p>Please contact your manager for more information.</p>

<p>Best regards,<br>
<strong>Zenithive Leave Management System</strong></p>
</div>
`, employeeName, rejectedBy, leaveType, startDate, endDate, days)

	if err := SendEmail(empEmail, subject, empBody); err != nil {
		return err
	}

	// ------------------------
	// 2) ADMIN EMAIL TEMPLATE
	// ------------------------
	adminBody := fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; line-height: 1.6; font-size: 15px;">
<p><strong>Dear Admin,</strong></p>

<p>A leave request has been <strong style="color:#d9534f;">REJECTED</strong> by <strong>%s</strong>.</p>

<p>
<strong>Employee:</strong> %s<br>
<strong>Leave Type:</strong> %s<br>
<strong>Start Date:</strong> %s<br>
<strong>End Date:</strong> %s<br>
<strong>Duration:</strong> %.1f days<br>
<strong>Status:</strong> <span style="color:#d9534f;">REJECTED</span>
</p>

<p>Best regards,<br>
<strong>Zenithive Leave Management System</strong></p>
</div>
`, rejectedBy, employeeName, leaveType, startDate, endDate, days)

	for _, email := range AdminEmail {
		if err := SendEmail(email, subject, adminBody); err != nil {
			return err
		}
	}

	return nil
}

// SendLeaveAddedByAdminEmail sends notification to employee when admin/manager adds leave on their behalf
func SendLeaveAddedByAdminEmail(employeeEmail, employeeName, leaveType, startDate, endDate string, days float64, addedBy, addedByRole string) error {
	subject := fmt.Sprintf("Leave Added to Your Account - %s", leaveType)
	body := fmt.Sprintf(`
Dear %s,

A leave has been added to your account by %s (%s).

Leave Type: %s
Start Date: %s
End Date: %s
Duration: %.1f days
Status: APPROVED

This leave has been automatically approved and your leave balance has been updated accordingly.

If you have any questions about this leave entry, please contact your manager or HR department.

Best regards,
Zenithive Leave Management System
`, employeeName, addedBy, addedByRole, leaveType, startDate, endDate, days)

	return SendEmail(employeeEmail, subject, body)
}

// SendPasswordUpdateEmail sends notification to employee when their password is updated by admin
func SendPasswordUpdateEmail(employeeEmail, employeeName, newPassword, updatedByEmail, updatedByRole string) error {
	subject := "Your Password Has Been Updated"
	body := fmt.Sprintf(`
Dear %s,

Your account password has been updated by %s (%s).

Your new login credentials are:
Email: %s
Password: %s

If you did not request this change, please contact your HR department immediately.

For security reasons, we recommend:
1. Login with your new password
2. Change your password to something memorable
3. Keep your password secure and do not share it with anyone

Login URL: [https://zenithiveapp.netlify.app]

Best regards,
Zenithive HR Team
`, employeeName, updatedByEmail, updatedByRole, employeeEmail, newPassword)

	return SendEmail(employeeEmail, subject, body)
}

// SendLeaveCancellationEmail sends notification when leave is cancelled
func SendLeaveCancellationEmail(employeeEmail, employeeName, leaveType, startDate, endDate string, days float64) error {
	subject := "Leave Request Cancelled"
	body := fmt.Sprintf(`
Dear %s,

Your leave request has been cancelled.

Leave Type: %s
Start Date: %s
End Date: %s
Duration: %.1f days
Status: CANCELLED

If you did not cancel this leave request, please contact your manager or HR department immediately.

Best regards,
Zenithive Leave Management System
`, employeeName, leaveType, startDate, endDate, days)

	return SendEmail(employeeEmail, subject, body)
}

// SendLeaveWithdrawalPendingEmail sends notification to admins when manager requests withdrawal
func SendLeaveWithdrawalPendingEmail(recipients []string, employeeName, leaveType, startDate, endDate string, days float64, requestedBy, reason string) error {
	subject := fmt.Sprintf("Leave Withdrawal Request - %s", employeeName)

	reasonText := ""
	if reason != "" {
		reasonText = fmt.Sprintf("\nReason: %s", reason)
	}

	body := fmt.Sprintf(`
Dear Admin,

A leave withdrawal request has been submitted and requires your approval.

Employee: %s
Leave Type: %s
Start Date: %s
End Date: %s
Duration: %.1f days
Requested By: %s (MANAGER)
Status: Pending Withdrawal Approval%s

Please login to the system to approve or reject this withdrawal request.

Best regards,
Zenithive Leave Management System
`, employeeName, leaveType, startDate, endDate, days, requestedBy, reasonText)

	for _, recipient := range recipients {
		if err := SendEmail(recipient, subject, body); err != nil {
			// Log error but continue sending to other recipients
			fmt.Printf("Failed to send email to %s: %v\n", recipient, err)
		}
	}

	return nil
}

// SendLeaveWithdrawalEmail sends notification when approved leave is withdrawn
// SendLeaveWithdrawalEmail sends notification when a leave is withdrawn
func SendLeaveWithdrawalEmail(
	adminEmails []string,
	employeeEmail, employeeName, leaveType, startDate, endDate string,
	days float64, withdrawnBy, withdrawnByRole, reason string,
) error {

	subject := "Leave Request Withdrawn"

	// Optional reason text
	reasonText := ""
	if reason != "" {
		reasonText = fmt.Sprintf("<br><strong>Reason:</strong> %s", reason)
	}

	// ------------------------
	// 1) EMPLOYEE EMAIL
	// ------------------------
	empBody := fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; line-height: 1.6; font-size: 15px;">
<p><strong>Dear %s,</strong></p>

<p>Your approved leave request has been <strong style="color:#f0ad4e;">WITHDRAWN</strong> by %s (%s).</p>

<p>
<strong>Leave Type:</strong> %s<br>
<strong>Start Date:</strong> %s<br>
<strong>End Date:</strong> %s<br>
<strong>Duration:</strong> %.1f days<br>
<strong>Status:</strong> <span style="color:#f0ad4e;">WITHDRAWN</span>%s
</p>

<p>Your leave balance has been restored. The %.1f days have been credited back to your account.</p>

<p>If you have any questions about this withdrawal, please contact your manager or HR department.</p>

<p>Best regards,<br>
<strong>Zenithive Leave Management System</strong></p>
</div>
`, employeeName, withdrawnBy, withdrawnByRole, leaveType, startDate, endDate, days, reasonText, days)

	if err := SendEmail(employeeEmail, subject, empBody); err != nil {
		return err
	}

	// ------------------------
	// 2) ADMIN EMAIL TEMPLATE
	// ------------------------
	adminBody := fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; line-height: 1.6; font-size: 15px;">
<p><strong>Dear Admin,</strong></p>

<p>The leave request of employee <strong>%s</strong> has been <strong style="color:#f0ad4e;">WITHDRAWN</strong> by %s (%s).</p>

<p>
<strong>Leave Type:</strong> %s<br>
<strong>Start Date:</strong> %s<br>
<strong>End Date:</strong> %s<br>
<strong>Duration:</strong> %.1f days<br>
<strong>Status:</strong> <span style="color:#f0ad4e;">WITHDRAWN</span>%s
</p>

<p>The employee's leave balance has been restored.</p>

<p>Best regards,<br>
<strong>Zenithive Leave Management System</strong></p>
</div>
`, employeeName, withdrawnBy, withdrawnByRole, leaveType, startDate, endDate, days, reasonText)

	for _, email := range adminEmails {
		if err := SendEmail(email, subject, adminBody); err != nil {
			return err
		}
	}

	return nil
}

// SendPayslipWithdrawalEmail sends notification when payslip is withdrawn
func SendPayslipWithdrawalEmail(employeeEmail, employeeName string, month, year int, netSalary float64, withdrawnBy, withdrawnByRole, reason string) error {
	monthNames := []string{"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"}

	subject := fmt.Sprintf("Payslip Withdrawn - %s %d", monthNames[month], year)

	reasonText := ""
	if reason != "" {
		reasonText = fmt.Sprintf("\nReason: %s", reason)
	}

	body := fmt.Sprintf(`
Dear %s,

Your payslip for %s %d has been withdrawn by %s (%s).

Pay Period: %s %d
Net Salary: ₹%.2f
Status: WITHDRAWN%s

This payslip has been marked as withdrawn and may require reprocessing. Please contact your HR department or payroll administrator for more information.

If you have any questions about this withdrawal, please reach out to your manager or HR department.

Best regards,
Zenithive Payroll Management System
`, employeeName, monthNames[month], year, withdrawnBy, withdrawnByRole, monthNames[month], year, netSalary, reasonText)

	return SendEmail(employeeEmail, subject, body)
}
