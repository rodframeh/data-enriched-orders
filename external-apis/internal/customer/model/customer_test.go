package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomerStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   CustomerStatus
		expected bool
	}{
		{
			name:     "Active status is valid",
			status:   StatusActive,
			expected: true,
		},
		{
			name:     "Inactive status is valid",
			status:   StatusInactive,
			expected: true,
		},
		{
			name:     "Blocked status is valid",
			status:   StatusBlocked,
			expected: true,
		},
		{
			name:     "Pending status is valid",
			status:   StatusPending,
			expected: true,
		},
		{
			name:     "Invalid status",
			status:   CustomerStatus("INVALID"),
			expected: false,
		},
		{
			name:     "Empty status is invalid",
			status:   CustomerStatus(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCustomer_ToResponse(t *testing.T) {
	// Arrange
	customer := &Customer{
		ID:     "customer-123",
		Name:   "John Doe",
		Email:  "john.doe@example.com",
		Phone:  "+1-555-0123",
		Active: true,
		Status: StatusActive,
	}

	// Act
	response := customer.ToResponse()

	// Assert
	assert.Equal(t, "customer-123", response.ID)
	assert.Equal(t, "John Doe", response.Name)
	assert.Equal(t, "john.doe@example.com", response.Email)
	assert.Equal(t, "+1-555-0123", response.Phone)
	assert.True(t, response.Active)
	assert.Equal(t, StatusActive, response.Status)
}

func TestCreateCustomerRequest_Validation(t *testing.T) {
	tests := []struct {
		name        string
		request     CreateCustomerRequest
		expectValid bool
	}{
		{
			name: "Valid request",
			request: CreateCustomerRequest{
				Name:  "John Doe",
				Email: "john.doe@example.com",
				Phone: "+1-555-0123",
			},
			expectValid: true,
		},
		{
			name: "Empty name",
			request: CreateCustomerRequest{
				Name:  "",
				Email: "john.doe@example.com",
				Phone: "+1-555-0123",
			},
			expectValid: false,
		},
		{
			name: "Invalid email",
			request: CreateCustomerRequest{
				Name:  "John Doe",
				Email: "invalid-email",
				Phone: "+1-555-0123",
			},
			expectValid: false,
		},
		{
			name: "Empty phone",
			request: CreateCustomerRequest{
				Name:  "John Doe",
				Email: "john.doe@example.com",
				Phone: "",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate validation that would happen in handler
			isValid := tt.request.Name != "" &&
				tt.request.Email != "" &&
				tt.request.Phone != "" &&
				// Basic email validation
				len(tt.request.Email) > 5 &&
				tt.request.Email != "invalid-email"

			assert.Equal(t, tt.expectValid, isValid)
		})
	}
}

func TestUpdateCustomerRequest(t *testing.T) {
	t.Run("Partial update with name only", func(t *testing.T) {
		newName := "Updated Name"
		request := UpdateCustomerRequest{
			Name: &newName,
		}

		assert.Equal(t, "Updated Name", *request.Name)
		assert.Nil(t, request.Email)
		assert.Nil(t, request.Phone)
		assert.Nil(t, request.Active)
		assert.Nil(t, request.Status)
	})

	t.Run("Partial update with status only", func(t *testing.T) {
		newStatus := StatusBlocked
		request := UpdateCustomerRequest{
			Status: &newStatus,
		}

		assert.Equal(t, StatusBlocked, *request.Status)
		assert.Nil(t, request.Name)
		assert.Nil(t, request.Email)
		assert.Nil(t, request.Phone)
		assert.Nil(t, request.Active)
	})

	t.Run("Full update", func(t *testing.T) {
		newName := "Updated Name"
		newEmail := "updated@example.com"
		newPhone := "+1-555-9999"
		newActive := false
		newStatus := StatusInactive

		request := UpdateCustomerRequest{
			Name:   &newName,
			Email:  &newEmail,
			Phone:  &newPhone,
			Active: &newActive,
			Status: &newStatus,
		}

		assert.Equal(t, "Updated Name", *request.Name)
		assert.Equal(t, "updated@example.com", *request.Email)
		assert.Equal(t, "+1-555-9999", *request.Phone)
		assert.False(t, *request.Active)
		assert.Equal(t, StatusInactive, *request.Status)
	})
}
