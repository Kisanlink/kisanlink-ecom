package catalog

import (
	"time"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

// ProductAttributes represents type-specific attributes for products
type ProductAttributes struct {
	SKU           string            `json:"sku" validate:"required"`
	Brand         string            `json:"brand"`
	Weight        *decimal.Decimal  `json:"weight"`
	Dimensions    *Dimensions       `json:"dimensions"`
	Variants      []ProductVariant  `json:"variants"`
	Perishable    bool              `json:"perishable"`
	ShelfLife     *int              `json:"shelf_life_days"`
	StorageTemp   *TemperatureRange `json:"storage_temperature"`
	Organic       bool              `json:"organic"`
	Certification []string          `json:"certification"`
}

// ServiceAttributes represents type-specific attributes for services
type ServiceAttributes struct {
	Duration      string         `json:"duration" validate:"required"`
	SLA           SLAInfo        `json:"sla"`
	Location      string         `json:"location"`
	Skills        []string       `json:"skills"`
	ServiceArea   GeographicArea `json:"service_area"`
	Availability  []TimeSlot     `json:"availability"`
	Equipment     []string       `json:"equipment_required"`
	Certification []string       `json:"certification"`
}

// LabourAttributes represents type-specific attributes for labour
type LabourAttributes struct {
	Skills        []string       `json:"skills" validate:"required"`
	Experience    int            `json:"experience"`
	UnitRate      Money          `json:"unit_rate" validate:"required"`
	RateType      string         `json:"rate_type" validate:"oneof=hourly daily weekly monthly"`
	Availability  []TimeSlot     `json:"availability"`
	Location      GeographicArea `json:"location"`
	Languages     []string       `json:"languages"`
	Tools         []string       `json:"tools_owned"`
	Certification []string       `json:"certification"`
}

// ContractAttributes represents type-specific attributes for contracts
type ContractAttributes struct {
	Term         string        `json:"term" validate:"required"`
	Duration     int           `json:"duration" validate:"required"`
	StartDate    time.Time     `json:"start_date"`
	EndDate      time.Time     `json:"end_date"`
	Terms        string        `json:"terms"`
	Deliverables []Deliverable `json:"deliverables"`
	Milestones   []Milestone   `json:"milestones"`
	PaymentTerms PaymentTerms  `json:"payment_terms"`
	Penalties    []Penalty     `json:"penalties"`
	Renewals     RenewalTerms  `json:"renewal_terms"`
}

// Supporting structures

// Dimensions represents physical dimensions
type Dimensions struct {
	Length decimal.Decimal `json:"length"`
	Width  decimal.Decimal `json:"width"`
	Height decimal.Decimal `json:"height"`
	Unit   string          `json:"unit"` // cm, m, inch, ft
}

// ProductVariant represents a product variant
type ProductVariant struct {
	Name       string                 `json:"name"`
	SKU        string                 `json:"sku"`
	Price      decimal.Decimal        `json:"price"`
	Attributes map[string]interface{} `json:"attributes"`
}

// TemperatureRange represents temperature storage requirements
type TemperatureRange struct {
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
	Unit string  `json:"unit"` // celsius, fahrenheit
}

// SLAInfo represents SLA information
type SLAInfo struct {
	ResponseTime   int     `json:"response_time_minutes"`
	ResolutionTime int     `json:"resolution_time_hours"`
	Availability   float64 `json:"availability_percentage"`
	Support        string  `json:"support_hours"`
}

// GeographicArea represents a geographic service area
type GeographicArea struct {
	Type        string    `json:"type"` // city, state, country, radius
	Value       string    `json:"value"`
	Coordinates *Location `json:"coordinates,omitempty"`
	Radius      *float64  `json:"radius,omitempty"` // in km
}

// Location represents geographic coordinates
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// TimeSlot represents availability time slots
type TimeSlot struct {
	DayOfWeek string    `json:"day_of_week"` // MON, TUE, etc.
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Available bool      `json:"available"`
}

// Money represents monetary values
type Money struct {
	Amount   decimal.Decimal `json:"amount"`
	Currency string          `json:"currency"`
}

// Deliverable represents contract deliverables
type Deliverable struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	Status      string    `json:"status"`
	Value       Money     `json:"value"`
}

// Milestone represents contract milestones
type Milestone struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	Status      string    `json:"status"`
	Payment     Money     `json:"payment"`
}

// PaymentTerms represents payment terms
type PaymentTerms struct {
	Type       string         `json:"type"` // advance, milestone, completion, monthly
	Percentage float64        `json:"percentage"`
	DueDays    int            `json:"due_days"`
	Method     pq.StringArray `json:"method"`
	Currency   string         `json:"currency"`
}

// Penalty represents contract penalties
type Penalty struct {
	Type        string          `json:"type"` // delay, quality, breach
	Description string          `json:"description"`
	Amount      decimal.Decimal `json:"amount"`
	Percentage  *float64        `json:"percentage,omitempty"`
}

// RenewalTerms represents contract renewal terms
type RenewalTerms struct {
	AutoRenewal     bool     `json:"auto_renewal"`
	NoticePeriod    int      `json:"notice_period_days"`
	RenewalPeriod   int      `json:"renewal_period_months"`
	PriceAdjustment *float64 `json:"price_adjustment_percentage,omitempty"`
}
