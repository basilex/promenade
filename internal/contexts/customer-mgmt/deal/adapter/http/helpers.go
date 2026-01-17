package http

import (
"github.com/basilex/promenade/pkg/uuidv7"
)

// parseCustomerID converts string to UUID
func parseCustomerID(customerIDStr string) (uuidv7.UUID, error) {
	return uuidv7.Parse(customerIDStr)
}

// parseAssignedTo converts optional string to UUID pointer
func parseAssignedTo(assignedToStr *string) (*uuidv7.UUID, error) {
	if assignedToStr == nil || *assignedToStr == "" {
		return nil, nil
	}
	id, err := uuidv7.Parse(*assignedToStr)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
