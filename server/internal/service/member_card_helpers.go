package service

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
)

func hasActiveMemberCard(customerID *uint) bool {
	if customerID == nil || *customerID == 0 {
		return false
	}

	var count int64
	if err := database.DB.Model(&model.MemberCard{}).
		Where("customer_id = ? AND status = 1", *customerID).
		Count(&count).Error; err != nil {
		return false
	}
	return count > 0
}

func normalizeAppointmentDeposit(customerID *uint, deposit, totalAmount float64) float64 {
	if hasActiveMemberCard(customerID) {
		return 0
	}

	totalAmount = roundOrderAmount(totalAmount)
	if totalAmount <= 0 {
		return 0
	}

	deposit = roundOrderAmount(deposit)
	if deposit <= 0 {
		return 0
	}
	if deposit > totalAmount {
		return totalAmount
	}
	return deposit
}

func applyAppointmentDepositToOrder(order *model.Order, appointment *model.Appointment) {
	if order == nil {
		return
	}

	order.AppointmentDepositAmount = 0
	order.AppointmentDepositDeductionAmount = 0
	if appointment == nil {
		return
	}

	rawDeposit := roundOrderAmount(appointment.Deposit)
	order.AppointmentDepositAmount = rawDeposit
	if rawDeposit <= 0 || hasActiveMemberCard(&appointment.CustomerID) {
		return
	}

	deduction := rawDeposit
	if order.AppointmentIsLate {
		deduction = roundOrderAmount(rawDeposit * 0.7)
	}
	if deduction > order.PayAmount {
		deduction = order.PayAmount
	}
	deduction = roundOrderAmount(deduction)
	if deduction <= 0 {
		return
	}

	order.AppointmentDepositDeductionAmount = deduction
	order.PayAmount = roundOrderAmount(order.PayAmount - deduction)
	if order.PayAmount < 0 {
		order.PayAmount = 0
	}
}
