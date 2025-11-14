package data

import (
	"time"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/actors"
)

// CreateTestCollaborator creates a test collaborator with default values
func CreateTestCollaborator() *actors.Collaborator {
	userID := "user-test-id-123"
	return &actors.Collaborator{
		AAAEntityID:           "aaa-entity-id-123",
		AAAEntityType:         actors.AAAEntityTypeUser,
		UserID:                &userID,
		ContextOrganizationID: "org-test-id-123",
		Role:                  actors.CollaboratorRoleEmployee,
		Status:                actors.CollaboratorStatusActive,
		DisplayName:           "Test User",
		Email:                 "test@example.com",
		Phone:                 "+1234567890",
		AccessLevel:           5,
		CanInviteOthers:       false,
		Department:            "Engineering",
		JobTitle:              "Software Engineer",
		InvitedByUserID:       "admin-user-id",
		CreatedBy:             "admin-user-id",
		UpdatedBy:             "admin-user-id",
	}
}

// CreateTestCollaboratorWithID creates a test collaborator with a specific ID
func CreateTestCollaboratorWithID(id string) *actors.Collaborator {
	collab := CreateTestCollaborator()
	collab.ID = id
	return collab
}

// CreateTestCollaboratorWithContextOrg creates a test collaborator with a specific context organization
func CreateTestCollaboratorWithContextOrg(contextOrgID string) *actors.Collaborator {
	collab := CreateTestCollaborator()
	collab.ContextOrganizationID = contextOrgID
	return collab
}

// CreateTestUserCollaborator creates a user-type collaborator
func CreateTestUserCollaborator(userID, contextOrgID string) *actors.Collaborator {
	collab := CreateTestCollaborator()
	collab.UserID = &userID
	collab.AAAEntityID = userID
	collab.AAAEntityType = actors.AAAEntityTypeUser
	collab.ContextOrganizationID = contextOrgID
	return collab
}

// CreateTestOrganizationCollaborator creates an organization-type collaborator
func CreateTestOrganizationCollaborator(orgID, contextOrgID string) *actors.Collaborator {
	collab := CreateTestCollaborator()
	collab.OrganizationID = &orgID
	collab.AAAEntityID = orgID
	collab.AAAEntityType = actors.AAAEntityTypeOrganization
	collab.ContextOrganizationID = contextOrgID
	collab.DisplayName = "Test Organization"
	collab.Email = "org@example.com"
	return collab
}

// CreateTestCollaboratorWithRole creates a collaborator with a specific role
func CreateTestCollaboratorWithRole(role actors.CollaboratorRole) *actors.Collaborator {
	collab := CreateTestCollaborator()
	collab.Role = role
	return collab
}

// CreateTestCollaboratorWithStatus creates a collaborator with a specific status
func CreateTestCollaboratorWithStatus(status actors.CollaboratorStatus) *actors.Collaborator {
	collab := CreateTestCollaborator()
	collab.Status = status
	return collab
}

// CreateTestAdminCollaborator creates an admin collaborator
func CreateTestAdminCollaborator() *actors.Collaborator {
	collab := CreateTestCollaborator()
	collab.Role = actors.CollaboratorRoleAdmin
	collab.AccessLevel = 9
	collab.CanInviteOthers = true
	return collab
}

// CreateTestOwnerCollaborator creates an owner collaborator
func CreateTestOwnerCollaborator() *actors.Collaborator {
	collab := CreateTestCollaborator()
	collab.Role = actors.CollaboratorRoleOwner
	collab.AccessLevel = 10
	collab.CanInviteOthers = true
	return collab
}

// CreateTestManagerCollaborator creates a manager collaborator
func CreateTestManagerCollaborator() *actors.Collaborator {
	collab := CreateTestCollaborator()
	collab.Role = actors.CollaboratorRoleManager
	collab.AccessLevel = 7
	collab.CanInviteOthers = true
	collab.Department = "Engineering"
	collab.JobTitle = "Engineering Manager"
	return collab
}

// CreateTestInactiveCollaborator creates an inactive collaborator
func CreateTestInactiveCollaborator() *actors.Collaborator {
	collab := CreateTestCollaborator()
	collab.Status = actors.CollaboratorStatusInactive
	return collab
}

// CreateTestSuspendedCollaborator creates a suspended collaborator
func CreateTestSuspendedCollaborator() *actors.Collaborator {
	collab := CreateTestCollaborator()
	collab.Status = actors.CollaboratorStatusSuspended
	return collab
}

// CreateTestPendingCollaborator creates a pending collaborator
func CreateTestPendingCollaborator() *actors.Collaborator {
	collab := CreateTestCollaborator()
	collab.Status = actors.CollaboratorStatusPending
	return collab
}

// CreateTestCollaboratorWithDepartment creates a collaborator in a specific department
func CreateTestCollaboratorWithDepartment(department string) *actors.Collaborator {
	collab := CreateTestCollaborator()
	collab.Department = department
	return collab
}

// CreateTestCollaboratorWithEmployeeID creates a collaborator with a specific employee ID
func CreateTestCollaboratorWithEmployeeID(employeeID string) *actors.Collaborator {
	collab := CreateTestCollaborator()
	collab.EmployeeID = employeeID
	return collab
}

// CreateTestCollaboratorWithAccessLevel creates a collaborator with a specific access level
func CreateTestCollaboratorWithAccessLevel(accessLevel int) *actors.Collaborator {
	collab := CreateTestCollaborator()
	collab.AccessLevel = accessLevel
	return collab
}

// CreateTestCollaboratorWithAAAEntityType creates a collaborator with a specific AAA entity type
func CreateTestCollaboratorWithAAAEntityType(aaaEntityID string, entityType actors.AAAEntityType, contextOrgID string) *actors.Collaborator {
	collab := CreateTestCollaborator()
	collab.AAAEntityID = aaaEntityID
	collab.AAAEntityType = entityType
	collab.ContextOrganizationID = contextOrgID

	if entityType == actors.AAAEntityTypeUser {
		collab.UserID = &aaaEntityID
	} else {
		collab.OrganizationID = &aaaEntityID
	}

	return collab
}

// CreateTestCollaboratorWithInvitation creates a collaborator with invitation details
func CreateTestCollaboratorWithInvitation(invitedBy string, invitedAt, acceptedAt *time.Time) *actors.Collaborator {
	collab := CreateTestCollaborator()
	collab.InvitedByUserID = invitedBy
	collab.InvitedAt = invitedAt
	collab.AcceptedAt = acceptedAt
	return collab
}

// CreateTestCollaboratorsArray creates a slice of test collaborators
func CreateTestCollaboratorsArray(count int, contextOrgID string) []*actors.Collaborator {
	collabs := make([]*actors.Collaborator, count)
	for i := 0; i < count; i++ {
		userID := generateCollaboratorUserID(i)
		collabs[i] = CreateTestUserCollaborator(userID, contextOrgID)
		collabs[i].DisplayName = generateCollaboratorName(i)
		collabs[i].Email = generateCollaboratorEmail(i)
	}
	return collabs
}

// CreateTestCollaboratorsArrayWithRole creates collaborators with a specific role
func CreateTestCollaboratorsArrayWithRole(count int, contextOrgID string, role actors.CollaboratorRole) []*actors.Collaborator {
	collabs := make([]*actors.Collaborator, count)
	for i := 0; i < count; i++ {
		userID := generateCollaboratorUserID(i)
		collabs[i] = CreateTestUserCollaborator(userID, contextOrgID)
		collabs[i].Role = role
		collabs[i].DisplayName = generateCollaboratorName(i)
	}
	return collabs
}

// CreateTestCollaboratorsArrayWithStatus creates collaborators with a specific status
func CreateTestCollaboratorsArrayWithStatus(count int, contextOrgID string, status actors.CollaboratorStatus) []*actors.Collaborator {
	collabs := make([]*actors.Collaborator, count)
	for i := 0; i < count; i++ {
		userID := generateCollaboratorUserID(i)
		collabs[i] = CreateTestUserCollaborator(userID, contextOrgID)
		collabs[i].Status = status
		collabs[i].DisplayName = generateCollaboratorName(i)
	}
	return collabs
}

// Helper functions
func generateCollaboratorUserID(index int) string {
	return "user-collab-id-" + string(rune('A'+index))
}

func generateCollaboratorName(index int) string {
	return "Test User " + string(rune('A'+index))
}

func generateCollaboratorEmail(index int) string {
	return "testuser" + string(rune('A'+index)) + "@example.com"
}
