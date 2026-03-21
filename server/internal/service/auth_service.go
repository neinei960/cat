package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/pkg/auth"
	"github.com/neinei960/cat/server/pkg/wechat"
)

type AuthService struct {
	staffRepo    *repository.StaffRepository
	customerRepo *repository.CustomerRepository
}

func NewAuthService(staffRepo *repository.StaffRepository, customerRepo *repository.CustomerRepository) *AuthService {
	return &AuthService{staffRepo: staffRepo, customerRepo: customerRepo}
}

type StaffLoginResult struct {
	Token string      `json:"token"`
	Staff *model.Staff `json:"staff"`
}

func (s *AuthService) StaffLogin(phone, password string) (*StaffLoginResult, error) {
	staff, err := s.staffRepo.FindByPhone(phone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("手机号或密码错误")
		}
		return nil, err
	}

	if staff.Status != 1 {
		return nil, errors.New("账号已停用")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("手机号或密码错误")
	}

	token, err := auth.GenerateToken(staff.ID, staff.ShopID, staff.Role)
	if err != nil {
		return nil, errors.New("生成token失败")
	}

	return &StaffLoginResult{Token: token, Staff: staff}, nil
}

type WxLoginResult struct {
	Token    string          `json:"token"`
	Customer *model.Customer `json:"customer"`
	IsNew    bool            `json:"is_new"`
}

func (s *AuthService) WxLogin(code string, shopID uint) (*WxLoginResult, error) {
	session, err := wechat.Code2Session(code)
	if err != nil {
		return nil, errors.New("微信登录失败: " + err.Error())
	}

	customer, err := s.customerRepo.FindByOpenID(session.OpenID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	isNew := false
	if errors.Is(err, gorm.ErrRecordNotFound) {
		customer = &model.Customer{
			ShopID:  shopID,
			OpenID:  session.OpenID,
			UnionID: session.UnionID,
		}
		if err := s.customerRepo.Create(customer); err != nil {
			return nil, errors.New("创建用户失败")
		}
		isNew = true
	}

	token, err := auth.GenerateCustomerToken(customer.ID, customer.ShopID, customer.OpenID)
	if err != nil {
		return nil, errors.New("生成token失败")
	}

	return &WxLoginResult{Token: token, Customer: customer, IsNew: isNew}, nil
}

func (s *AuthService) WxBindPhone(customerID uint, phone string) error {
	customer, err := s.customerRepo.FindByID(customerID)
	if err != nil {
		return errors.New("用户不存在")
	}
	customer.Phone = phone
	return s.customerRepo.Update(customer)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
