package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	FeedingPlanStatusDraft     = "draft"
	FeedingPlanStatusActive    = "active"
	FeedingPlanStatusPaused    = "paused"
	FeedingPlanStatusCompleted = "completed"
	FeedingPlanStatusCancelled = "cancelled"

	FeedingVisitStatusPending    = "pending"
	FeedingVisitStatusAssigned   = "assigned"
	FeedingVisitStatusInProgress = "in_progress"
	FeedingVisitStatusDone       = "done"
	FeedingVisitStatusException  = "exception"
	FeedingVisitStatusCancelled  = "cancelled"

	FeedingWindowMorning   = "morning"
	FeedingWindowAfternoon = "afternoon"
	FeedingWindowEvening   = "evening"
	FeedingWindowAllDay    = "all_day"
)

type FeedingSetting struct {
	gorm.Model
	ShopID      uint   `json:"shop_id" gorm:"not null;uniqueIndex"`
	PricingJSON string `json:"pricing_json" gorm:"type:text"`
	ItemsJSON   string `json:"items_json" gorm:"type:text"`
	UpdatedBy   uint   `json:"updated_by" gorm:"default:0"`
}

type FeedingPlan struct {
	gorm.Model
	ShopID              uint              `json:"shop_id" gorm:"not null;index"`
	CustomerID          uint              `json:"customer_id" gorm:"not null;index"`
	OrderID             *uint             `json:"order_id,omitempty" gorm:"index"`
	AddressSnapshotJSON string            `json:"address_snapshot_json" gorm:"type:text"`
	ContactName         string            `json:"contact_name" gorm:"size:100"`
	ContactPhone        string            `json:"contact_phone" gorm:"size:20"`
	StartDate           string            `json:"start_date" gorm:"size:10;not null;index"`
	EndDate             string            `json:"end_date" gorm:"size:10;not null;index"`
	TimeGranularity     string            `json:"time_granularity" gorm:"size:20;default:window"`
	Status              string            `json:"status" gorm:"size:20;default:active;index"`
	Remark              string            `json:"remark" gorm:"type:text"`
	PricingSnapshotJSON string            `json:"pricing_snapshot_json" gorm:"type:text"`
	SelectedItemsJSON   string            `json:"selected_items_json" gorm:"type:text"`
	SelectedDatesJSON   string            `json:"selected_dates_json" gorm:"type:text"`
	PlayDatesJSON       string            `json:"play_dates_json" gorm:"type:text"`
	PlayMode            string            `json:"play_mode" gorm:"size:20"`
	PlayCount           int               `json:"play_count" gorm:"default:0"`
	OtherPrice          float64           `json:"other_price" gorm:"type:decimal(10,2);default:0"`
	Deposit             float64           `json:"deposit" gorm:"type:decimal(10,2);default:0"`
	TotalAmount         float64           `json:"total_amount" gorm:"type:decimal(10,2);default:0"`
	UnpaidAmount        float64           `json:"unpaid_amount" gorm:"type:decimal(10,2);default:0"`
	Customer            *Customer         `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Order               *Order            `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	Pets                []FeedingPlanPet  `json:"pets,omitempty" gorm:"foreignKey:FeedingPlanID"`
	Rules               []FeedingPlanRule `json:"rules,omitempty" gorm:"foreignKey:FeedingPlanID"`
	Visits              []FeedingVisit    `json:"visits,omitempty" gorm:"foreignKey:FeedingPlanID"`
}

type FeedingPlanPet struct {
	gorm.Model
	FeedingPlanID   uint   `json:"feeding_plan_id" gorm:"not null;index"`
	PetID           uint   `json:"pet_id" gorm:"not null;index"`
	PetNameSnapshot string `json:"pet_name_snapshot" gorm:"size:100"`
	Remark          string `json:"remark" gorm:"size:500"`
	Pet             *Pet   `json:"pet,omitempty" gorm:"foreignKey:PetID"`
}

type FeedingPlanRule struct {
	gorm.Model
	FeedingPlanID uint   `json:"feeding_plan_id" gorm:"not null;index"`
	Weekday       int    `json:"weekday" gorm:"not null;index;comment:0-6 对应周日到周六"`
	WindowCode    string `json:"window_code" gorm:"size:20;not null;index"`
	VisitCount    int    `json:"visit_count" gorm:"default:1"`
}

type FeedingVisit struct {
	gorm.Model
	ShopID        uint                `json:"shop_id" gorm:"not null;index"`
	FeedingPlanID uint                `json:"feeding_plan_id" gorm:"not null;index"`
	ScheduledDate string              `json:"scheduled_date" gorm:"size:10;not null;index"`
	WindowCode    string              `json:"window_code" gorm:"size:20;not null;index"`
	StaffID       *uint               `json:"staff_id,omitempty" gorm:"index"`
	Status        string              `json:"status" gorm:"size:20;default:pending;index"`
	VisitPrice    float64             `json:"visit_price" gorm:"type:decimal(10,2);default:0"`
	ArrivedAt     *time.Time          `json:"arrived_at,omitempty"`
	CompletedAt   *time.Time          `json:"completed_at,omitempty"`
	CustomerNote  string              `json:"customer_note" gorm:"type:text"`
	InternalNote  string              `json:"internal_note" gorm:"type:text"`
	ExceptionType string              `json:"exception_type" gorm:"size:50"`
	Plan          *FeedingPlan        `json:"plan,omitempty" gorm:"foreignKey:FeedingPlanID"`
	Staff         *Staff              `json:"staff,omitempty" gorm:"foreignKey:StaffID"`
	Items         []FeedingVisitItem  `json:"items,omitempty" gorm:"foreignKey:FeedingVisitID"`
	Logs          []FeedingVisitLog   `json:"logs,omitempty" gorm:"foreignKey:FeedingVisitID"`
	Media         []FeedingVisitMedia `json:"media,omitempty" gorm:"foreignKey:FeedingVisitID"`
}

type FeedingVisitItem struct {
	gorm.Model
	FeedingVisitID   uint    `json:"feeding_visit_id" gorm:"not null;index"`
	ItemCode         string  `json:"item_code" gorm:"size:50;not null;index"`
	ItemNameSnapshot string  `json:"item_name_snapshot" gorm:"size:100"`
	ExtraPrice       float64 `json:"extra_price" gorm:"type:decimal(10,2);default:0"`
	Checked          bool    `json:"checked" gorm:"default:false"`
}

type FeedingVisitLog struct {
	gorm.Model
	FeedingVisitID uint   `json:"feeding_visit_id" gorm:"not null;index"`
	OperatorID     uint   `json:"operator_id" gorm:"not null;index"`
	Action         string `json:"action" gorm:"size:50;not null"`
	Content        string `json:"content" gorm:"type:text"`
	Operator       *Staff `json:"operator,omitempty" gorm:"foreignKey:OperatorID"`
}

type FeedingVisitMedia struct {
	gorm.Model
	FeedingVisitID uint   `json:"feeding_visit_id" gorm:"not null;index"`
	MediaType      string `json:"media_type" gorm:"size:20;not null"`
	URL            string `json:"url" gorm:"size:500;not null"`
}
