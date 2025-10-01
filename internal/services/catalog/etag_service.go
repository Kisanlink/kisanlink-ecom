package catalog

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	catalogModels "kisanlink-ecom/entities/models/catalog"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ETagService provides ETag generation and validation for catalog items
type ETagService struct {
	logger *logrus.Logger
}

// NewETagService creates a new ETag service
func NewETagService(logger *logrus.Logger) *ETagService {
	if logger == nil {
		logger = logrus.StandardLogger()
	}

	return &ETagService{
		logger: logger,
	}
}

// GenerateETag generates an ETag for a catalog item
func (s *ETagService) GenerateETag(item *catalogModels.CatalogItem) string {
	if item == nil {
		return ""
	}

	// Create a hash based on entity ID, version, and updated timestamp
	data := fmt.Sprintf("%s:%d:%d", item.ID, item.Version, item.UpdatedAt.Unix())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf(`"%s"`, hex.EncodeToString(hash[:8])) // Use first 8 bytes for shorter ETag
}

// CheckETagMatch checks if the client's ETag matches the current item ETag
// Returns true if ETags match (indicating client has current version)
func (s *ETagService) CheckETagMatch(c *gin.Context, item *catalogModels.CatalogItem) bool {
	ifNoneMatch := c.GetHeader("If-None-Match")
	if ifNoneMatch == "" {
		return false
	}

	currentETag := s.GenerateETag(item)

	// Handle multiple ETags in If-None-Match (comma-separated)
	etags := s.ParseIfNoneMatch(ifNoneMatch)

	// Check if current ETag matches any of the provided ETags
	for _, etag := range etags {
		if etag == "*" || etag == currentETag {
			s.logger.WithFields(logrus.Fields{
				"item_id":       item.ID,
				"current_etag":  currentETag,
				"if_none_match": ifNoneMatch,
				"matched_etag":  etag,
			}).Debug("ETag match found - returning 304 Not Modified")
			return true
		}
	}

	return false
}

// HandleConditionalRequest handles conditional GET requests with ETag validation
// Returns true if a 304 Not Modified response was sent
func (s *ETagService) HandleConditionalRequest(c *gin.Context, item *catalogModels.CatalogItem) bool {
	if item == nil {
		return false
	}

	// Generate current ETag
	currentETag := s.GenerateETag(item)

	// Check if client's ETag matches
	if s.CheckETagMatch(c, item) {
		// Client has current version, return 304 Not Modified
		s.ReturnNotModified(c, currentETag)
		return true
	}

	// Set ETag header for response
	s.SetETagHeaders(c, item)

	s.logger.WithFields(logrus.Fields{
		"item_id":           item.ID,
		"item_type":         item.ItemType,
		"etag":              currentETag,
		"has_if_none_match": c.GetHeader("If-None-Match") != "",
	}).Debug("Generated ETag for catalog item")

	return false
}

// SetETagHeaders sets ETag and cache control headers for a catalog item response
func (s *ETagService) SetETagHeaders(c *gin.Context, item *catalogModels.CatalogItem) {
	if item == nil {
		return
	}

	// Generate and set ETag
	etag := s.GenerateETag(item)
	c.Header("ETag", etag)

	// Set cache control headers for catalog items
	c.Header("Cache-Control", "private, must-revalidate, max-age=300") // 5 minutes

	s.logger.WithFields(logrus.Fields{
		"item_id":   item.ID,
		"item_type": item.ItemType,
		"etag":      etag,
	}).Debug("Set ETag headers for catalog item")
}

// ReturnNotModified returns a 304 Not Modified response
func (s *ETagService) ReturnNotModified(c *gin.Context, etag string) {
	// Set ETag header
	c.Header("ETag", etag)

	// Set cache control headers
	c.Header("Cache-Control", "private, must-revalidate")

	s.logger.WithFields(logrus.Fields{
		"path":   c.Request.URL.Path,
		"etag":   etag,
		"status": http.StatusNotModified,
	}).Info("Returning 304 Not Modified response")

	c.Status(http.StatusNotModified)
	c.Abort()
}

// ValidateETagFormat validates ETag format
func (s *ETagService) ValidateETagFormat(etag string) error {
	if etag == "" {
		return nil // Empty ETag is valid (no caching)
	}

	// ETag should be quoted
	if !strings.HasPrefix(etag, `"`) || !strings.HasSuffix(etag, `"`) {
		return fmt.Errorf("ETag must be quoted")
	}

	// Extract content without quotes
	content := strings.Trim(etag, `"`)
	if len(content) == 0 {
		return fmt.Errorf("ETag content cannot be empty")
	}

	return nil
}

// ParseIfNoneMatch parses the If-None-Match header value
// Handles both single ETags and comma-separated lists
func (s *ETagService) ParseIfNoneMatch(ifNoneMatch string) []string {
	// Handle wildcard
	if strings.TrimSpace(ifNoneMatch) == "*" {
		return []string{"*"}
	}

	// Split by comma and clean up each ETag
	parts := strings.Split(ifNoneMatch, ",")
	etags := make([]string, 0, len(parts))

	for _, part := range parts {
		etag := strings.TrimSpace(part)
		if etag != "" {
			// Normalize ETag format
			etag = s.normalizeETag(etag)
			etags = append(etags, etag)
		}
	}

	return etags
}

// normalizeETag ensures ETag is properly quoted
func (s *ETagService) normalizeETag(etag string) string {
	if etag == "" {
		return ""
	}

	// Remove existing quotes
	etag = strings.Trim(etag, `"`)

	// Add quotes back
	return `"` + etag + `"`
}

// CompareETags compares two ETags for equality
func (s *ETagService) CompareETags(etag1, etag2 string) bool {
	// Handle empty ETags
	if etag1 == "" || etag2 == "" {
		return false
	}

	// Normalize ETags (ensure they are quoted)
	etag1 = s.normalizeETag(etag1)
	etag2 = s.normalizeETag(etag2)

	return etag1 == etag2
}

// GetETagInfo extracts ETag information from a catalog item
func (s *ETagService) GetETagInfo(item *catalogModels.CatalogItem) (etag string, version int64, lastModified time.Time) {
	if item == nil {
		return "", 0, time.Time{}
	}

	etag = s.GenerateETag(item)
	version = item.Version
	lastModified = item.UpdatedAt

	return etag, version, lastModified
}

// LogETagOperation logs ETag-related operations for debugging
func (s *ETagService) LogETagOperation(c *gin.Context, operation string, item *catalogModels.CatalogItem, details map[string]interface{}) {
	fields := logrus.Fields{
		"operation": operation,
		"method":    c.Request.Method,
		"path":      c.Request.URL.Path,
	}

	if item != nil {
		fields["item_id"] = item.ID
		fields["item_type"] = item.ItemType
		fields["item_version"] = item.Version

		etag := s.GenerateETag(item)
		fields["generated_etag"] = etag
	}

	// Add conditional request headers
	if ifNoneMatch := c.GetHeader("If-None-Match"); ifNoneMatch != "" {
		fields["if_none_match"] = ifNoneMatch
	}

	// Add any additional details
	for key, value := range details {
		fields[key] = value
	}

	s.logger.WithFields(fields).Debug("ETag operation")
}

// ValidateConditionalRequest validates conditional request headers
func (s *ETagService) ValidateConditionalRequest(c *gin.Context) error {
	// Validate If-None-Match header format if present
	ifNoneMatch := c.GetHeader("If-None-Match")
	if ifNoneMatch != "" {
		if err := s.ValidateETagFormat(ifNoneMatch); err != nil {
			s.logger.WithFields(logrus.Fields{
				"if_none_match": ifNoneMatch,
				"error":         err.Error(),
			}).Warn("Invalid If-None-Match header format")
			return err
		}
	}

	return nil
}
