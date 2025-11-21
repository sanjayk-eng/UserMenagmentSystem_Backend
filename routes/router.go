package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/controllers"
	middleware "github.com/sanjayk-eng/UserMenagmentSystem_Backend/middlewere"
)

func SetupRoutes(r *gin.Engine, h *controllers.HandlerFunc) {

	// ----------------- Auth -----------------
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", h.Login)
	}

	// ----------------- Employees -----------------
	employees := r.Group("/api/employee")
	employees.Use(middleware.AuthMiddleware(h)) // Protect employee routes
	{
		employees.GET("", h.GetEmployee)                         // GET get all employee by Admin ,SuperAdmin and HR
		employees.POST("", h.CreateEmployee)                     // Add Employee by Admin ,SuperAdmin and HR
		employees.PATCH("/:id/role", h.UpdateEmployeeRole)       // PATCH /api/employees/:id/role
		employees.PATCH("/:id/manager", h.UpdateEmployeeManager) // PATCH /api/employees/:id/manager
		employees.GET("/:id/reports", h.GetEmployeeReports)      // GET /api/employees/:id/reports
	}

	// ----------------- Leaves -----------------
	leaves := r.Group("/api/leaves")
	leaves.Use(middleware.AuthMiddleware(h))
	{
		leaves.POST("/apply", h.ApplyLeave)
		leaves.GET("/", h.GetAllLeavePolicies)     //show all leave policy
		leaves.POST("/admin-add", h.AdminAddLeave) // Admin adds leave
		leaves.POST("/:id/action", h.ActionLeave)  // Approve/Reject leave
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
		payroll.POST("/run", h.RunPayroll)                // POST /api/payroll/run
		payroll.POST("/:id/finalize", h.FinalizePayroll)  // POST /api/payroll/:id/finalize
		payroll.GET("/payslips/:id/pdf", h.GetPayslipPDF) // GET /api/payslips/:id/pdf
	}

	// ----------------- Settings -----------------
	settings := r.Group("/api/settings")
	settings.Use(middleware.AuthMiddleware(h))
	{
		settings.GET("/company", h.GetCompanySettings)
		settings.POST("/company", h.UpdateCompanySettings)
		settings.GET("/permissions", h.GetPermissions)
		settings.POST("/permissions", h.UpdatePermissions)
	}
}
