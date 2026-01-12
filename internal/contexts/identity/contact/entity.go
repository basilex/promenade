package contact

import (
	"errors"
	"fmt"
	"strings"

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
}

// NewContact creates a new contact with email
func NewContact(userID uuidv7.UUID, contactType ContactType, label string) (*Contact, error) {
	if err := validateContactType(contactType); err != nil {
		return nil, err
	}

	if label == "" {
		return nil, ErrLabelRequired
	}

	contact := &Contact{
		BaseAggregate: aggregate.NewBaseAggregate(),
		UserID:        userID,
		Type:          contactType,
		Label:         label,
		IsPrimary:     false,
		IsVerified:    false,
		IsPublic:      false,
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
		return nil, fmt.Errorf("%w", err)
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
		return nil, fmt.Errorf("%w", err)
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
		return nil, fmt.Errorf("%w", err)
	}
	contact.Address = &addressVO
	return contact, nil
}

// SetEmail sets the email for this contact
func (c *Contact) SetEmail(email string) error {
	if c.Type != ContactTypeEmail {
		return ErrCannotSetEmailOnNonEmailContact
	}

	emailVO, err := valueobject.NewEmail(email)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	c.Email = &emailVO
	c.Touch()
	return nil
}

// SetPhone sets the phone for this contact
func (c *Contact) SetPhone(phone string) error {
	if c.Type != ContactTypePhone {
		return ErrCannotSetPhoneOnNonPhoneContact
	}

	phoneVO, err := valueobject.NewPhone(phone)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	c.Phone = &phoneVO
	c.Touch()
	return nil
}

// SetAddress sets the address for this contact
func (c *Contact) SetAddress(street, city, country, postalCode string) error {
	if c.Type != ContactTypeAddress {
		return ErrCannotSetAddressOnNonAddressContact
	}

	addressVO, err := valueobject.NewAddress(street, city, postalCode, country)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	c.Address = &addressVO
	c.Touch()
	return nil
}

// SetAsPrimary marks this contact as primary
func (c *Contact) SetAsPrimary() {
	c.IsPrimary = true
	c.Touch()
}

// UnsetAsPrimary unmarks this contact as primary
func (c *Contact) UnsetAsPrimary() {
	c.IsPrimary = false
	c.Touch()
}

// Verify marks contact as verified
func (c *Contact) Verify() {
	c.IsVerified = true
	c.Touch()
}

// Unverify marks contact as not verified
func (c *Contact) Unverify() {
	c.IsVerified = false
	c.Touch()
}

// MakePublic makes contact visible to others
func (c *Contact) MakePublic() {
	c.IsPublic = true
	c.Touch()
}

// MakePrivate makes contact private
func (c *Contact) MakePrivate() {
	c.IsPublic = false
	c.Touch()
}

// UpdateLabel updates the contact label
func (c *Contact) UpdateLabel(label string) error {
	if label == "" {
		return ErrLabelEmpty
	}
	c.Label = label
	c.Touch()
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
		return ErrLabelRequired
	}

	// Check that at least one contact field is present
	switch c.Type {
	case ContactTypeEmail:
		if c.Email == nil {
			return ErrEmailRequiredForEmailContact
		}
		// Email is already validated in NewEmail
	case ContactTypePhone:
		if c.Phone == nil {
			return ErrPhoneRequiredForPhoneContact
		}
		// Phone is already validated in NewPhone
	case ContactTypeAddress:
		if c.Address == nil {
			return ErrAddressRequiredForAddressContact
		}
		// Address is already validated in NewAddress
	default:
		return ErrUnknownContactType
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
		return ErrInvalidContactType
	}
}
