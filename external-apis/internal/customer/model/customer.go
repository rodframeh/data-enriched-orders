package model

// CustomerStatus represents the status of a customer
type CustomerStatus string

const (
	StatusActive   CustomerStatus = "ACTIVE"
	StatusInactive CustomerStatus = "INACTIVE"
	StatusBlocked  CustomerStatus = "BLOCKED"
	StatusPending  CustomerStatus = "PENDING"
)

// Customer represents a customer
type Customer struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Email  string         `json:"email"`
	Phone  string         `json:"phone"`
	Active bool           `json:"active"`
	Status CustomerStatus `json:"status"`
}

// CustomerResponse represents the API response for a customer
type CustomerResponse struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Email  string         `json:"email"`
	Phone  string         `json:"phone"`
	Active bool           `json:"active"`
	Status CustomerStatus `json:"status"`
}

// ToResponse converts a Customer to CustomerResponse
func (c *Customer) ToResponse() CustomerResponse {
	return CustomerResponse{
		ID:     c.ID,
		Name:   c.Name,
		Email:  c.Email,
		Phone:  c.Phone,
		Active: c.Active,
		Status: c.Status,
	}
}

// CreateCustomerRequest represents the request to create a customer
type CreateCustomerRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Phone string `json:"phone" binding:"required"`
}

// UpdateCustomerRequest represents the request to update a customer
type UpdateCustomerRequest struct {
	Name   *string         `json:"name,omitempty"`
	Email  *string         `json:"email,omitempty"`
	Phone  *string         `json:"phone,omitempty"`
	Active *bool           `json:"active,omitempty"`
	Status *CustomerStatus `json:"status,omitempty"`
}

// IsValid checks if the customer status is valid
func (s CustomerStatus) IsValid() bool {
	switch s {
	case StatusActive, StatusInactive, StatusBlocked, StatusPending:
		return true
	default:
		return false
	}
}
