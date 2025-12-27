package contact

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// Common errors
var (
	ErrNotFound = errors.New("contact not found")
)

// ContactType defines the type of contact information
type ContactType string

const (
	ContactTypeEmail   ContactType = "email"
	ContactTypePhone   ContactType = "phone"
	ContactTypeAddress ContactType = "address"
)

// Contact represents a user's contact information (email, phone, or address)
// This is an Aggregate Root in the Identity bounded context
type Contact struct {
	aggregate.BaseAggregate

	// Identity
	ID     uuidv7.UUID
	UserID uuidv7.UUID

	// Type and Label
	Type  ContactType
	Label string // "Work", "Home", "Mobile", etc.

	// Contact Information (at least one must be present)
	Email   *valueobject.Email
	Phone   *valueobject.Phone
	Address *valueobject.Address

	// Flags
	IsPrimary  bool // Only one primary contact per user per type
	IsVerified bool // Whether contact has been verified
	IsPublic   bool // Whether contact is visible to others

	// Lifecycle
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewContact creates a new contact with email
func NewContact(userID uuidv7.UUID, contactType ContactType, label string) (*Contact, error) {
	if err := validateContactType(contactType); err != nil {
		return nil, err
	}

	if label == "" {
		return nil, fmt.Errorf("label is required")
	}

	now := time.Now()
	contact := &Contact{
		BaseAggregate: aggregate.NewBaseAggregate(),
		ID:            uuidv7.New(),
		UserID:        userID,
		Type:          contactType,
		Label:         label,
		IsPrimary:     false,
		IsVerified:    false,
		IsPublic:      false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	return contact, nil
}

// NewEmailContact creates a contact with email
func NewEmailContact(userID uuidv7.UUID, email string, label string) (*Contact, error) {
	contact, err := NewContact(userID, ContactTypeEmail, label)
	if err != nil {
		return nil, err
	}

	emailVO, err := valueobject.NewEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	contact.Email = &emailVO
	return contact, nil
}

// NewPhoneContact creates a contact with phone
func NewPhoneContact(userID uuidv7.UUID, phone string, label string) (*Contact, error) {
	contact, err := NewContact(userID, ContactTypePhone, label)
	if err != nil {
		return nil, err
	}

	phoneVO, err := valueobject.NewPhone(phone)
	if err != nil {
		return nil, fmt.Errorf("invalid phone: %w", err)
	}

	contact.Phone = &phoneVO
	return contact, nil
}

// NewAddressContact creates a contact with address
func NewAddressContact(userID uuidv7.UUID, street, city, country, postalCode string, label string) (*Contact, error) {
	contact, err := NewContact(userID, ContactTypeAddress, label)
	if err != nil {
		return nil, err
	}

	addressVO, err := valueobject.NewAddress(street, city, postalCode, country)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}
	contact.Address = &addressVO
	return contact, nil
}

// SetEmail sets the email for this contact
func (c *Contact) SetEmail(email string) error {
	if c.Type != ContactTypeEmail {
		return fmt.Errorf("cannot set email on non-email contact")
	}

	emailVO, err := valueobject.NewEmail(email)
	if err != nil {
		return fmt.Errorf("invalid email: %w", err)
	}

	c.Email = &emailVO
	c.UpdatedAt = time.Now()
	return nil
}

// SetPhone sets the phone for this contact
func (c *Contact) SetPhone(phone string) error {
	if c.Type != ContactTypePhone {
		return fmt.Errorf("cannot set phone on non-phone contact")
	}

	phoneVO, err := valueobject.NewPhone(phone)
	if err != nil {
		return fmt.Errorf("invalid phone: %w", err)
	}

	c.Phone = &phoneVO
	c.UpdatedAt = time.Now()
	return nil
}

// SetAddress sets the address for this contact
func (c *Contact) SetAddress(street, city, country, postalCode string) error {
	if c.Type != ContactTypeAddress {
		return fmt.Errorf("cannot set address on non-address contact")
	}

	addressVO, err := valueobject.NewAddress(street, city, postalCode, country)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}
	c.Address = &addressVO
	c.UpdatedAt = time.Now()
	return nil
}

// SetAsPrimary marks this contact as primary
func (c *Contact) SetAsPrimary() {
	c.IsPrimary = true
	c.UpdatedAt = time.Now()
}

// UnsetAsPrimary unmarks this contact as primary
func (c *Contact) UnsetAsPrimary() {
	c.IsPrimary = false
	c.UpdatedAt = time.Now()
}

// Verify marks contact as verified
func (c *Contact) Verify() {
	c.IsVerified = true
	c.UpdatedAt = time.Now()
}

// Unverify marks contact as not verified
func (c *Contact) Unverify() {
	c.IsVerified = false
	c.UpdatedAt = time.Now()
}

// MakePublic makes contact visible to others
func (c *Contact) MakePublic() {
	c.IsPublic = true
	c.UpdatedAt = time.Now()
}

// MakePrivate makes contact private
func (c *Contact) MakePrivate() {
	c.IsPublic = false
	c.UpdatedAt = time.Now()
}

// UpdateLabel updates the contact label
func (c *Contact) UpdateLabel(label string) error {
	if label == "" {
		return fmt.Errorf("label cannot be empty")
	}
	c.Label = label
	c.UpdatedAt = time.Now()
	return nil
}

// Validate validates the contact
func (c *Contact) Validate() error {
	// Check contact type
	if err := validateContactType(c.Type); err != nil {
		return err
	}

	// Check label
	if c.Label == "" {
		return fmt.Errorf("label is required")
	}

	// Check that at least one contact field is present
	switch c.Type {
	case ContactTypeEmail:
		if c.Email == nil {
			return fmt.Errorf("email is required for email contact")
		}
		// Email is already validated in NewEmail
	case ContactTypePhone:
		if c.Phone == nil {
			return fmt.Errorf("phone is required for phone contact")
		}
		// Phone is already validated in NewPhone
	case ContactTypeAddress:
		if c.Address == nil {
			return fmt.Errorf("address is required for address contact")
		}
		// Address is already validated in NewAddress
	default:
		return fmt.Errorf("unknown contact type: %s", c.Type)
	}

	return nil
}

// GetValue returns the string representation of the contact value
func (c *Contact) GetValue() string {
	switch c.Type {
	case ContactTypeEmail:
		if c.Email != nil {
			return c.Email.Value()
		}
	case ContactTypePhone:
		if c.Phone != nil {
			return c.Phone.Value()
		}
	case ContactTypeAddress:
		if c.Address != nil {
			return c.Address.String()
		}
	}
	return ""
}

// validateContactType validates the contact type
func validateContactType(t ContactType) error {
	t = ContactType(strings.ToLower(string(t)))

	switch t {
	case ContactTypeEmail, ContactTypePhone, ContactTypeAddress:
		return nil
	default:
		return fmt.Errorf("invalid contact type: %s (must be email, phone, or address)", t)
	}
}
