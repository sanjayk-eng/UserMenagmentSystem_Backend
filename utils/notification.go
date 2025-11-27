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

var GOOGLE_SCRIPT_URL = os.Getenv("GOOGLE_SCRIPT_URL")

// SendEmail sends an email using Google Apps Script
func SendEmail(to, subject, body string) error {
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
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(GOOGLE_SCRIPT_URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("email service returned status: %d", resp.StatusCode)
	}

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

Login URL: [Your Frontend URL]

If you have any questions, please contact your HR department.

Best regards,
Zenithive HR Team
`, employeeName, employeeEmail, password)

	return SendEmail(employeeEmail, subject, body)
}

// SendLeaveApplicationEmail sends notification to manager, admin, and superadmin
func SendLeaveApplicationEmail(recipients []string, employeeName, leaveType, startDate, endDate string, days float64) error {
	subject := fmt.Sprintf("Leave Application - %s", employeeName)
	body := fmt.Sprintf(`
Dear Manager/Admin,

A new leave application has been submitted and requires your review.

Employee: %s
Leave Type: %s
Start Date: %s
End Date: %s
Duration: %.1f days
Status: Pending Approval

Please login to the system to approve or reject this leave request.

Best regards,
Zenithive Leave Management System
`, employeeName, leaveType, startDate, endDate, days)

	for _, recipient := range recipients {
		if err := SendEmail(recipient, subject, body); err != nil {
			// Log error but continue sending to other recipients
			fmt.Printf("Failed to send email to %s: %v\n", recipient, err)
		}
	}

	return nil
}

// SendLeaveApprovalEmail sends notification to employee when leave is approved
func SendLeaveApprovalEmail(employeeEmail, employeeName, leaveType, startDate, endDate string, days float64) error {
	subject := "Leave Approved"
	body := fmt.Sprintf(`
Dear %s,

Your leave application has been approved.

Leave Type: %s
Start Date: %s
End Date: %s
Duration: %.1f days
Status: APPROVED

Enjoy your time off!

Best regards,
Zenithive Leave Management System
`, employeeName, leaveType, startDate, endDate, days)

	return SendEmail(employeeEmail, subject, body)
}

// SendLeaveRejectionEmail sends notification to employee when leave is rejected
func SendLeaveRejectionEmail(employeeEmail, employeeName, leaveType, startDate, endDate string, days float64) error {
	subject := "Leave Request Rejected"
	body := fmt.Sprintf(`
Dear %s,

We regret to inform you that your leave application has been rejected.

Leave Type: %s
Start Date: %s
End Date: %s
Duration: %.1f days
Status: REJECTED

Please contact your manager for more information.

Best regards,
Zenithive Leave Management System
`, employeeName, leaveType, startDate, endDate, days)

	return SendEmail(employeeEmail, subject, body)
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
func SendPasswordUpdateEmail(employeeEmail, employeeName, updatedBy, updatedByRole string) error {
	subject := "Your Password Has Been Updated"
	body := fmt.Sprintf(`
Dear %s,

Your account password has been updated by %s (%s).

If you did not request this change, please contact your HR department immediately.

For security reasons, we recommend:
1. Login with your new password
2. Change your password to something memorable
3. Keep your password secure and do not share it with anyone

Login URL: [Your Frontend URL]

Best regards,
Zenithive HR Team
`, employeeName, updatedBy, updatedByRole)

	return SendEmail(employeeEmail, subject, body)
}
