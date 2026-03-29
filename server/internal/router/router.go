package router

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/config"
	"github.com/neinei960/cat/server/internal/handler"
	"github.com/neinei960/cat/server/internal/middleware"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/response"
)

func Setup(mode string) *gin.Engine {
	gin.SetMode(mode)

	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.Cors())
	r.Use(gin.Recovery())

	// Repositories
	shopRepo := repository.NewShopRepository()
	staffRepo := repository.NewStaffRepository()
	customerRepo := repository.NewCustomerRepository()
	customerTagRepo := repository.NewCustomerTagRepository()
	petRepo := repository.NewPetRepository()
	serviceRepo := repository.NewServiceRepository()
	scheduleRepo := repository.NewScheduleRepository()

	apptRepo := repository.NewAppointmentRepository()
	orderRepo := repository.NewOrderRepository()
	notifRepo := repository.NewNotificationRepository()
	statsRepo := repository.NewStatsRepository()

	// Services
	authService := service.NewAuthService(staffRepo, customerRepo)
	shopService := service.NewShopService(shopRepo)
	staffService := service.NewStaffService(staffRepo, scheduleRepo, serviceRepo)
	serviceService := service.NewServiceService(serviceRepo)
	customerService := service.NewCustomerService(customerRepo)
	customerTagService := service.NewCustomerTagService(customerTagRepo)
	petService := service.NewPetService(petRepo)
	apptService := service.NewAppointmentService(apptRepo, scheduleRepo, serviceRepo, staffRepo)
	orderService := service.NewOrderService(orderRepo, apptRepo)
	notifService := service.NewNotificationService(notifRepo, customerRepo)
	dashService := service.NewDashboardService(statsRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	shopHandler := handler.NewShopHandler(shopService)
	staffHandler := handler.NewStaffHandler(staffService)
	serviceHandler := handler.NewServiceHandler(serviceService)
	customerHandler := handler.NewCustomerHandler(customerService, petService)
	customerTagHandler := handler.NewCustomerTagHandler(customerTagService)
	petHandler := handler.NewPetHandler(petService, customerService)
	apptHandler := handler.NewAppointmentHandler(apptService)
	orderHandler := handler.NewOrderHandler(orderService, petService, customerService, serviceService)
	dashHandler := handler.NewDashboardHandler(dashService)
	addonHandler := handler.NewAddonHandler()
	memberCardHandler := handler.NewMemberCardHandler()
	svcCategoryHandler := handler.NewServiceCategoryHandler()
	furCategoryHandler := handler.NewFurCategoryHandler()
	productRepo := repository.NewProductRepository()
	productService := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productService)
	productCategoryHandler := handler.NewProductCategoryHandler()
	cHandler := handler.NewCAppointmentHandler(apptService, serviceService, staffService, petService)
	_ = notifService // used for async notifications in appointment status changes

	// Static files (uploads) - for local dev; nginx handles this in production
	r.Static("/uploads", config.AppConfig.Upload.Path)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", handler.Health)
	}

	// Auth routes (public)
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/staff/login", authHandler.StaffLogin)
		authGroup.POST("/wx/login", authHandler.WxLogin)
	}
	authGroup.POST("/wx/bindphone", middleware.WxAuth(), authHandler.WxBindPhone)

	// B-end routes (JWT auth)
	b := v1.Group("/b", middleware.JWTAuth())
	{
		// Upload
		b.POST("/upload", handler.Upload)

		// Shop
		b.GET("/shop", shopHandler.Get)
		b.PUT("/shop", middleware.RequireRole(model.StaffRoleAdmin), shopHandler.Update)

		// Staff
		b.POST("/staffs", middleware.RequireRole(model.StaffRoleAdmin), staffHandler.Create)
		b.GET("/staffs", staffHandler.List)
		b.GET("/staffs/:id", staffHandler.Get)
		b.PUT("/staffs/:id", middleware.RequireRole(model.StaffRoleAdmin), staffHandler.Update)
		b.DELETE("/staffs/:id", middleware.RequireRole(model.StaffRoleAdmin), staffHandler.Delete)
		b.PUT("/staffs/:id/password", middleware.RequireRole(model.StaffRoleAdmin), staffHandler.ResetPassword)
		b.PUT("/staffs/:id/schedule", middleware.RequireMinRole(model.StaffRoleManager), staffHandler.SetSchedule)
		b.PUT("/staffs/:id/schedule/batch", middleware.RequireMinRole(model.StaffRoleManager), staffHandler.BatchSetSchedule)
		b.GET("/staffs/:id/schedule", staffHandler.GetSchedule)
		b.PUT("/staffs/:id/services", middleware.RequireMinRole(model.StaffRoleManager), staffHandler.SetServices)
		b.GET("/staffs/:id/services", staffHandler.GetServices)

		// Service Categories
		b.POST("/service-categories", middleware.RequireRole("admin"), svcCategoryHandler.Create)
		b.GET("/service-categories", svcCategoryHandler.Tree)
		b.PUT("/service-categories/:id", middleware.RequireRole("admin"), svcCategoryHandler.Update)
		b.DELETE("/service-categories/:id", middleware.RequireRole("admin"), svcCategoryHandler.Delete)

		// Services
		b.POST("/services", middleware.RequireRole("admin"), serviceHandler.Create)
		b.GET("/services", serviceHandler.List)
		b.GET("/services/:id", serviceHandler.Get)
		b.PUT("/services/:id", middleware.RequireRole("admin"), serviceHandler.Update)
		b.DELETE("/services/:id", middleware.RequireRole("admin"), serviceHandler.Delete)
		b.POST("/services/:id/prices", middleware.RequireRole("admin"), serviceHandler.CreatePriceRule)
		b.GET("/services/:id/prices", serviceHandler.GetPriceRules)
		b.DELETE("/services/:id/prices/:rule_id", middleware.RequireRole("admin"), serviceHandler.DeletePriceRule)
		b.POST("/services/:id/discounts", middleware.RequireRole("admin"), serviceHandler.CreateDiscount)
		b.GET("/services/:id/discounts", serviceHandler.GetDiscounts)
		b.DELETE("/services/:id/discounts/:discount_id", middleware.RequireRole("admin"), serviceHandler.DeleteDiscount)

		// Customers
		b.POST("/customers", customerHandler.Create)
		b.GET("/customers", customerHandler.List)
		b.GET("/customers/trash", customerHandler.ListDeleted)
		b.GET("/customers/:id", customerHandler.Get)
		b.PUT("/customers/:id", customerHandler.Update)
		b.DELETE("/customers/:id", middleware.RequireRole("admin"), customerHandler.Delete)
		b.POST("/customers/:id/restore", middleware.RequireRole("admin"), customerHandler.Restore)
		b.GET("/customers/:id/pets", customerHandler.GetPets)

		// Customer Tags
		b.POST("/customer-tags", middleware.RequireRole(model.StaffRoleAdmin), customerTagHandler.Create)
		b.GET("/customer-tags", customerTagHandler.List)
		b.PUT("/customer-tags/:id", middleware.RequireRole(model.StaffRoleAdmin), customerTagHandler.Update)
		b.DELETE("/customer-tags/:id", middleware.RequireRole(model.StaffRoleAdmin), customerTagHandler.Delete)

		// Pets
		b.POST("/pets", petHandler.Create)
		b.GET("/pets", petHandler.List)
		b.GET("/pets/:id", petHandler.Get)
		b.PUT("/pets/:id", petHandler.Update)
		b.DELETE("/pets/:id", petHandler.Delete)

		// Member Card Templates
		b.POST("/member-card-templates", middleware.RequireRole("admin"), memberCardHandler.CreateTemplate)
		b.GET("/member-card-templates", memberCardHandler.ListTemplates)
		b.PUT("/member-card-templates/:id", middleware.RequireRole("admin"), memberCardHandler.UpdateTemplate)
		b.DELETE("/member-card-templates/:id", middleware.RequireRole("admin"), memberCardHandler.DeleteTemplate)
		b.PUT("/member-card-templates/:id/discounts", middleware.RequireRole("admin"), memberCardHandler.SetDiscounts)

		// Member Card Operations (on customer)
		b.POST("/customers/:id/member-card", memberCardHandler.OpenCard)
		b.POST("/customers/:id/recharge", memberCardHandler.Recharge)
		b.GET("/customers/:id/member-card", memberCardHandler.GetCard)
		b.PUT("/customers/:id/adjust-balance", middleware.RequireMinRole(model.StaffRoleManager), memberCardHandler.AdjustBalance)
		b.GET("/customers/:id/recharge-records", memberCardHandler.GetRecords)
		b.PUT("/recharge-records/:id", middleware.RequireMinRole(model.StaffRoleManager), memberCardHandler.UpdateRecord)
		b.DELETE("/recharge-records/:id", middleware.RequireMinRole(model.StaffRoleManager), memberCardHandler.DeleteRecord)

		// Service Addons
		b.POST("/addons", middleware.RequireRole("admin"), addonHandler.Create)
		b.GET("/addons", addonHandler.List)
		b.PUT("/addons/:id", middleware.RequireRole("admin"), addonHandler.Update)
		b.DELETE("/addons/:id", middleware.RequireRole("admin"), addonHandler.Delete)

		// Fur Categories
		b.POST("/fur-categories", middleware.RequireRole("admin"), furCategoryHandler.Create)
		b.GET("/fur-categories", furCategoryHandler.List)
		b.PUT("/fur-categories/:id", middleware.RequireRole("admin"), furCategoryHandler.Update)
		b.DELETE("/fur-categories/:id", middleware.RequireRole("admin"), furCategoryHandler.Delete)

		// Products
		b.GET("/products/brands", productHandler.GetBrands)
		b.POST("/products", middleware.RequireRole("admin"), productHandler.Create)
		b.GET("/products", productHandler.List)
		b.GET("/products/:id", productHandler.Get)
		b.PUT("/products/:id", middleware.RequireRole("admin"), productHandler.Update)
		b.DELETE("/products/:id", middleware.RequireRole("admin"), productHandler.Delete)

		// Product Categories
		b.POST("/product-categories", middleware.RequireRole("admin"), productCategoryHandler.Create)
		b.GET("/product-categories", productCategoryHandler.List)
		b.PUT("/product-categories/:id", middleware.RequireRole("admin"), productCategoryHandler.Update)
		b.DELETE("/product-categories/:id", middleware.RequireRole("admin"), productCategoryHandler.Delete)

		// Appointments
		b.GET("/appointments/slots", apptHandler.GetSlots)
		b.GET("/appointments/calendar", apptHandler.Calendar)
		b.POST("/appointments", apptHandler.Create)
		b.GET("/appointments", apptHandler.List)
		b.GET("/appointments/:id", apptHandler.Get)
		b.PUT("/appointments/:id", apptHandler.Update)
		b.PUT("/appointments/:id/status", apptHandler.UpdateStatus)
		b.PUT("/appointments/:id/assign", middleware.RequireMinRole(model.StaffRoleManager), apptHandler.AssignStaff)
		b.PUT("/appointments/:id/reschedule", middleware.RequireMinRole(model.StaffRoleManager), apptHandler.Reschedule)

		// Service Records
		svcRecordRepo := repository.NewServiceRecordRepository()
		petBathReportRepo := repository.NewPetBathReportRepository()
		b.POST("/service-records", func(c *gin.Context) {
			var req struct {
				AppointmentID uint   `json:"appointment_id" binding:"required"`
				PetID         uint   `json:"pet_id" binding:"required"`
				Notes         string `json:"notes"`
				Photos        string `json:"photos"`
				SkinIssues    string `json:"skin_issues"`
				FurCondition  string `json:"fur_condition"`
				Weight        string `json:"weight"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				response.Error(c, 400, "参数错误")
				return
			}
			record := &model.ServiceRecord{
				ShopID:        c.GetUint("shop_id"),
				AppointmentID: req.AppointmentID,
				PetID:         req.PetID,
				StaffID:       c.GetUint("staff_id"),
				Notes:         req.Notes,
				Photos:        req.Photos,
				SkinIssues:    req.SkinIssues,
				FurCondition:  req.FurCondition,
				Weight:        req.Weight,
			}
			if err := svcRecordRepo.Create(record); err != nil {
				response.Error(c, 500, "保存失败")
				return
			}
			response.Success(c, record)
		})
		b.GET("/service-records", func(c *gin.Context) {
			apptID, _ := strconv.ParseUint(c.Query("appointment_id"), 10, 64)
			petID, _ := strconv.ParseUint(c.Query("pet_id"), 10, 64)
			if apptID > 0 {
				records, _ := svcRecordRepo.FindByAppointment(uint(apptID))
				response.Success(c, records)
			} else if petID > 0 {
				records, _ := svcRecordRepo.FindByPet(uint(petID), 20)
				response.Success(c, records)
			} else {
				response.Error(c, 400, "请提供appointment_id或pet_id")
			}
		})

		// Pet Bath Reports
		b.GET("/pets/:id/bath-reports", func(c *gin.Context) {
			petID, err := strconv.ParseUint(c.Param("id"), 10, 64)
			if err != nil || petID == 0 {
				response.Error(c, 400, "宠物ID错误")
				return
			}
			reports, err := petBathReportRepo.FindByPet(c.GetUint("shop_id"), uint(petID))
			if err != nil {
				response.Error(c, 500, "查询失败")
				return
			}
			response.Success(c, reports)
		})
		b.POST("/pets/:id/bath-reports", func(c *gin.Context) {
			petID, err := strconv.ParseUint(c.Param("id"), 10, 64)
			if err != nil || petID == 0 {
				response.Error(c, 400, "宠物ID错误")
				return
			}
			var req struct {
				ImageURL string `json:"image_url" binding:"required"`
				BathDate string `json:"bath_date"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				response.Error(c, 400, "参数错误")
				return
			}
			now := time.Now()
			bathDate := &now
			if req.BathDate != "" {
				parsed, err := time.Parse("2006-01-02", req.BathDate)
				if err != nil {
					response.Error(c, 400, "洗浴日期格式错误")
					return
				}
				bathDate = &parsed
			}
			report := &model.PetBathReport{
				ShopID:   c.GetUint("shop_id"),
				PetID:    uint(petID),
				ImageURL: req.ImageURL,
				BathDate: bathDate,
			}
			if err := petBathReportRepo.Create(report); err != nil {
				response.Error(c, 500, "保存失败")
				return
			}
			response.Success(c, report)
		})
		b.PUT("/pets/:id/bath-reports/:report_id", func(c *gin.Context) {
			petID, err := strconv.ParseUint(c.Param("id"), 10, 64)
			if err != nil || petID == 0 {
				response.Error(c, 400, "宠物ID错误")
				return
			}
			reportID, err := strconv.ParseUint(c.Param("report_id"), 10, 64)
			if err != nil || reportID == 0 {
				response.Error(c, 400, "报告ID错误")
				return
			}
			var req struct {
				BathDate string `json:"bath_date" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				response.Error(c, 400, "参数错误")
				return
			}
			parsed, err := time.Parse("2006-01-02", req.BathDate)
			if err != nil {
				response.Error(c, 400, "洗浴日期格式错误")
				return
			}
			if err := petBathReportRepo.UpdateBathDate(c.GetUint("shop_id"), uint(petID), uint(reportID), &parsed); err != nil {
				response.Error(c, 500, "更新失败")
				return
			}
			response.Success(c, gin.H{"updated": true, "bath_date": req.BathDate})
		})
		b.DELETE("/pets/:id/bath-reports/:report_id", func(c *gin.Context) {
			petID, err := strconv.ParseUint(c.Param("id"), 10, 64)
			if err != nil || petID == 0 {
				response.Error(c, 400, "宠物ID错误")
				return
			}
			reportID, err := strconv.ParseUint(c.Param("report_id"), 10, 64)
			if err != nil || reportID == 0 {
				response.Error(c, 400, "报告ID错误")
				return
			}
			if err := petBathReportRepo.Delete(c.GetUint("shop_id"), uint(petID), uint(reportID)); err != nil {
				response.Error(c, 500, "删除失败")
				return
			}
			response.Success(c, gin.H{"deleted": true})
		})

		// Orders
		b.POST("/orders", orderHandler.Create)
		b.POST("/orders/from-appointment", orderHandler.CreateFromAppointment)
		b.POST("/orders/from-appointment/batch", orderHandler.CreateBatchFromAppointment)
		b.GET("/orders", orderHandler.List)
		b.GET("/orders/price-lookup", orderHandler.PriceLookup)
		b.GET("/orders/:id", orderHandler.Get)
		b.PUT("/orders/:id/pay", orderHandler.Pay)
		b.PUT("/orders/:id/refund", middleware.RequireMinRole(model.StaffRoleManager), orderHandler.Refund)
		b.PUT("/orders/:id/cancel", middleware.RequireMinRole(model.StaffRoleManager), orderHandler.Cancel)

		// Dashboard
		b.GET("/dashboard/overview", dashHandler.Overview)
		b.GET("/dashboard/revenue", middleware.RequireMinRole(model.StaffRoleManager), dashHandler.Revenue)
		b.GET("/dashboard/services", dashHandler.ServiceRanking)
		b.GET("/dashboard/staff", middleware.RequireMinRole(model.StaffRoleManager), dashHandler.StaffPerformance)
		b.GET("/dashboard/category", dashHandler.CategoryStats)
		b.GET("/dashboard/members", dashHandler.MemberStats)
		b.POST("/dashboard/aggregate", middleware.RequireMinRole(model.StaffRoleManager), dashHandler.Aggregate)
	}

	// C-end routes (WeChat auth)
	cg := v1.Group("/c", middleware.WxAuth())
	{
		cg.GET("/services", cHandler.ListServices)
		cg.GET("/staffs", cHandler.ListStaffs)
		cg.GET("/slots", cHandler.GetSlots)
		cg.POST("/appointments", cHandler.CreateAppointment)
		cg.GET("/appointments", cHandler.ListAppointments)
		cg.GET("/appointments/:id", cHandler.GetAppointment)
		cg.PUT("/appointments/:id/cancel", cHandler.CancelAppointment)
		cg.GET("/pets", cHandler.ListPets)
		cg.POST("/pets", cHandler.CreatePet)
		cg.PUT("/pets/:id", cHandler.UpdatePet)
	}

	// Public callback routes
	_ = v1.Group("/public")
	// WeChat pay callbacks etc

	return r
}
