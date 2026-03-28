package repository

import (
	"strings"
	"time"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
	"gorm.io/gorm"
)

type CustomerRepository struct{}

func NewCustomerRepository() *CustomerRepository {
	return &CustomerRepository{}
}

func (r *CustomerRepository) Create(customer *model.Customer) error {
	return database.DB.Create(customer).Error
}

func (r *CustomerRepository) FindByID(id uint) (*model.Customer, error) {
	var customer model.Customer
	err := database.DB.Preload("MemberCard.Template").
		Preload("Pets", "deleted_at IS NULL").
		Preload("CustomerTags", "status = 1").
		First(&customer, id).Error
	return &customer, err
}

func (r *CustomerRepository) FindByOpenID(openID string) (*model.Customer, error) {
	var customer model.Customer
	err := database.DB.Where("open_id = ?", openID).First(&customer).Error
	return &customer, err
}

func (r *CustomerRepository) FindByPhone(phone string, shopID uint) (*model.Customer, error) {
	var customer model.Customer
	err := database.DB.Where("phone = ? AND shop_id = ?", phone, shopID).First(&customer).Error
	return &customer, err
}

func (r *CustomerRepository) FindByShopID(shopID uint, page, pageSize int, memberCardTemplateID uint, customerTagID uint) ([]model.Customer, int64, error) {
	return r.listCustomers(shopID, "", memberCardTemplateID, customerTagID, page, pageSize)
}

func (r *CustomerRepository) Search(shopID uint, keyword string, page, pageSize int, memberCardTemplateID uint, customerTagID uint) ([]model.Customer, int64, error) {
	return r.listCustomers(shopID, keyword, memberCardTemplateID, customerTagID, page, pageSize)
}

func (r *CustomerRepository) buildCustomerListQuery(shopID uint, keyword string, memberCardTemplateID uint, customerTagID uint) *gorm.DB {
	db := database.DB.Model(&model.Customer{}).Where("customers.shop_id = ?", shopID)
	if memberCardTemplateID > 0 {
		db = db.Joins("JOIN member_cards ON member_cards.id = customers.member_card_id AND member_cards.deleted_at IS NULL").
			Where("member_cards.template_id = ? AND member_cards.status = 1", memberCardTemplateID)
	}
	if customerTagID > 0 {
		db = db.Joins("JOIN customer_tag_relations ON customer_tag_relations.customer_id = customers.id").
			Where("customer_tag_relations.tag_id = ?", customerTagID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Joins("LEFT JOIN pets ON pets.customer_id = customers.id AND pets.deleted_at IS NULL").
			Where("customers.nickname LIKE ? OR customers.phone LIKE ? OR customers.remark LIKE ? OR pets.name LIKE ?",
				like, like, like, like)
	}
	return db.Group("customers.id")
}

func (r *CustomerRepository) listCustomers(shopID uint, keyword string, memberCardTemplateID uint, customerTagID uint, page, pageSize int) ([]model.Customer, int64, error) {
	var customers []model.Customer
	var total int64

	countDB := r.buildCustomerListQuery(shopID, keyword, memberCardTemplateID, customerTagID)
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	type idRow struct{ ID uint }
	var idRows []idRow
	idDB := r.buildCustomerListQuery(shopID, keyword, memberCardTemplateID, customerTagID)
	// 按最近到店时间排序，相同则按会员卡等级（储值门槛）高的优先
	idDB = idDB.Select("customers.id, MAX(customers.last_visit_at) AS last_visit_at, COALESCE(MAX(mct_sort.min_recharge), 0) AS sort_recharge").
		Joins("LEFT JOIN member_cards mc_sort ON mc_sort.id = customers.member_card_id AND mc_sort.deleted_at IS NULL").
		Joins("LEFT JOIN member_card_templates mct_sort ON mct_sort.id = mc_sort.template_id AND mct_sort.deleted_at IS NULL").
		Order("last_visit_at DESC").
		Order("sort_recharge DESC").
		Order("customers.id DESC")
	if err := idDB.Offset(offset).Limit(pageSize).Find(&idRows).Error; err != nil {
		return nil, total, err
	}
	ids := make([]uint, len(idRows))
	for i, r := range idRows {
		ids[i] = r.ID
	}
	if len(ids) == 0 {
		return []model.Customer{}, total, nil
	}

	if err := database.DB.Preload("MemberCard.Template").
		Preload("Pets", "deleted_at IS NULL").
		Preload("CustomerTags", "status = 1").
		Where("id IN ?", ids).Find(&customers).Error; err != nil {
		return nil, total, err
	}

	customerMap := make(map[uint]model.Customer, len(customers))
	for _, customer := range customers {
		customerMap[customer.ID] = customer
	}

	ordered := make([]model.Customer, 0, len(ids))
	for _, id := range ids {
		if customer, ok := customerMap[id]; ok {
			ordered = append(ordered, customer)
		}
	}

	return ordered, total, nil
}

func (r *CustomerRepository) Update(customer *model.Customer) error {
	return database.DB.Omit("CustomerTags").Save(customer).Error
}

func (r *CustomerRepository) SetTags(customer *model.Customer, tagIDs []uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var tags []model.CustomerTag
		if len(tagIDs) > 0 {
			if err := tx.Where("shop_id = ? AND id IN ?", customer.ShopID, tagIDs).
				Order("sort_order ASC, id ASC").
				Find(&tags).Error; err != nil {
				return err
			}
		}

		if err := tx.Model(customer).Association("CustomerTags").Replace(&tags); err != nil {
			return err
		}

		customer.CustomerTags = tags
		customer.Tags = buildCustomerTagSummary(tags)
		return tx.Model(customer).Update("tags", customer.Tags).Error
	})
}

func (r *CustomerRepository) Delete(id uint) error {
	return database.DB.Delete(&model.Customer{}, id).Error
}

func (r *CustomerRepository) FindDeleted(shopID uint, page, pageSize int) ([]model.Customer, int64, error) {
	var customers []model.Customer
	var total int64
	db := database.DB.Unscoped().Model(&model.Customer{}).Where("shop_id = ? AND deleted_at IS NOT NULL", shopID)
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("deleted_at DESC").Offset(offset).Limit(pageSize).Find(&customers).Error
	return customers, total, err
}

func (r *CustomerRepository) Restore(id uint) error {
	return database.DB.Unscoped().Model(&model.Customer{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *CustomerRepository) CleanupExpired(before time.Time) (int64, error) {
	result := database.DB.Unscoped().Where("deleted_at IS NOT NULL AND deleted_at < ?", before).Delete(&model.Customer{})
	return result.RowsAffected, result.Error
}

func buildCustomerTagSummary(tags []model.CustomerTag) string {
	if len(tags) == 0 {
		return ""
	}
	names := make([]string, 0, len(tags))
	for _, tag := range tags {
		if tag.Name == "" {
			continue
		}
		names = append(names, tag.Name)
	}
	return strings.Join(names, ",")
}
