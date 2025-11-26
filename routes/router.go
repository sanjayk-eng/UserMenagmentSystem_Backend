package routes

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/controllers"
	middleware "github.com/sanjayk-eng/UserMenagmentSystem_Backend/middlewere"
)

func SetupRoutes(r *gin.Engine, h *controllers.HandlerFunc) {

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{h.Env.FRONTEND_SERVER},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Authorization", "token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ----------------- Auth -----------------
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", h.Login)
	}

	// ----------------- Employees -----------------
	employees := r.Group("/api/employee")
	employees.Use(middleware.AuthMiddleware(h)) // Protect employee routes
	{
		employees.GET("/", h.GetEmployee) // List all employees (SUPER_ADMIN, ADMIN/HR)
		//employees.GET("/:id", h.GetEmployeeByID)                 // Get employee details (Self/Manager/Admin)
		employees.POST("/", h.CreateEmployee) // Create employee (SUPER_ADMIN, ADMIN/HR)
		//employees.PATCH("/:id", h.UpdateEmployeeInfo)            // Update employee info (SUPER_ADMIN, ADMIN/HR)
		employees.PATCH("/:id/role", h.UpdateEmployeeRole) // Change employee role (SUPER_ADMIN, ADMIN/HR)
		employees.PATCH("/:id/manager", h.UpdateEmployeeManager)
		employees.PUT("/deactivate/:id", h.DeleteEmployeeStatus) // Set/change manager (SUPER_ADMIN, ADMIN/HR)
		employees.GET("/:id/reports", h.GetEmployeeReports)      // Get direct reports (Self/Manager/Admin)
	}

	// ----------------- Leaves -----------------
	leaves := r.Group("/api/leaves")
	leaves.Use(middleware.AuthMiddleware(h))
	{
		leaves.POST("/apply", h.ApplyLeave) // Employee applies for leave
		leaves.POST("/admin-add", h.AdminAddLeave)
		leaves.POST("/admin-add/policy", h.AdminAddLeavePolicy)
		leaves.GET("/Get-All-Leave-Policy", h.GetAllLeavePolicies) // Admin/Manager adds leave on behalf of employee
		leaves.POST("/:id/action", h.ActionLeave)                  // Approve/Reject leave
		leaves.GET("/all", h.GetAllLeaves)
		//leaves.GET("/:id", h.GetLeaveByID)         // Get leave details
	}

	// ----------------- Leave Balances -----------------
	leaveBalances := r.Group("/api/leave-balances")
	leaveBalances.Use(middleware.AuthMiddleware(h))
	{

		leaveBalances.GET("/employee/:id", h.GetLeaveBalances)  // GET /api/employees/:id/leave-balances
		leaveBalances.POST("/:id/adjust", h.AdjustLeaveBalance) // POST /api/leave-balances/:id/adjust
	}

	// ----------------- Payroll -----------------
	payroll := r.Group("/api/payroll")
	payroll.Use(middleware.AuthMiddleware(h))
	{
		// Run payroll for a given month & year
		payroll.POST("/run", h.RunPayroll)
		// POST /api/payroll/run

		// Finalize payroll for a specific payroll run ID
		payroll.POST("/:id/finalize", h.FinalizePayroll)
		// POST /api/payroll/{id}/finalize

		payroll.GET("/payslip", h.GetFinalizedPayslips)

		// Download payslip PDF for a specific employee payslip ID
		payroll.GET("/payslips/:id/pdf", h.GetPayslipPDF)
		// GET /api/payroll/payslips/{id}/pdf
	}

	// ----------------- Settings -----------------
	settings := r.Group("/api/settings")
	settings.Use(middleware.AuthMiddleware(h)) // Only admin/superadmin
	{
		settings.GET("/", h.GetCompanySettings)    // Get current settings
		settings.PUT("/", h.UpdateCompanySettings) // Update settings
	}
	holidays := r.Group("/api/settings/holidays")
	holidays.Use(middleware.AuthMiddleware(h))
	{
		holidays.POST("/", h.AddHoliday)         // SUPERADMIN adds holiday
		holidays.GET("/", h.GetHolidays)         // List all holidays
		holidays.DELETE("/:id", h.DeleteHoliday) // Remove holiday
	}
}
