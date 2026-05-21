package MODULENAME

import "github.com/google/uuid"

// CreateRequest is the DTO for creating an entity.
type CreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateRequest is the DTO for updating an entity.
type UpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ResponseDTO is the DTO for entity responses.
type ResponseDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}
