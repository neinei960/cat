package model

import (
	"time"

	"gorm.io/gorm"
)

type Customer struct {
	gorm.Model
	ShopID        uint       `json:"shop_id" gorm:"not null;index"`
	OpenID        string     `json:"openid" gorm:"size:100;index"`
	UnionID       string     `json:"unionid" gorm:"size:100"`
	Phone         string     `json:"phone" gorm:"size:20;index"`
	Nickname      string     `json:"nickname" gorm:"size:100"`
	Avatar        string     `json:"avatar" gorm:"size:500"`
	Gender        int        `json:"gender" gorm:"default:0;comment:0未知 1男 2女"`
	Remark        string     `json:"remark" gorm:"size:500"`
	Tags          string     `json:"tags" gorm:"size:500;comment:逗号分隔标签"`
	TotalSpent    float64    `json:"total_spent" gorm:"type:decimal(10,2);default:0"`
	VisitCount    int        `json:"visit_count" gorm:"default:0"`
	LastVisitAt   *time.Time `json:"last_visit_at"`
	MemberBalance float64    `json:"member_balance" gorm:"type:decimal(10,2);default:0;comment:会员余额"`
	DiscountRate  float64    `json:"discount_rate" gorm:"type:decimal(3,2);default:1;comment:折扣率 1/0.9/0.86"`
	MemberCardID  *uint      `json:"member_card_id" gorm:"index;comment:当前会员卡ID"`

	Shop         *Shop         `json:"shop,omitempty" gorm:"foreignKey:ShopID"`
	MemberCard   *MemberCard   `json:"member_card,omitempty" gorm:"foreignKey:MemberCardID"`
	Pets         []Pet         `json:"pets,omitempty" gorm:"foreignKey:CustomerID"`
	CustomerTags []CustomerTag `json:"customer_tags,omitempty" gorm:"many2many:customer_tag_relations;joinForeignKey:CustomerID;joinReferences:TagID"`
}
