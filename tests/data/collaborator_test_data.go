package data

import (
	"kisanlink-ecom/entities/models/collaborator"
	collaboratorRequests "kisanlink-ecom/entities/requests/collaborator"
)

// TestCollaboratorData provides test data for collaborator tests
var TestCollaboratorData = struct {
	ValidCreateRequest   collaboratorRequests.CreateCollaboratorRequest
	ValidUpdateRequest   collaboratorRequests.UpdateCollaboratorRequest
	ValidCollaborator    *collaborator.Collaborator
	InvalidCreateRequest collaboratorRequests.CreateCollaboratorRequest
}{
	ValidCreateRequest: collaboratorRequests.CreateCollaboratorRequest{
		UserID:              "test-user-123",
		Username:            "testfarmer",
		Email:               "farmer@test.com",
		FirstName:           "John",
		LastName:            "Doe",
		Phone:               "+1234567890",
		CollaboratorType:    collaborator.CollaboratorTypeFarmer,
		OrganizationID:      "org-123",
		Bio:                 "Experienced farmer specializing in organic crops",
		Location:            "Rural Valley, State",
		Coordinates:         "40.7128,-74.0060",
		BusinessName:        "Doe Organic Farms",
		BusinessType:        "Agriculture",
		BusinessLicense:     "AGR-2024-001",
		TaxID:               "TAX123456789",
		BusinessAddress:     "123 Farm Road, Rural Valley, State 12345",
		BusinessPhone:       "+1234567891",
		BusinessEmail:       "business@doefarms.com",
		BusinessWebsite:     "https://doefarms.com",
		BusinessDescription: "Family-owned organic farm producing vegetables and fruits",
		LanguagePreference:  "en",
		TimezonePreference:  "America/New_York",
		InvitedBy:           "admin-user-456",
	},
	ValidUpdateRequest: collaboratorRequests.UpdateCollaboratorRequest{
		FirstName: collabStringPtr("Jane"),
		LastName:  collabStringPtr("Smith"),
		Bio:       collabStringPtr("Updated bio for testing"),
		Location:  collabStringPtr("New Location"),
	},
	ValidCollaborator: collaborator.NewCollaborator(
		"test-user-456",
		"testuser",
		"test@example.com",
		"Test",
		"User",
		collaborator.CollaboratorTypeSupplier,
	),
	InvalidCreateRequest: collaboratorRequests.CreateCollaboratorRequest{
		// Missing required fields
		UserID:   "",
		Username: "",
		Email:    "invalid-email",
	},
}

// Helper function to create string pointers for collaborator tests
func collabStringPtr(s string) *string {
	return &s
}

// TestCollaboratorFilters provides test filter data
var TestCollaboratorFilters = struct {
	ByType         collaboratorRequests.CollaboratorFilter
	ByStatus       collaboratorRequests.CollaboratorFilter
	ByVerification collaboratorRequests.CollaboratorFilter
	BySearch       collaboratorRequests.CollaboratorFilter
}{
	ByType: collaboratorRequests.CollaboratorFilter{
		CollaboratorType: func() *collaborator.CollaboratorType { t := collaborator.CollaboratorTypeFarmer; return &t }(),
	},
	ByStatus: collaboratorRequests.CollaboratorFilter{
		Status: func() *collaborator.CollaboratorStatus { s := collaborator.CollaboratorStatusActive; return &s }(),
	},
	ByVerification: collaboratorRequests.CollaboratorFilter{
		IsVerified: collabBoolPtr(true),
	},
	BySearch: collaboratorRequests.CollaboratorFilter{
		Search: collabStringPtr("John"),
	},
}

// Helper function to create bool pointers for collaborator tests
func collabBoolPtr(b bool) *bool {
	return &b
}
