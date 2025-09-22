package media

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// MediaType represents the type of media
type MediaType string

const (
	MediaTypeImage    MediaType = "image"
	MediaTypeVideo    MediaType = "video"
	MediaTypeAudio    MediaType = "audio"
	MediaTypeDocument MediaType = "document"
)

// Media represents associated media files
type Media struct {
	base.BaseModel

	// Tenant isolation
	OrganizationID string `json:"organization_id" gorm:"type:varchar(255);not null;index:idx_tenant"`

	// Parent catalog item or variant
	CatalogItemID *string `json:"catalog_item_id" gorm:"type:varchar(255);index:idx_catalog_item"`
	VariantID     *string `json:"variant_id" gorm:"type:varchar(255);index:idx_variant"`

	// Media details
	Type        MediaType `json:"type" gorm:"type:varchar(20);not null;check:type IN ('image', 'video', 'audio', 'document')"`
	Title       string    `json:"title" gorm:"type:varchar(255)"`
	Description string    `json:"description" gorm:"type:text"`

	// File information
	URL      string `json:"url" gorm:"type:varchar(500);not null"`
	FileName string `json:"file_name" gorm:"type:varchar(255);not null"`
	FileSize int64  `json:"file_size" gorm:"check:file_size >= 0"`
	MimeType string `json:"mime_type" gorm:"type:varchar(100);not null"`

	// Dimensions for images/videos
	Width  *int `json:"width" gorm:"check:width IS NULL OR width > 0"`
	Height *int `json:"height" gorm:"check:height IS NULL OR height > 0"`

	// Ordering and display
	SortOrder int  `json:"sort_order" gorm:"not null;default:0"`
	IsPrimary bool `json:"is_primary" gorm:"not null;default:false"`
	IsActive  bool `json:"is_active" gorm:"not null;default:true;index:idx_active"`

	// Alt text for accessibility
	AltText string `json:"alt_text" gorm:"type:varchar(255)"`

	// Audit fields
	CreatedBy string `json:"created_by" gorm:"type:varchar(255);not null"`
	UpdatedBy string `json:"updated_by" gorm:"type:varchar(255);not null"`

	// Soft delete support
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index:idx_media_deleted"`
}

// TableName returns the table name for GORM
func (Media) TableName() string {
	return "media"
}

// NewMedia creates a new media record
func NewMedia(orgID, url, fileName string, mediaType MediaType) *Media {
	return &Media{
		BaseModel:      *base.NewBaseModel("MED", "large"),
		OrganizationID: orgID,
		Type:           mediaType,
		URL:            url,
		FileName:       fileName,
		SortOrder:      0,
		IsPrimary:      false,
		IsActive:       true,
	}
}

// IsImage checks if the media is an image
func (m *Media) IsImage() bool {
	return m.Type == MediaTypeImage
}

// IsVideo checks if the media is a video
func (m *Media) IsVideo() bool {
	return m.Type == MediaTypeVideo
}

// GetDisplayURL returns the URL for display purposes
func (m *Media) GetDisplayURL() string {
	return m.URL
}
