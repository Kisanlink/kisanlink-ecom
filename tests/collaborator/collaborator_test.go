package collaborator_test

import (
	"testing"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/collaborator"
)

func TestCollaboratorModel(t *testing.T) {
	// Test creating a new collaborator
	collab := collaborator.NewCollaborator(
		"user123",
		"testuser",
		"test@example.com",
		"John",
		"Doe",
		collaborator.CollaboratorTypeFarmer,
	)

	// Test basic properties
	if collab.UserID != "user123" {
		t.Errorf("Expected UserID to be 'user123', got %s", collab.UserID)
	}

	if collab.Username != "testuser" {
		t.Errorf("Expected Username to be 'testuser', got %s", collab.Username)
	}

	if collab.Email != "test@example.com" {
		t.Errorf("Expected Email to be 'test@example.com', got %s", collab.Email)
	}

	if collab.CollaboratorType != collaborator.CollaboratorTypeFarmer {
		t.Errorf("Expected CollaboratorType to be FARMER, got %s", collab.CollaboratorType)
	}

	if collab.Status != collaborator.CollaboratorStatusPending {
		t.Errorf("Expected Status to be PENDING, got %s", collab.Status)
	}

	// Test methods
	fullName := collab.GetFullName()
	expectedFullName := "John Doe"
	if fullName != expectedFullName {
		t.Errorf("Expected full name to be '%s', got '%s'", expectedFullName, fullName)
	}

	// Test status methods
	if collab.IsActive() {
		t.Error("Expected collaborator to not be active initially")
	}

	if !collab.CanLogin() {
		t.Error("Expected collaborator to be able to login in PENDING status")
	}

	// Test activation
	collab.Activate()
	if collab.Status != collaborator.CollaboratorStatusActive {
		t.Errorf("Expected Status to be ACTIVE after activation, got %s", collab.Status)
	}

	if !collab.IsActive() {
		t.Error("Expected collaborator to be active after activation")
	}

	// Test verification
	collab.Verify("admin123")
	if !collab.IsVerified {
		t.Error("Expected collaborator to be verified")
	}

	if collab.VerifiedBy != "admin123" {
		t.Errorf("Expected VerifiedBy to be 'admin123', got %s", collab.VerifiedBy)
	}

	// Test onboarding
	collab.SetOnboardingStep(3)
	if collab.OnboardingStep != 3 {
		t.Errorf("Expected OnboardingStep to be 3, got %d", collab.OnboardingStep)
	}

	collab.CompleteOnboarding()
	if !collab.OnboardingCompleted {
		t.Error("Expected onboarding to be completed")
	}

	if collab.OnboardingStep != 0 {
		t.Errorf("Expected OnboardingStep to be reset to 0 after completion, got %d", collab.OnboardingStep)
	}

	// Test trust score calculation
	collab.CompletedOrders = 10
	collab.CancelledOrders = 2
	collab.AverageRating = 4.5
	collab.UpdateTrustScore()

	// Trust score should be calculated based on the formula
	// Base: 50, Orders: 10*2=20, Cancelled: 2*-5=-10, Rating: 4.5*10=45, Verified: 20
	// Total: 50 + 20 - 10 + 45 + 20 = 125, capped at 100
	expectedTrustScore := 100.0
	if collab.TrustScore != expectedTrustScore {
		t.Errorf("Expected TrustScore to be %.1f, got %.1f", expectedTrustScore, collab.TrustScore)
	}
}

func TestCollaboratorTypes(t *testing.T) {
	types := []collaborator.CollaboratorType{
		collaborator.CollaboratorTypeFarmer,
		collaborator.CollaboratorTypeSupplier,
		collaborator.CollaboratorTypeBuyer,
		collaborator.CollaboratorTypeAgent,
		collaborator.CollaboratorTypeAdmin,
	}

	expectedTypes := []string{"FARMER", "SUPPLIER", "BUYER", "AGENT", "ADMIN"}

	for i, collaboratorType := range types {
		if string(collaboratorType) != expectedTypes[i] {
			t.Errorf("Expected type %d to be %s, got %s", i, expectedTypes[i], string(collaboratorType))
		}
	}
}

func TestCollaboratorStatuses(t *testing.T) {
	statuses := []collaborator.CollaboratorStatus{
		collaborator.CollaboratorStatusActive,
		collaborator.CollaboratorStatusInactive,
		collaborator.CollaboratorStatusSuspended,
		collaborator.CollaboratorStatusPending,
	}

	expectedStatuses := []string{"ACTIVE", "INACTIVE", "SUSPENDED", "PENDING"}

	for i, status := range statuses {
		if string(status) != expectedStatuses[i] {
			t.Errorf("Expected status %d to be %s, got %s", i, expectedStatuses[i], string(status))
		}
	}
}
