package utils

import (
	"html"
	"regexp"
	"strings"
	"unicode"

	"github.com/microcosm-cc/bluemonday"
)

// SanitizerConfig holds configuration for input sanitization
type SanitizerConfig struct {
	// StrictMode removes all HTML tags and scripts
	StrictMode bool
	// AllowBasicHTML allows basic HTML tags like <b>, <i>, <em>
	AllowBasicHTML bool
	// MaxLength sets maximum length for strings (0 = no limit)
	MaxLength int
	// RemoveNewlines removes all newline characters
	RemoveNewlines bool
	// NormalizeSpaces converts multiple spaces to single space
	NormalizeSpaces bool
}

// DefaultSanitizerConfig returns default sanitizer configuration
func DefaultSanitizerConfig() *SanitizerConfig {
	return &SanitizerConfig{
		StrictMode:      true,
		AllowBasicHTML:  false,
		MaxLength:       10000,
		RemoveNewlines:  false,
		NormalizeSpaces: true,
	}
}

// Sanitizer provides comprehensive input sanitization
type Sanitizer struct {
	config        *SanitizerConfig
	htmlPolicy    *bluemonday.Policy
	xssPattern    *regexp.Regexp
	sqlPattern    *regexp.Regexp
	scriptPattern *regexp.Regexp
}

// NewSanitizer creates a new sanitizer instance
func NewSanitizer(config *SanitizerConfig) *Sanitizer {
	if config == nil {
		config = DefaultSanitizerConfig()
	}

	var htmlPolicy *bluemonday.Policy
	if config.StrictMode {
		htmlPolicy = bluemonday.StrictPolicy()
	} else if config.AllowBasicHTML {
		htmlPolicy = bluemonday.UGCPolicy()
		// Allow only basic formatting tags
		htmlPolicy = htmlPolicy.AllowElements("b", "i", "em", "strong", "u")
	} else {
		htmlPolicy = bluemonday.NewPolicy()
	}

	// Compile security patterns
	xssPattern := regexp.MustCompile(`(?i)(<script[\s\S]*?</script>|javascript:|vbscript:|on\w+\s*=|<iframe|<object|<embed|<applet|<meta)`)
	sqlPattern := regexp.MustCompile(`(?i)(\b(union|select|insert|update|delete|drop|create|alter|exec|execute|declare|cast|char|concat|substr|ascii|benchmark|sleep|waitfor|sys|xp_)\b|--|;|/\*|\*/|'|\"|xor|or\s+1=1|and\s+1=1)`)
	scriptPattern := regexp.MustCompile(`(?i)(<script[\s\S]*?</script>|<\s*script[^>]*>[\s\S]*?<\s*/\s*script\s*>)`)

	return &Sanitizer{
		config:        config,
		htmlPolicy:    htmlPolicy,
		xssPattern:    xssPattern,
		sqlPattern:    sqlPattern,
		scriptPattern: scriptPattern,
	}
}

// SanitizeString sanitizes a single string input
func (s *Sanitizer) SanitizeString(input string) string {
	if input == "" {
		return input
	}

	// Remove null bytes
	sanitized := strings.ReplaceAll(input, "\x00", "")

	// Remove dangerous script tags first
	sanitized = s.scriptPattern.ReplaceAllString(sanitized, "")

	// HTML sanitization
	sanitized = s.htmlPolicy.Sanitize(sanitized)

	// Additional XSS protection
	sanitized = s.removeXSSPatterns(sanitized)

	// Remove SQL injection patterns
	sanitized = s.removeSQLPatterns(sanitized)

	// HTML escape remaining content
	sanitized = html.EscapeString(sanitized)

	// Normalize whitespace
	if s.config.NormalizeSpaces {
		sanitized = s.normalizeSpaces(sanitized)
	}

	// Remove newlines if configured
	if s.config.RemoveNewlines {
		sanitized = strings.ReplaceAll(sanitized, "\n", " ")
		sanitized = strings.ReplaceAll(sanitized, "\r", " ")
	}

	// Trim length if configured
	if s.config.MaxLength > 0 && len(sanitized) > s.config.MaxLength {
		sanitized = sanitized[:s.config.MaxLength]
	}

	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

// SanitizeHTML sanitizes HTML content while preserving safe markup
func (s *Sanitizer) SanitizeHTML(input string) string {
	if input == "" {
		return input
	}

	// Use HTML policy for sanitization
	sanitized := s.htmlPolicy.Sanitize(input)

	// Additional security checks
	sanitized = s.removeXSSPatterns(sanitized)
	sanitized = s.removeDangerousAttributes(sanitized)

	return sanitized
}

// SanitizeEmail sanitizes and validates email addresses
func (s *Sanitizer) SanitizeEmail(email string) string {
	if email == "" {
		return email
	}

	// Basic sanitization
	sanitized := strings.TrimSpace(strings.ToLower(email))

	// Remove dangerous characters
	sanitized = s.removeDangerousChars(sanitized)

	// Validate email format
	if !IsValidEmail(sanitized) {
		return ""
	}

	return sanitized
}

// SanitizeUsername sanitizes username input
func (s *Sanitizer) SanitizeUsername(username string) string {
	if username == "" {
		return username
	}

	// Remove whitespace and convert to lowercase
	sanitized := strings.TrimSpace(strings.ToLower(username))

	// Remove dangerous characters
	sanitized = regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(sanitized, "")

	// Limit length
	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}

	return sanitized
}

// SanitizeFilename sanitizes filename input
func (s *Sanitizer) SanitizeFilename(filename string) string {
	if filename == "" {
		return filename
	}

	// Remove path traversal patterns
	sanitized := strings.ReplaceAll(filename, "../", "")
	sanitized = strings.ReplaceAll(sanitized, "..\\", "")
	sanitized = strings.ReplaceAll(sanitized, "...", "")

	// Remove dangerous characters
	dangerousChars := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	sanitized = dangerousChars.ReplaceAllString(sanitized, "_")

	// Remove leading/trailing dots and spaces
	sanitized = strings.Trim(sanitized, ". ")

	// Limit length
	if len(sanitized) > 255 {
		sanitized = sanitized[:255]
	}

	return sanitized
}

// SanitizeURL sanitizes URL input
func (s *Sanitizer) SanitizeURL(url string) string {
	if url == "" {
		return url
	}

	// Remove whitespace
	sanitized := strings.TrimSpace(url)

	// Check for dangerous protocols
	dangerousProtocols := []string{"javascript:", "vbscript:", "data:", "file:"}
	lowerURL := strings.ToLower(sanitized)
	for _, protocol := range dangerousProtocols {
		if strings.HasPrefix(lowerURL, protocol) {
			return ""
		}
	}

	// Remove XSS patterns
	sanitized = s.removeXSSPatterns(sanitized)

	return sanitized
}

// SanitizePhoneNumber sanitizes phone number input
func (s *Sanitizer) SanitizePhoneNumber(phone string) string {
	if phone == "" {
		return phone
	}

	// Remove all non-digit characters except + and -
	sanitized := regexp.MustCompile(`[^\d+\-\s()]`).ReplaceAllString(phone, "")

	// Remove extra spaces
	sanitized = regexp.MustCompile(`\s+`).ReplaceAllString(sanitized, " ")
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

// SanitizeSearchQuery sanitizes search query input
func (s *Sanitizer) SanitizeSearchQuery(query string) string {
	if query == "" {
		return query
	}

	// Basic sanitization
	sanitized := strings.TrimSpace(query)

	// Remove SQL injection patterns
	sanitized = s.removeSQLPatterns(sanitized)

	// Remove XSS patterns
	sanitized = s.removeXSSPatterns(sanitized)

	// Remove special regex characters that could cause issues
	specialChars := regexp.MustCompile(`[.*+?^${}()|[\]\\]`)
	sanitized = specialChars.ReplaceAllString(sanitized, " ")

	// Normalize spaces
	sanitized = s.normalizeSpaces(sanitized)

	// Limit length
	if len(sanitized) > 100 {
		sanitized = sanitized[:100]
	}

	return sanitized
}

// RemoveNonPrintable removes non-printable characters
func (s *Sanitizer) RemoveNonPrintable(input string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) || unicode.IsSpace(r) {
			return r
		}
		return -1
	}, input)
}

// StripHTML completely removes all HTML tags
func (s *Sanitizer) StripHTML(input string) string {
	// Remove all HTML tags
	htmlTag := regexp.MustCompile(`<[^>]*>`)
	stripped := htmlTag.ReplaceAllString(input, "")

	// Remove HTML entities
	stripped = html.UnescapeString(stripped)

	// Remove any remaining dangerous patterns
	stripped = s.removeXSSPatterns(stripped)

	return stripped
}

// Helper methods

func (s *Sanitizer) removeXSSPatterns(input string) string {
	// Remove common XSS patterns
	xssPatterns := []string{
		`javascript:`,
		`vbscript:`,
		`onload=`,
		`onerror=`,
		`onclick=`,
		`onmouseover=`,
		`onfocus=`,
		`onblur=`,
		`eval\(`,
		`expression\(`,
		`<script`,
		`</script>`,
		`<iframe`,
		`<object`,
		`<embed`,
		`<applet`,
		`<meta`,
	}

	result := input
	for _, pattern := range xssPatterns {
		re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(pattern))
		result = re.ReplaceAllString(result, "")
	}

	return result
}

func (s *Sanitizer) removeSQLPatterns(input string) string {
	// Remove common SQL injection patterns
	sqlPatterns := []string{
		`union\s+select`,
		`;\s*drop\s+table`,
		`;\s*delete\s+from`,
		`;\s*insert\s+into`,
		`;\s*update\s+set`,
		`;\s*create\s+table`,
		`;\s*alter\s+table`,
		`exec\s*\(`,
		`execute\s*\(`,
		`xp_cmdshell`,
		`sp_executesql`,
	}

	result := input
	for _, pattern := range sqlPatterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		result = re.ReplaceAllString(result, "")
	}

	return result
}

func (s *Sanitizer) removeDangerousAttributes(input string) string {
	// Remove dangerous HTML attributes
	dangerousAttrs := []string{
		`on\w+\s*=\s*["'][^"']*["']`,
		`style\s*=\s*["'][^"']*expression[^"']*["']`,
		`href\s*=\s*["']javascript:[^"']*["']`,
		`src\s*=\s*["']javascript:[^"']*["']`,
	}

	result := input
	for _, pattern := range dangerousAttrs {
		re := regexp.MustCompile(`(?i)` + pattern)
		result = re.ReplaceAllString(result, "")
	}

	return result
}

func (s *Sanitizer) removeDangerousChars(input string) string {
	// Remove null bytes and control characters
	dangerous := regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	return dangerous.ReplaceAllString(input, "")
}

func (s *Sanitizer) normalizeSpaces(input string) string {
	// Replace multiple spaces with single space
	spaces := regexp.MustCompile(`\s+`)
	return spaces.ReplaceAllString(input, " ")
}

// Validation helper methods

// IsSafeString checks if a string contains only safe characters
func IsSafeString(input string) bool {
	// Allow alphanumeric, spaces, and basic punctuation
	safePattern := regexp.MustCompile(`^[a-zA-Z0-9\s\-_.,!?()@#$%^&*+=\[\]{}|\\:";'<>?/~` + "`" + `]+$`)
	return safePattern.MatchString(input)
}

// ContainsHTML checks if a string contains HTML tags
func ContainsHTML(input string) bool {
	htmlPattern := regexp.MustCompile(`<[^>]*>`)
	return htmlPattern.MatchString(input)
}

// ContainsScript checks if a string contains script tags
func ContainsScript(input string) bool {
	scriptPattern := regexp.MustCompile(`(?i)<script[\s\S]*?</script>`)
	return scriptPattern.MatchString(input)
}

// ContainsSQLKeywords checks if a string contains SQL keywords
func ContainsSQLKeywords(input string) bool {
	sqlKeywords := []string{
		"select", "insert", "update", "delete", "drop", "create", "alter",
		"union", "exec", "execute", "declare", "cast", "char", "concat",
	}

	lowerInput := strings.ToLower(input)
	for _, keyword := range sqlKeywords {
		if strings.Contains(lowerInput, keyword) {
			return true
		}
	}
	return false
}

// Global sanitizer instance for convenience
var defaultSanitizer *Sanitizer

func init() {
	defaultSanitizer = NewSanitizer(DefaultSanitizerConfig())
}

// Convenience functions using default sanitizer

// SanitizeString sanitizes a string using default configuration
func SanitizeString(input string) string {
	return defaultSanitizer.SanitizeString(input)
}

// SanitizeEmail sanitizes an email using default configuration
func SanitizeEmail(email string) string {
	return defaultSanitizer.SanitizeEmail(email)
}

// SanitizeUsername sanitizes a username using default configuration
func SanitizeUsername(username string) string {
	return defaultSanitizer.SanitizeUsername(username)
}

// SanitizeFilename sanitizes a filename using default configuration
func SanitizeFilename(filename string) string {
	return defaultSanitizer.SanitizeFilename(filename)
}

// SanitizeURL sanitizes a URL using default configuration
func SanitizeURL(url string) string {
	return defaultSanitizer.SanitizeURL(url)
}

// SanitizeSearchQuery sanitizes a search query using default configuration
func SanitizeSearchQuery(query string) string {
	return defaultSanitizer.SanitizeSearchQuery(query)
}
