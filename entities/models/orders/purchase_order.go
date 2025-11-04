package orders

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

// POSource represents the source of a purchase order
type POSource string

const (
	POSourceKisanlink POSource = "KISANLINK" // Auto-created from marketplace order
	POSourceManual    POSource = "MANUAL"    // Manually created by FPO
)

// POStatus represents the possible purchase order states
type POStatus string

const (
	POStatusPlaced    POStatus = "PLACED"    // PO created
	POStatusConfirmed POStatus = "CONFIRMED" // Vendor confirmed
	POStatusDelivered POStatus = "DELIVERED" // Goods received
	POStatusPaid      POStatus = "PAID"      // Payment completed
	POStatusCancelled POStatus = "CANCELLED" // PO cancelled
)

// GRNStatus represents the status of goods receipt note
type GRNStatus string

const (
	GRNStatusPending   GRNStatus = "PENDING"   // GRN created, awaiting verification
	GRNStatusConfirmed GRNStatus = "CONFIRMED" // GRN confirmed, goods accepted
	GRNStatusRejected  GRNStatus = "REJECTED"  // GRN rejected, quality issues
)

// ItemCondition represents the condition of received items
type ItemCondition string

const (
	ItemConditionGood    ItemCondition = "GOOD"    // Good condition
	ItemConditionDamaged ItemCondition = "DAMAGED" // Damaged items
	ItemConditionPartial ItemCondition = "PARTIAL" // Partial delivery
)

// PurchaseOrder represents an FPO purchase order
type PurchaseOrder struct {
	base.BaseModel

	FPOOrgID        string          `json:"fpo_org_id" gorm:"type:varchar(255);not null;index"`
	PONumber        string          `json:"po_number" gorm:"type:varchar(50);uniqueIndex;not null"`
	Source          POSource        `json:"source" gorm:"type:varchar(20);not null"`
	VendorID        sql.NullString  `json:"vendor_id,omitempty" gorm:"type:varchar(255)"` // Collaborator ID (null for manual)
	VendorName      string          `json:"vendor_name" gorm:"type:varchar(255);not null"`
	VendorContact   string          `json:"vendor_contact" gorm:"type:varchar(100)"`
	TotalAmount     decimal.Decimal `json:"total_amount" gorm:"type:decimal(12,2);not null;default:0.00"`
	Status          POStatus        `json:"status" gorm:"type:varchar(20);not null;default:'PLACED'"`
	DeliveryAddress datatypes.JSON  `json:"delivery_address" gorm:"type:jsonb"`
	Notes           string          `json:"notes" gorm:"type:text"`
	OrderID         sql.NullString  `json:"order_id,omitempty" gorm:"type:varchar(255)"` // Link to Kisanlink order if source=KISANLINK

	// Relationships
	Items []POItem `json:"items,omitempty" gorm:"foreignKey:POID;constraint:OnDelete:CASCADE"`
	GRNs  []GRN    `json:"grns,omitempty" gorm:"foreignKey:POID;constraint:OnDelete:CASCADE"`

	// Metadata
	Metadata sql.NullString `json:"metadata" gorm:"type:jsonb"`
}

// TableName returns the table name for GORM
func (PurchaseOrder) TableName() string {
	return "purchase_orders"
}

// NewPurchaseOrder creates a new purchase order instance
func NewPurchaseOrder(fpoOrgID, vendorName string, source POSource) *PurchaseOrder {
	return &PurchaseOrder{
		BaseModel:   *base.NewBaseModel("PO", "large"),
		FPOOrgID:    fpoOrgID,
		PONumber:    generatePONumber(),
		Source:      source,
		VendorName:  vendorName,
		Status:      POStatusPlaced,
		TotalAmount: decimal.Zero,
	}
}

// generatePONumber generates a unique PO number (PO-YYYYMMDD-XXXXX)
func generatePONumber() string {
	now := time.Now()
	timestamp := now.UnixNano() / 1000000 // milliseconds
	return fmt.Sprintf("PO-%s-%05d", now.Format("20060102"), timestamp%100000)
}

// POItem represents an item in a purchase order
type POItem struct {
	base.BaseModel

	POID        string          `json:"po_id" gorm:"type:varchar(255);not null;index"`
	ProductName string          `json:"product_name" gorm:"type:varchar(255);not null"`
	ProductSKU  string          `json:"product_sku" gorm:"type:varchar(100)"`
	HSNCode     string          `json:"hsn_code" gorm:"type:varchar(20)"`
	Quantity    decimal.Decimal `json:"quantity" gorm:"type:decimal(10,3);not null"`
	UnitPrice   decimal.Decimal `json:"unit_price" gorm:"type:decimal(10,2);not null"`
	GSTPercent  decimal.Decimal `json:"gst_percent" gorm:"type:decimal(5,2);default:0.00"`
	TotalAmount decimal.Decimal `json:"total_amount" gorm:"type:decimal(12,2);not null"`

	// Metadata
	Metadata sql.NullString `json:"metadata" gorm:"type:jsonb"`
}

// TableName returns the table name for GORM
func (POItem) TableName() string {
	return "purchase_order_items"
}

// NewPOItem creates a new purchase order item
func NewPOItem(poID, productName, productSKU, hsnCode string, quantity, unitPrice, gstPercent decimal.Decimal) *POItem {
	totalAmount := quantity.Mul(unitPrice)
	gstAmount := totalAmount.Mul(gstPercent).Div(decimal.NewFromInt(100))
	totalAmount = totalAmount.Add(gstAmount)

	return &POItem{
		BaseModel:   *base.NewBaseModel("POITEM", "large"),
		POID:        poID,
		ProductName: productName,
		ProductSKU:  productSKU,
		HSNCode:     hsnCode,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		GSTPercent:  gstPercent,
		TotalAmount: totalAmount,
	}
}

// GRN represents a Goods Received Note
type GRN struct {
	base.BaseModel

	POID         string    `json:"po_id" gorm:"type:varchar(255);not null;index"`
	GRNNumber    string    `json:"grn_number" gorm:"type:varchar(50);uniqueIndex;not null"`
	ReceivedDate time.Time `json:"received_date" gorm:"not null"`
	ReceivedBy   string    `json:"received_by" gorm:"type:varchar(255);not null"` // User ID
	Status       GRNStatus `json:"status" gorm:"type:varchar(20);not null;default:'PENDING'"`
	Notes        string    `json:"notes" gorm:"type:text"`

	// Relationships
	Items []GRNItem `json:"items,omitempty" gorm:"foreignKey:GRNID;constraint:OnDelete:CASCADE"`

	// Metadata
	Metadata sql.NullString `json:"metadata" gorm:"type:jsonb"`
}

// TableName returns the table name for GORM
func (GRN) TableName() string {
	return "goods_received_notes"
}

// NewGRN creates a new GRN instance
func NewGRN(poID, receivedBy string) *GRN {
	return &GRN{
		BaseModel:    *base.NewBaseModel("GRN", "large"),
		POID:         poID,
		GRNNumber:    generateGRNNumber(),
		ReceivedDate: time.Now(),
		ReceivedBy:   receivedBy,
		Status:       GRNStatusPending,
	}
}

// generateGRNNumber generates a unique GRN number
func generateGRNNumber() string {
	now := time.Now()
	timestamp := now.UnixNano() / 1000000 // milliseconds
	return fmt.Sprintf("GRN-%s-%05d", now.Format("20060102"), timestamp%100000)
}

// GRNItem represents an item in a goods received note
type GRNItem struct {
	base.BaseModel

	GRNID            string          `json:"grn_id" gorm:"type:varchar(255);not null;index"`
	POItemID         string          `json:"po_item_id" gorm:"type:varchar(255);not null"`
	QuantityOrdered  decimal.Decimal `json:"quantity_ordered" gorm:"type:decimal(10,3);not null"`
	QuantityReceived decimal.Decimal `json:"quantity_received" gorm:"type:decimal(10,3);not null"`
	Condition        ItemCondition   `json:"condition" gorm:"type:varchar(20);not null"`
	Notes            string          `json:"notes" gorm:"type:text"`

	// Metadata
	Metadata sql.NullString `json:"metadata" gorm:"type:jsonb"`
}

// TableName returns the table name for GORM
func (GRNItem) TableName() string {
	return "grn_items"
}

// NewGRNItem creates a new GRN item
func NewGRNItem(grnID, poItemID string, quantityOrdered, quantityReceived decimal.Decimal, condition ItemCondition) *GRNItem {
	return &GRNItem{
		BaseModel:        *base.NewBaseModel("GRNITEM", "large"),
		GRNID:            grnID,
		POItemID:         poItemID,
		QuantityOrdered:  quantityOrdered,
		QuantityReceived: quantityReceived,
		Condition:        condition,
	}
}

// PurchaseOrder methods

// CalculateTotal calculates the total amount for the purchase order
func (po *PurchaseOrder) CalculateTotal() {
	var total decimal.Decimal
	for _, item := range po.Items {
		total = total.Add(item.TotalAmount)
	}
	po.TotalAmount = total
}

// CanTransitionTo checks if the PO can transition to the given status
func (po *PurchaseOrder) CanTransitionTo(newStatus POStatus) bool {
	switch po.Status {
	case POStatusPlaced:
		return newStatus == POStatusConfirmed || newStatus == POStatusCancelled
	case POStatusConfirmed:
		return newStatus == POStatusDelivered || newStatus == POStatusCancelled
	case POStatusDelivered:
		return newStatus == POStatusPaid
	case POStatusPaid, POStatusCancelled:
		return false // Terminal states
	default:
		return false
	}
}

// Address handling methods

// SetDeliveryAddress sets the delivery address from an Address struct
func (po *PurchaseOrder) SetDeliveryAddress(address *Address) error {
	if address == nil {
		po.DeliveryAddress = nil
		return nil
	}

	addressJSON, err := json.Marshal(address)
	if err != nil {
		return fmt.Errorf("failed to marshal delivery address: %w", err)
	}

	po.DeliveryAddress = addressJSON
	return nil
}

// GetDeliveryAddress returns the delivery address as an Address struct
func (po *PurchaseOrder) GetDeliveryAddress() (*Address, error) {
	if len(po.DeliveryAddress) == 0 {
		return nil, nil
	}

	var address Address
	if err := json.Unmarshal(po.DeliveryAddress, &address); err != nil {
		return nil, fmt.Errorf("failed to unmarshal delivery address: %w", err)
	}

	return &address, nil
}

// Metadata handling methods

// SetMetadata sets the metadata from a map
func (po *PurchaseOrder) SetMetadata(metadata map[string]interface{}) error {
	if metadata == nil {
		po.Metadata = sql.NullString{Valid: false}
		return nil
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	po.Metadata = sql.NullString{String: string(metadataJSON), Valid: true}
	return nil
}

// GetMetadata returns the metadata as a map
func (po *PurchaseOrder) GetMetadata() (map[string]interface{}, error) {
	if !po.Metadata.Valid || po.Metadata.String == "" {
		return nil, nil
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(po.Metadata.String), &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return metadata, nil
}

// POItem metadata handling methods

// SetMetadata sets the metadata from a map
func (pi *POItem) SetMetadata(metadata map[string]interface{}) error {
	if metadata == nil {
		pi.Metadata = sql.NullString{Valid: false}
		return nil
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	pi.Metadata = sql.NullString{String: string(metadataJSON), Valid: true}
	return nil
}

// GetMetadata returns the metadata as a map
func (pi *POItem) GetMetadata() (map[string]interface{}, error) {
	if !pi.Metadata.Valid || pi.Metadata.String == "" {
		return nil, nil
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(pi.Metadata.String), &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return metadata, nil
}

// GRN metadata handling methods

// SetMetadata sets the metadata from a map
func (grn *GRN) SetMetadata(metadata map[string]interface{}) error {
	if metadata == nil {
		grn.Metadata = sql.NullString{Valid: false}
		return nil
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	grn.Metadata = sql.NullString{String: string(metadataJSON), Valid: true}
	return nil
}

// GetMetadata returns the metadata as a map
func (grn *GRN) GetMetadata() (map[string]interface{}, error) {
	if !grn.Metadata.Valid || grn.Metadata.String == "" {
		return nil, nil
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(grn.Metadata.String), &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return metadata, nil
}

// GRNItem metadata handling methods

// SetMetadata sets the metadata from a map
func (gi *GRNItem) SetMetadata(metadata map[string]interface{}) error {
	if metadata == nil {
		gi.Metadata = sql.NullString{Valid: false}
		return nil
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	gi.Metadata = sql.NullString{String: string(metadataJSON), Valid: true}
	return nil
}

// GetMetadata returns the metadata as a map
func (gi *GRNItem) GetMetadata() (map[string]interface{}, error) {
	if !gi.Metadata.Valid || gi.Metadata.String == "" {
		return nil, nil
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(gi.Metadata.String), &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return metadata, nil
}
