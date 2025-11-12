package gst

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// GST number format: GGPPPPPPPPPCZD
// GG = State Code (01-37)
// PPPPPPPPP = PAN (5 letters + 4 digits + 1 letter)
// C = Entity Number (1-9, A-Z)
// Z = Default 'Z'
// D = Check digit (0-9, A-Z)

const (
	// GSTLength is the standard length of Indian GST numbers
	GSTLength = 15
	// StateCodeMinValue is the minimum valid state code
	StateCodeMinValue = 1
	// StateCodeMaxValue is the maximum valid state code
	StateCodeMaxValue = 37
)

var (
	// gstPattern validates the overall format of GST number
	gstPattern = regexp.MustCompile(`^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$`)

	// panPattern validates the PAN portion (positions 2-11)
	panPattern = regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]{1}$`)

	// Valid state codes (01-37 covering all Indian states and union territories)
	validStateCodes = map[string]bool{
		"01": true, // Jammu and Kashmir
		"02": true, // Himachal Pradesh
		"03": true, // Punjab
		"04": true, // Chandigarh
		"05": true, // Uttarakhand
		"06": true, // Haryana
		"07": true, // Delhi
		"08": true, // Rajasthan
		"09": true, // Uttar Pradesh
		"10": true, // Bihar
		"11": true, // Sikkim
		"12": true, // Arunachal Pradesh
		"13": true, // Nagaland
		"14": true, // Manipur
		"15": true, // Mizoram
		"16": true, // Tripura
		"17": true, // Meghalaya
		"18": true, // Assam
		"19": true, // West Bengal
		"20": true, // Jharkhand
		"21": true, // Odisha
		"22": true, // Chhattisgarh
		"23": true, // Madhya Pradesh
		"24": true, // Gujarat
		"25": true, // Daman and Diu
		"26": true, // Dadra and Nagar Haveli
		"27": true, // Maharashtra
		"28": true, // Andhra Pradesh (Old)
		"29": true, // Karnataka
		"30": true, // Goa
		"31": true, // Lakshadweep
		"32": true, // Kerala
		"33": true, // Tamil Nadu
		"34": true, // Puducherry
		"35": true, // Andaman and Nicobar Islands
		"36": true, // Telangana
		"37": true, // Andhra Pradesh (New)
	}

	// ErrInvalidGSTFormat indicates GST format is invalid
	ErrInvalidGSTFormat = fmt.Errorf("invalid GST number format")

	// ErrInvalidStateCode indicates state code is invalid
	ErrInvalidStateCode = fmt.Errorf("invalid state code")

	// ErrInvalidPAN indicates PAN portion is invalid
	ErrInvalidPAN = fmt.Errorf("invalid PAN in GST number")

	// ErrInvalidEntityNumber indicates entity number is invalid
	ErrInvalidEntityNumber = fmt.Errorf("invalid entity number")

	// ErrInvalidChecksum indicates checksum validation failed
	ErrInvalidChecksum = fmt.Errorf("invalid GST checksum")
)

// Validator provides server-side validation for GST numbers
type Validator struct{}

// NewValidator creates a new GST validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate performs comprehensive validation of GST number
func (v *Validator) Validate(gst string) error {
	// Normalize: trim and uppercase
	normalized := v.Normalize(gst)

	// Check length
	if len(normalized) != GSTLength {
		return fmt.Errorf("%w: length must be %d characters", ErrInvalidGSTFormat, GSTLength)
	}

	// Check overall format
	if !gstPattern.MatchString(normalized) {
		return fmt.Errorf("%w: must match pattern GGPPPPPPPPPCZD", ErrInvalidGSTFormat)
	}

	// Validate state code (first 2 digits)
	if err := v.ValidateStateCode(normalized[0:2]); err != nil {
		return err
	}

	// Validate PAN portion (positions 2-11)
	if err := v.ValidatePAN(normalized[2:12]); err != nil {
		return err
	}

	// Validate entity number (position 12)
	if err := v.ValidateEntityNumber(normalized[12:13]); err != nil {
		return err
	}

	// Validate default 'Z' at position 13
	if normalized[13] != 'Z' {
		return fmt.Errorf("%w: position 14 must be 'Z'", ErrInvalidGSTFormat)
	}

	// Validate checksum (position 14)
	if err := v.ValidateChecksum(normalized); err != nil {
		return err
	}

	return nil
}

// Normalize converts GST to uppercase and trims whitespace
func (v *Validator) Normalize(gst string) string {
	return strings.ToUpper(strings.TrimSpace(gst))
}

// ValidateStateCode validates the state code portion (first 2 digits)
func (v *Validator) ValidateStateCode(stateCode string) error {
	if len(stateCode) != 2 {
		return fmt.Errorf("%w: state code must be 2 digits", ErrInvalidStateCode)
	}

	// Check if it's a valid number
	code, err := strconv.Atoi(stateCode)
	if err != nil {
		return fmt.Errorf("%w: state code must be numeric", ErrInvalidStateCode)
	}

	// Check range
	if code < StateCodeMinValue || code > StateCodeMaxValue {
		return fmt.Errorf("%w: state code must be between %02d and %02d", ErrInvalidStateCode, StateCodeMinValue, StateCodeMaxValue)
	}

	// Check against valid state codes
	if !validStateCodes[stateCode] {
		return fmt.Errorf("%w: state code %s is not assigned", ErrInvalidStateCode, stateCode)
	}

	return nil
}

// ValidatePAN validates the PAN portion (5 letters + 4 digits + 1 letter)
func (v *Validator) ValidatePAN(pan string) error {
	if len(pan) != 10 {
		return fmt.Errorf("%w: PAN must be 10 characters", ErrInvalidPAN)
	}

	if !panPattern.MatchString(pan) {
		return fmt.Errorf("%w: PAN must match format LLLLLNNNNL (L=letter, N=digit)", ErrInvalidPAN)
	}

	// Additional PAN validation rules
	// 4th character indicates entity type
	entityType := pan[3]
	validEntityTypes := "CFHAPLJTBGLS" // Company, Firm, HUF, etc.
	if !strings.ContainsRune(validEntityTypes, rune(entityType)) {
		return fmt.Errorf("%w: invalid entity type in PAN", ErrInvalidPAN)
	}

	return nil
}

// ValidateEntityNumber validates the entity number (position 12)
// Valid values: 1-9, A-Z (indicating registration sequence)
func (v *Validator) ValidateEntityNumber(entityNum string) error {
	if len(entityNum) != 1 {
		return fmt.Errorf("%w: entity number must be 1 character", ErrInvalidEntityNumber)
	}

	char := entityNum[0]
	// Valid: 1-9, A-Z
	if (char < '1' || char > '9') && (char < 'A' || char > 'Z') {
		return fmt.Errorf("%w: must be 1-9 or A-Z", ErrInvalidEntityNumber)
	}

	return nil
}

// ValidateChecksum validates the check digit using modified Luhn algorithm
// GST uses a variant of Luhn algorithm for checksum validation
func (v *Validator) ValidateChecksum(gst string) error {
	if len(gst) != GSTLength {
		return ErrInvalidChecksum
	}

	// Calculate checksum for first 14 characters
	expectedCheckDigit := v.CalculateChecksum(gst[:14])
	actualCheckDigit := string(gst[14])

	if expectedCheckDigit != actualCheckDigit {
		return fmt.Errorf("%w: expected %s, got %s", ErrInvalidChecksum, expectedCheckDigit, actualCheckDigit)
	}

	return nil
}

// CalculateChecksum calculates the check digit for GST number
// Uses GST-specific checksum algorithm (variant of Luhn)
func (v *Validator) CalculateChecksum(gstWithoutCheckDigit string) string {
	// Character to value mapping
	charValue := func(c byte) int {
		if c >= '0' && c <= '9' {
			return int(c - '0')
		}
		if c >= 'A' && c <= 'Z' {
			return int(c-'A') + 10
		}
		return 0
	}

	// Calculate weighted sum
	factor := 2
	sum := 0

	// Process from right to left
	for i := len(gstWithoutCheckDigit) - 1; i >= 0; i-- {
		value := charValue(gstWithoutCheckDigit[i])
		product := value * factor

		// Add digits of product to sum
		sum += product / 36
		sum += product % 36

		// Alternate factor between 2 and 1
		if factor == 2 {
			factor = 1
		} else {
			factor = 2
		}
	}

	// Calculate check digit
	checkValue := (36 - (sum % 36)) % 36

	// Convert value to check digit character
	if checkValue < 10 {
		return string(byte('0' + checkValue))
	}
	return string(byte('A' + checkValue - 10))
}

// IsValid is a convenience method that returns bool instead of error
func (v *Validator) IsValid(gst string) bool {
	return v.Validate(gst) == nil
}

// ExtractStateCode extracts the state code from GST number
func (v *Validator) ExtractStateCode(gst string) (string, error) {
	normalized := v.Normalize(gst)
	if len(normalized) < 2 {
		return "", ErrInvalidGSTFormat
	}
	return normalized[0:2], nil
}

// ExtractPAN extracts the PAN from GST number
func (v *Validator) ExtractPAN(gst string) (string, error) {
	normalized := v.Normalize(gst)
	if len(normalized) < 12 {
		return "", ErrInvalidGSTFormat
	}
	return normalized[2:12], nil
}

// GetStateName returns the state name for a given state code
func (v *Validator) GetStateName(stateCode string) string {
	stateNames := map[string]string{
		"01": "Jammu and Kashmir",
		"02": "Himachal Pradesh",
		"03": "Punjab",
		"04": "Chandigarh",
		"05": "Uttarakhand",
		"06": "Haryana",
		"07": "Delhi",
		"08": "Rajasthan",
		"09": "Uttar Pradesh",
		"10": "Bihar",
		"11": "Sikkim",
		"12": "Arunachal Pradesh",
		"13": "Nagaland",
		"14": "Manipur",
		"15": "Mizoram",
		"16": "Tripura",
		"17": "Meghalaya",
		"18": "Assam",
		"19": "West Bengal",
		"20": "Jharkhand",
		"21": "Odisha",
		"22": "Chhattisgarh",
		"23": "Madhya Pradesh",
		"24": "Gujarat",
		"25": "Daman and Diu",
		"26": "Dadra and Nagar Haveli",
		"27": "Maharashtra",
		"28": "Andhra Pradesh (Old)",
		"29": "Karnataka",
		"30": "Goa",
		"31": "Lakshadweep",
		"32": "Kerala",
		"33": "Tamil Nadu",
		"34": "Puducherry",
		"35": "Andaman and Nicobar Islands",
		"36": "Telangana",
		"37": "Andhra Pradesh (New)",
	}
	return stateNames[stateCode]
}
