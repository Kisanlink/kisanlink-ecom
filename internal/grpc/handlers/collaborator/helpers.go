package collaborator

import (
	"context"

	"kisanlink-ecom/internal/aaa"
	pb "kisanlink-ecom/proto/gen/go/collaborator/v1"

	"gorm.io/gorm"
)

const (
	maskString = "****"
)

// Removed unused helper functions that are placeholders for future implementation

// convertAddressToProto converts AAA address to proto message
//
//nolint:unused // Reserved for future AAA address expansion feature
func convertAddressToProto(addr *aaa.Address) *pb.Address {
	if addr == nil {
		return nil
	}

	pbAddr := &pb.Address{
		Line1:      addr.Line1,
		City:       addr.City,
		State:      addr.State,
		Country:    addr.Country,
		PostalCode: addr.PostalCode,
	}

	if addr.Line2 != "" {
		pbAddr.Line2 = &addr.Line2
	}

	// Map address type
	pbAddr.Type = mapAAAAddressTypeToPBAddressType(addr.Type)

	return pbAddr
}

// mapAAAAddressTypeToPBAddressType maps AAA address type to proto address type
//
//nolint:unused // Reserved for future AAA address expansion feature
func mapAAAAddressTypeToPBAddressType(aaaType aaa.AddressType) pb.AddressType {
	switch aaaType {
	case aaa.AddressTypeHome:
		return pb.AddressType_ADDRESS_TYPE_HOME
	case aaa.AddressTypeBusiness:
		return pb.AddressType_ADDRESS_TYPE_BUSINESS
	case aaa.AddressTypeBilling:
		return pb.AddressType_ADDRESS_TYPE_BILLING
	case aaa.AddressTypeShipping:
		return pb.AddressType_ADDRESS_TYPE_SHIPPING
	case aaa.AddressTypeWarehouse:
		return pb.AddressType_ADDRESS_TYPE_WAREHOUSE
	default:
		return pb.AddressType_ADDRESS_TYPE_UNSPECIFIED
	}
}

// isCollaboratorAccessibleByUser checks if user can access the collaborator
// Implements FPO and role-based access control
//
//nolint:unused // Reserved for future FPO access control feature
func isCollaboratorAccessibleByUser(
	ctx context.Context,
	db *gorm.DB,
	collaboratorID uint64,
	userID string,
	_ uint64, // fpoID - reserved for future use
	roles []string,
) bool {
	// Admin can access all collaborators
	if hasRole(roles, "ADMIN") {
		return true
	}

	// Check if user owns the collaborator
	var count int64
	err := db.WithContext(ctx).
		Table("collaborators").
		Where("id = ? AND user_id = ?", collaboratorID, userID).
		Count(&count).Error

	if err != nil || count == 0 {
		return false
	}

	// Additional FPO checks can be added here when fpoID field is added to model

	return true
}

// sanitizeSensitiveFields removes or masks sensitive information based on user roles
func sanitizeSensitiveFields(pbCollab *pb.Collaborator, userRoles []string) {
	// If user is not admin, finance manager, or manager, mask sensitive fields
	if !hasRole(userRoles, "ADMIN", "FINANCE_MANAGER", "MANAGER") {
		// Mask tax information in business info
		if pbCollab.BusinessInfo != nil && pbCollab.BusinessInfo.GstNumber != nil {
			masked := maskGST(*pbCollab.BusinessInfo.GstNumber)
			pbCollab.BusinessInfo.GstNumber = &masked
		}

		// Note: Banking fields are in BusinessInfo, not in Collaborator
		// They can be masked here if needed when proto is updated
	}
}

// maskGST masks GST number showing only first 2 and last 3 characters
func maskGST(gst string) string {
	if gst == "" {
		return ""
	}
	if len(gst) <= 5 {
		return maskString
	}
	return gst[:2] + maskString + gst[len(gst)-3:]
}

// validateUpdateRequest validates the update collaborator request
//
//nolint:unused // Reserved for future validation enhancement
func validateUpdateRequest(req *pb.UpdateCollaboratorRequest) error {
	// Basic validation
	if req.Id == "" {
		return gorm.ErrRecordNotFound
	}

	// Add more validation as needed
	return nil
}

// Placeholder for future enhancement - enrichCollaboratorWithMetrics
