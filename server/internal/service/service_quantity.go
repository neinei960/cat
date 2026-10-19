package service

import (
	"fmt"
	"github.com/neinei960/cat/server/internal/model"
	"strings"
)

// Annual package purchases are single-unit; per-visit card redemptions remain repeatable.
func ValidateServiceQuantities(items []model.OrderItem) error {
	packageCounts := map[string]int{}
	for _, item := range items {
		if item.ItemType != 1 {
			continue
		}
		if item.Quantity < 1 || item.Quantity > 999 {
			return fmt.Errorf("服务次数须为1至999次")
		}
		if strings.Contains(item.Name, "全年口腔护理") {
			var petID uint
			if item.PetID != nil {
				petID = *item.PetID
			}
			key := fmt.Sprintf("%d:%d", petID, item.ItemID)
			packageCounts[key] += item.Quantity
			if packageCounts[key] > 1 {
				return fmt.Errorf("全年口腔护理每只猫限购一份")
			}
		}
	}
	return nil
}
