package service

import (
	"log"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/pkg/wechat"
)

type NotificationService struct {
	repo         *repository.NotificationRepository
	customerRepo *repository.CustomerRepository
}

func NewNotificationService(repo *repository.NotificationRepository, customerRepo *repository.CustomerRepository) *NotificationService {
	return &NotificationService{repo: repo, customerRepo: customerRepo}
}

func (s *NotificationService) List(shopID uint, page, pageSize int) ([]model.NotificationLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.FindByShop(shopID, page, pageSize)
}

// SendAppointmentConfirm sends confirmation notification to customer
func (s *NotificationService) SendAppointmentConfirm(shopID, customerID uint, date, startTime, serviceName string) {
	go func() {
		customer, err := s.customerRepo.FindByID(customerID)
		if err != nil || customer.OpenID == "" {
			return
		}

		logEntry := &model.NotificationLog{
			ShopID:     shopID,
			CustomerID: customerID,
			Scene:      "confirm",
			Channel:    "wechat",
			Content:    "预约已确认: " + date + " " + startTime + " " + serviceName,
		}

		msg := &wechat.TemplateMessage{
			ToUser:     customer.OpenID,
			TemplateID: "", // configure in production
			Data: map[string]map[string]string{
				"thing1":  {"value": serviceName},
				"date2":   {"value": date + " " + startTime},
				"thing3":  {"value": "您的预约已确认"},
			},
		}

		if err := wechat.SendTemplateMessage(msg); err != nil {
			log.Printf("send notification failed: %v", err)
			logEntry.Status = 2
			logEntry.ErrorMsg = err.Error()
		} else {
			logEntry.Status = 1
		}
		s.repo.Create(logEntry)
	}()
}

// SendAppointmentReminder sends reminder before appointment
func (s *NotificationService) SendAppointmentReminder(shopID, customerID uint, date, startTime, serviceName string) {
	go func() {
		customer, err := s.customerRepo.FindByID(customerID)
		if err != nil || customer.OpenID == "" {
			return
		}

		logEntry := &model.NotificationLog{
			ShopID:     shopID,
			CustomerID: customerID,
			Scene:      "reminder",
			Channel:    "wechat",
			Content:    "预约提醒: " + date + " " + startTime + " " + serviceName,
		}

		msg := &wechat.TemplateMessage{
			ToUser:     customer.OpenID,
			TemplateID: "",
			Data: map[string]map[string]string{
				"thing1":  {"value": serviceName},
				"date2":   {"value": date + " " + startTime},
				"thing3":  {"value": "请按时到店"},
			},
		}

		if err := wechat.SendTemplateMessage(msg); err != nil {
			logEntry.Status = 2
			logEntry.ErrorMsg = err.Error()
		} else {
			logEntry.Status = 1
		}
		s.repo.Create(logEntry)
	}()
}

// SendAppointmentComplete sends completion notification
func (s *NotificationService) SendAppointmentComplete(shopID, customerID uint, serviceName string) {
	go func() {
		customer, err := s.customerRepo.FindByID(customerID)
		if err != nil || customer.OpenID == "" {
			return
		}

		logEntry := &model.NotificationLog{
			ShopID:     shopID,
			CustomerID: customerID,
			Scene:      "complete",
			Channel:    "wechat",
			Content:    "服务已完成: " + serviceName,
		}

		msg := &wechat.TemplateMessage{
			ToUser:     customer.OpenID,
			TemplateID: "",
			Data: map[string]map[string]string{
				"thing1":  {"value": serviceName},
				"thing2":  {"value": "服务已完成，欢迎再次光临"},
			},
		}

		if err := wechat.SendTemplateMessage(msg); err != nil {
			logEntry.Status = 2
			logEntry.ErrorMsg = err.Error()
		} else {
			logEntry.Status = 1
		}
		s.repo.Create(logEntry)
	}()
}

// SendAppointmentCancel sends cancellation notification
func (s *NotificationService) SendAppointmentCancel(shopID, customerID uint, date, startTime, reason string) {
	go func() {
		customer, err := s.customerRepo.FindByID(customerID)
		if err != nil || customer.OpenID == "" {
			return
		}

		logEntry := &model.NotificationLog{
			ShopID:     shopID,
			CustomerID: customerID,
			Scene:      "cancel",
			Channel:    "wechat",
			Content:    "预约已取消: " + date + " " + startTime,
		}

		msg := &wechat.TemplateMessage{
			ToUser:     customer.OpenID,
			TemplateID: "",
			Data: map[string]map[string]string{
				"thing1":  {"value": date + " " + startTime},
				"thing2":  {"value": reason},
			},
		}

		if err := wechat.SendTemplateMessage(msg); err != nil {
			logEntry.Status = 2
			logEntry.ErrorMsg = err.Error()
		} else {
			logEntry.Status = 1
		}
		s.repo.Create(logEntry)
	}()
}
