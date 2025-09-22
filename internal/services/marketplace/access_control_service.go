package marketplace

import (
	"context"
	"fmt"
	"time"

	"kisanlink-ecom/entities/models/marketplace"
	"kisanlink-ecom/internal/common"
	marketplaceRepo "kisanlink-ecom/internal/repositories/marketplace"
)

// AccessControlServiceInterface defines the interface for visibility and access control operations
type AccessControlServiceInterface interface {
	// Access Validation
	CanViewListing(ctx context.Context, listing *marketplace.Listing, userID string, orgID string) bool
	CanBidOnListing(ctx context.Context, listingID string, userID string, orgID string) (bool, error)
	CanEditListing(ctx context.Context, listingID string, userID string, orgID string) (bool, error)

	// Invitation Management for Private Auctions
	InviteParticipant(ctx context.Context, req *InviteParticipantRequest) error
	RemoveParticipant(ctx context.Context, req *RemoveParticipantRequest) error
	GetInvitedParticipants(ctx context.Context, listingID string, pagination *common.PaginationParams) ([]*AuctionParticipant, int, error)
	IsParticipantInvited(ctx context.Context, listingID string, userID string) (bool, error)

	// Network and Organization Membership Validation
	ValidateNetworkMembership(ctx context.Context, userOrgID string, listingOrgID string) (bool, error)
	ValidateOrganizationMembership(ctx context.Context, userOrgID string, listingOrgID string) (bool, error)

	// Visibility Configuration
	ValidateVisibilityConfiguration(ctx context.Context, listing *marketplace.Listing) error
	GetVisibilityRules(ctx context.Context, listingID string) (*VisibilityRules, error)

	// Access Audit
	LogAccessAttempt(ctx context.Context, req *AccessAttemptLog) error
	GetAccessLogs(ctx context.Context, listingID string, pagination *common.PaginationParams) ([]*AccessAttemptLog, int, error)
}

// InviteParticipantRequest represents a request to invite a participant to a private auction
type InviteParticipantRequest struct {
	ListingID     string     `json:"listing_id" validate:"required"`
	InviterID     string     `json:"inviter_id" validate:"required"`
	ParticipantID string     `json:"participant_id" validate:"required"`
	Message       string     `json:"message,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
}

// RemoveParticipantRequest represents a request to remove a participant from a private auction
type RemoveParticipantRequest struct {
	ListingID     string `json:"listing_id" validate:"required"`
	RemoverID     string `json:"remover_id" validate:"required"`
	ParticipantID string `json:"participant_id" validate:"required"`
	Reason        string `json:"reason,omitempty"`
}

// AuctionParticipant represents a participant in a private auction
type AuctionParticipant struct {
	ID            string     `json:"id"`
	ListingID     string     `json:"listing_id"`
	ParticipantID string     `json:"participant_id"`
	InviterID     string     `json:"inviter_id"`
	Status        string     `json:"status"` // INVITED, ACCEPTED, DECLINED, REMOVED
	InvitedAt     time.Time  `json:"invited_at"`
	RespondedAt   *time.Time `json:"responded_at,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	Message       string     `json:"message,omitempty"`
}

// VisibilityRules represents the visibility and access rules for a listing
type VisibilityRules struct {
	ListingID            string                        `json:"listing_id"`
	Visibility           marketplace.ListingVisibility `json:"visibility"`
	AuctionType          marketplace.AuctionType       `json:"auction_type"`
	BidVisibility        marketplace.BidVisibility     `json:"bid_visibility"`
	RequiresInvitation   bool                          `json:"requires_invitation"`
	AllowedOrganizations []string                      `json:"allowed_organizations,omitempty"`
	NetworkRestrictions  []string                      `json:"network_restrictions,omitempty"`
	AccessRules          map[string]interface{}        `json:"access_rules,omitempty"`
}

// AccessAttemptLog represents a log entry for access attempts
type AccessAttemptLog struct {
	ID        string    `json:"id"`
	ListingID string    `json:"listing_id"`
	UserID    string    `json:"user_id"`
	OrgID     string    `json:"org_id"`
	Action    string    `json:"action"` // VIEW, BID, EDIT
	Result    string    `json:"result"` // ALLOWED, DENIED
	Reason    string    `json:"reason,omitempty"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// NetworkServiceInterface defines the interface for network operations
type NetworkServiceInterface interface {
	AreOrganizationsInSameNetwork(ctx context.Context, orgID1, orgID2 string) (bool, error)
	GetNetworkMembers(ctx context.Context, orgID string) ([]string, error)
}

// OrganizationServiceInterface defines the interface for organization operations
type OrganizationServiceInterface interface {
	ValidateOrganizationMembership(ctx context.Context, userID, orgID string) (bool, error)
	GetUserOrganization(ctx context.Context, userID string) (string, error)
}

// AccessControlService provides business logic for visibility and access control
type AccessControlService struct {
	listingRepo    marketplaceRepo.ListingRepository
	eventService   EventServiceInterface
	networkService NetworkServiceInterface
	orgService     OrganizationServiceInterface

	// In-memory storage for invitations (in production, this would be a database)
	invitations map[string][]*AuctionParticipant
	accessLogs  []*AccessAttemptLog
}

// NewAccessControlService creates a new access control service
func NewAccessControlService(
	listingRepo marketplaceRepo.ListingRepository,
	eventService EventServiceInterface,
	networkService NetworkServiceInterface,
	orgService OrganizationServiceInterface,
) AccessControlServiceInterface {
	return &AccessControlService{
		listingRepo:    listingRepo,
		eventService:   eventService,
		networkService: networkService,
		orgService:     orgService,
		invitations:    make(map[string][]*AuctionParticipant),
		accessLogs:     make([]*AccessAttemptLog, 0),
	}
}

// CanViewListing checks if a user can view a specific listing
func (s *AccessControlService) CanViewListing(ctx context.Context, listing *marketplace.Listing, userID string, orgID string) bool {
	if listing == nil {
		return false
	}

	// Check visibility based on listing configuration
	canView, reason := s.checkVisibilityAccess(ctx, listing, userID, orgID)

	// Log access attempt
	_ = s.LogAccessAttempt(ctx, &AccessAttemptLog{
		ListingID: listing.ListingID,
		UserID:    userID,
		OrgID:     orgID,
		Action:    "VIEW",
		Result:    s.boolToResult(canView),
		Reason:    reason,
		Timestamp: time.Now(),
	})

	return canView
}

// CanBidOnListing checks if a user can place bids on a specific listing
func (s *AccessControlService) CanBidOnListing(ctx context.Context, listingID string, userID string, orgID string) (bool, error) {
	// Get the listing
	listing, err := s.listingRepo.GetByListingID(ctx, listingID)
	if err != nil {
		return false, fmt.Errorf("failed to get listing: %w", err)
	}
	if listing == nil {
		return false, fmt.Errorf("listing not found: %s", listingID)
	}

	// First check if user can view the listing
	canView := s.CanViewListing(ctx, listing, userID, orgID)
	if !canView {
		return false, nil
	}

	// Additional bidding checks
	canBid := true
	reason := "Access granted"

	// Check if user is the seller (cannot bid on own listing)
	if listing.SellerID == userID {
		canBid = false
		reason = "Cannot bid on own listing"
	}

	// Check if listing is active and not expired
	if !listing.CanAcceptBids() {
		canBid = false
		reason = "Listing cannot accept bids"
	}

	// Log access attempt
	_ = s.LogAccessAttempt(ctx, &AccessAttemptLog{
		ListingID: listingID,
		UserID:    userID,
		OrgID:     orgID,
		Action:    "BID",
		Result:    s.boolToResult(canBid),
		Reason:    reason,
		Timestamp: time.Now(),
	})

	return canBid, nil
}

// CanEditListing checks if a user can edit a specific listing
func (s *AccessControlService) CanEditListing(ctx context.Context, listingID string, userID string, orgID string) (bool, error) {
	// Get the listing
	listing, err := s.listingRepo.GetByListingID(ctx, listingID)
	if err != nil {
		return false, fmt.Errorf("failed to get listing: %w", err)
	}
	if listing == nil {
		return false, fmt.Errorf("listing not found: %s", listingID)
	}

	canEdit := false
	reason := "Access denied"

	// Check if user is the seller
	if listing.SellerID == userID {
		canEdit = true
		reason = "Listing owner"
	} else if listing.OrganizationID == orgID {
		// Check if user is from the same organization (with appropriate permissions)
		canEdit = true
		reason = "Same organization"
	}

	// Additional checks for listing state
	if canEdit && listing.Status != marketplace.ListingStatusActive {
		canEdit = false
		reason = "Listing is not active"
	}

	// Log access attempt
	_ = s.LogAccessAttempt(ctx, &AccessAttemptLog{
		ListingID: listingID,
		UserID:    userID,
		OrgID:     orgID,
		Action:    "EDIT",
		Result:    s.boolToResult(canEdit),
		Reason:    reason,
		Timestamp: time.Now(),
	})

	return canEdit, nil
}

// InviteParticipant invites a participant to a private auction
func (s *AccessControlService) InviteParticipant(ctx context.Context, req *InviteParticipantRequest) error {
	// Validate request
	if err := s.validateInviteRequest(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Get the listing
	listing, err := s.listingRepo.GetByListingID(ctx, req.ListingID)
	if err != nil {
		return fmt.Errorf("failed to get listing: %w", err)
	}
	if listing == nil {
		return fmt.Errorf("listing not found: %s", req.ListingID)
	}

	// Check if listing is private
	if listing.Visibility != marketplace.VisibilityPrivate {
		return fmt.Errorf("can only invite participants to private auctions")
	}

	// Check if inviter has permission
	canEdit, err := s.CanEditListing(ctx, req.ListingID, req.InviterID, "")
	if err != nil {
		return fmt.Errorf("failed to check inviter permissions: %w", err)
	}
	if !canEdit {
		return fmt.Errorf("inviter does not have permission to manage this auction")
	}

	// Check if participant is already invited
	isInvited, err := s.IsParticipantInvited(ctx, req.ListingID, req.ParticipantID)
	if err != nil {
		return fmt.Errorf("failed to check existing invitation: %w", err)
	}
	if isInvited {
		return fmt.Errorf("participant %s is already invited to auction %s", req.ParticipantID, req.ListingID)
	}

	// Create invitation
	invitation := &AuctionParticipant{
		ID:            fmt.Sprintf("INV_%d", time.Now().UnixNano()),
		ListingID:     req.ListingID,
		ParticipantID: req.ParticipantID,
		InviterID:     req.InviterID,
		Status:        "INVITED",
		InvitedAt:     time.Now(),
		ExpiresAt:     req.ExpiresAt,
		Message:       req.Message,
	}

	// Store invitation
	if s.invitations[req.ListingID] == nil {
		s.invitations[req.ListingID] = make([]*AuctionParticipant, 0)
	}
	s.invitations[req.ListingID] = append(s.invitations[req.ListingID], invitation)

	// Record event
	if s.eventService != nil {
		eventData := map[string]interface{}{
			"listing_id":     req.ListingID,
			"participant_id": req.ParticipantID,
			"inviter_id":     req.InviterID,
			"invitation_id":  invitation.ID,
		}
		_ = s.eventService.RecordListingEvent(ctx, req.ListingID, marketplace.EventBidPlaced, eventData, req.InviterID)
	}

	return nil
}

// RemoveParticipant removes a participant from a private auction
func (s *AccessControlService) RemoveParticipant(ctx context.Context, req *RemoveParticipantRequest) error {
	// Validate request
	if err := s.validateRemoveRequest(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if remover has permission
	canEdit, err := s.CanEditListing(ctx, req.ListingID, req.RemoverID, "")
	if err != nil {
		return fmt.Errorf("failed to check remover permissions: %w", err)
	}
	if !canEdit {
		return fmt.Errorf("remover does not have permission to manage this auction")
	}

	// Find and remove invitation
	invitations := s.invitations[req.ListingID]
	if invitations == nil {
		return fmt.Errorf("no invitations found for listing %s", req.ListingID)
	}

	found := false
	for _, invitation := range invitations {
		if invitation.ParticipantID == req.ParticipantID {
			invitation.Status = "REMOVED"
			now := time.Now()
			invitation.RespondedAt = &now
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("participant %s not found in auction %s", req.ParticipantID, req.ListingID)
	}

	return nil
}

// GetInvitedParticipants retrieves the list of invited participants for a private auction
func (s *AccessControlService) GetInvitedParticipants(ctx context.Context, listingID string, pagination *common.PaginationParams) ([]*AuctionParticipant, int, error) {
	invitations := s.invitations[listingID]
	if invitations == nil {
		return []*AuctionParticipant{}, 0, nil
	}

	// Filter active invitations
	activeInvitations := make([]*AuctionParticipant, 0)
	for _, invitation := range invitations {
		if invitation.Status == "INVITED" || invitation.Status == "ACCEPTED" {
			activeInvitations = append(activeInvitations, invitation)
		}
	}

	// Apply pagination
	total := len(activeInvitations)
	start := pagination.CalculateOffset()
	end := start + pagination.Limit

	if start >= total {
		return []*AuctionParticipant{}, total, nil
	}
	if end > total {
		end = total
	}

	return activeInvitations[start:end], total, nil
}

// IsParticipantInvited checks if a participant is invited to a private auction
func (s *AccessControlService) IsParticipantInvited(ctx context.Context, listingID string, userID string) (bool, error) {
	invitations := s.invitations[listingID]
	if invitations == nil {
		return false, nil
	}

	for _, invitation := range invitations {
		if invitation.ParticipantID == userID &&
			(invitation.Status == "INVITED" || invitation.Status == "ACCEPTED") {
			// Check if invitation has expired
			if invitation.ExpiresAt != nil && time.Now().After(*invitation.ExpiresAt) {
				return false, nil
			}
			return true, nil
		}
	}

	return false, nil
}

// ValidateNetworkMembership validates if two organizations are in the same network
func (s *AccessControlService) ValidateNetworkMembership(ctx context.Context, userOrgID string, listingOrgID string) (bool, error) {
	if s.networkService == nil {
		// If no network service is available, allow access (simplified)
		return true, nil
	}

	return s.networkService.AreOrganizationsInSameNetwork(ctx, userOrgID, listingOrgID)
}

// ValidateOrganizationMembership validates if user belongs to the same organization
func (s *AccessControlService) ValidateOrganizationMembership(ctx context.Context, userOrgID string, listingOrgID string) (bool, error) {
	return userOrgID == listingOrgID, nil
}

// ValidateVisibilityConfiguration validates the visibility configuration of a listing
func (s *AccessControlService) ValidateVisibilityConfiguration(ctx context.Context, listing *marketplace.Listing) error {
	// Validate visibility and auction type combination
	switch listing.Visibility {
	case marketplace.VisibilityPrivate:
		// Private listings require invitation management
		// Additional validation could be added here
	case marketplace.VisibilityPublic:
		// Public listings are always valid
	case marketplace.VisibilityNetwork:
		// Network listings require network service
		if s.networkService == nil {
			return fmt.Errorf("network visibility requires network service configuration")
		}
	case marketplace.VisibilityOrganization:
		// Organization listings are always valid
	default:
		return fmt.Errorf("invalid visibility setting: %s", listing.Visibility)
	}

	// Validate auction type
	switch listing.AuctionType {
	case marketplace.AuctionTypeOpen, marketplace.AuctionTypeClosed:
		// Both are valid
	default:
		return fmt.Errorf("invalid auction type: %s", listing.AuctionType)
	}

	// Validate bid visibility
	switch listing.BidVisibility {
	case marketplace.BidVisibilityFull, marketplace.BidVisibilityPartial,
		marketplace.BidVisibilityMinimal, marketplace.BidVisibilityHidden:
		// All are valid
	default:
		return fmt.Errorf("invalid bid visibility setting: %s", listing.BidVisibility)
	}

	return nil
}

// GetVisibilityRules retrieves the visibility rules for a listing
func (s *AccessControlService) GetVisibilityRules(ctx context.Context, listingID string) (*VisibilityRules, error) {
	// Get the listing
	listing, err := s.listingRepo.GetByListingID(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}
	if listing == nil {
		return nil, fmt.Errorf("listing not found: %s", listingID)
	}

	rules := &VisibilityRules{
		ListingID:          listingID,
		Visibility:         listing.Visibility,
		AuctionType:        listing.AuctionType,
		BidVisibility:      listing.BidVisibility,
		RequiresInvitation: listing.Visibility == marketplace.VisibilityPrivate,
		AccessRules:        make(map[string]interface{}),
	}

	// Add specific rules based on visibility
	switch listing.Visibility {
	case marketplace.VisibilityOrganization:
		rules.AllowedOrganizations = []string{listing.OrganizationID}
	case marketplace.VisibilityNetwork:
		// Would populate network restrictions if network service is available
	}

	return rules, nil
}

// LogAccessAttempt logs an access attempt for audit purposes
func (s *AccessControlService) LogAccessAttempt(ctx context.Context, req *AccessAttemptLog) error {
	// Set ID if not provided
	if req.ID == "" {
		req.ID = fmt.Sprintf("LOG_%d", time.Now().UnixNano())
	}

	// Store log entry (in production, this would be persisted to database)
	s.accessLogs = append(s.accessLogs, req)

	return nil
}

// GetAccessLogs retrieves access logs for a listing
func (s *AccessControlService) GetAccessLogs(ctx context.Context, listingID string, pagination *common.PaginationParams) ([]*AccessAttemptLog, int, error) {
	// Filter logs for the specific listing
	listingLogs := make([]*AccessAttemptLog, 0)
	for _, log := range s.accessLogs {
		if log.ListingID == listingID {
			listingLogs = append(listingLogs, log)
		}
	}

	// Apply pagination
	total := len(listingLogs)
	start := pagination.CalculateOffset()
	end := start + pagination.Limit

	if start >= total {
		return []*AccessAttemptLog{}, total, nil
	}
	if end > total {
		end = total
	}

	return listingLogs[start:end], total, nil
}

// Helper methods

// checkVisibilityAccess checks if a user has access based on visibility settings
func (s *AccessControlService) checkVisibilityAccess(ctx context.Context, listing *marketplace.Listing, userID string, orgID string) (bool, string) {
	switch listing.Visibility {
	case marketplace.VisibilityPrivate:
		// Check if user is invited
		isInvited, err := s.IsParticipantInvited(ctx, listing.ListingID, userID)
		if err != nil {
			return false, "Error checking invitation status"
		}
		if !isInvited {
			return false, "Not invited to private auction"
		}
		return true, "Invited participant"

	case marketplace.VisibilityPublic:
		return true, "Public listing"

	case marketplace.VisibilityNetwork:
		// Check network membership
		isNetworkMember, err := s.ValidateNetworkMembership(ctx, orgID, listing.OrganizationID)
		if err != nil {
			return false, "Error validating network membership"
		}
		if !isNetworkMember {
			return false, "Not in same network"
		}
		return true, "Network member"

	case marketplace.VisibilityOrganization:
		// Check organization membership
		isSameOrg, err := s.ValidateOrganizationMembership(ctx, orgID, listing.OrganizationID)
		if err != nil {
			return false, "Error validating organization membership"
		}
		if !isSameOrg {
			return false, "Not in same organization"
		}
		return true, "Same organization"

	default:
		return false, "Invalid visibility setting"
	}
}

// validateInviteRequest validates an invite participant request
func (s *AccessControlService) validateInviteRequest(req *InviteParticipantRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if req.ListingID == "" {
		return fmt.Errorf("listing ID is required")
	}
	if req.InviterID == "" {
		return fmt.Errorf("inviter ID is required")
	}
	if req.ParticipantID == "" {
		return fmt.Errorf("participant ID is required")
	}
	if req.InviterID == req.ParticipantID {
		return fmt.Errorf("cannot invite yourself")
	}
	return nil
}

// validateRemoveRequest validates a remove participant request
func (s *AccessControlService) validateRemoveRequest(req *RemoveParticipantRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if req.ListingID == "" {
		return fmt.Errorf("listing ID is required")
	}
	if req.RemoverID == "" {
		return fmt.Errorf("remover ID is required")
	}
	if req.ParticipantID == "" {
		return fmt.Errorf("participant ID is required")
	}
	return nil
}

// boolToResult converts a boolean to a result string
func (s *AccessControlService) boolToResult(b bool) string {
	if b {
		return "ALLOWED"
	}
	return "DENIED"
}
