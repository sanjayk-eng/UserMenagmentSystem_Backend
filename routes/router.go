package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/controllers"
	middleware "github.com/sanjayk-eng/UserMenagmentSystem_Backend/middlewere"
)

func SetupRoutes(r *gin.Engine) {

	// ----------------- Auth -----------------
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", controllers.Login)
	}

	// ----------------- Employees -----------------
	employees := r.Group("/api/employees")
	employees.Use(middleware.AuthMiddleware()) // Protect employee routes
	{
		employees.GET("", controllers.GetEmployees)                        // GET /api/employees
		employees.PATCH("/:id/role", controllers.UpdateEmployeeRole)       // PATCH /api/employees/:id/role
		employees.PATCH("/:id/manager", controllers.UpdateEmployeeManager) // PATCH /api/employees/:id/manager
		employees.GET("/:id/reports", controllers.GetEmployeeReports)      // GET /api/employees/:id/reports
	}

	// ----------------- Leaves -----------------
	leaves := r.Group("/api/leaves")
	leaves.Use(middleware.AuthMiddleware())
	{
		leaves.POST("/apply", controllers.ApplyLeave)        // Employee applies leave
		leaves.POST("/admin-add", controllers.AdminAddLeave) // Admin adds leave
		leaves.POST("/:id/action", controllers.ActionLeave)  // Approve/Reject leave
	}

	// ----------------- Leave Balances -----------------
	leaveBalances := r.Group("/api/leave-balances")
	leaveBalances.Use(middleware.AuthMiddleware())
	{
		leaveBalances.GET("/employee/:id", controllers.GetLeaveBalances)  // GET /api/employees/:id/leave-balances
		leaveBalances.POST("/:id/adjust", controllers.AdjustLeaveBalance) // POST /api/leave-balances/:id/adjust
	}

	// ----------------- Payroll -----------------
	payroll := r.Group("/api/payroll")
	payroll.Use(middleware.AuthMiddleware())
	{
		payroll.POST("/run", controllers.RunPayroll)                // POST /api/payroll/run
		payroll.POST("/:id/finalize", controllers.FinalizePayroll)  // POST /api/payroll/:id/finalize
		payroll.GET("/payslips/:id/pdf", controllers.GetPayslipPDF) // GET /api/payslips/:id/pdf
	}

	// ----------------- Settings -----------------
	settings := r.Group("/api/settings")
	settings.Use(middleware.AuthMiddleware())
	{
		settings.GET("/company", controllers.GetCompanySettings)
		settings.POST("/company", controllers.UpdateCompanySettings)
		settings.GET("/permissions", controllers.GetPermissions)
		settings.POST("/permissions", controllers.UpdatePermissions)
	}
}
