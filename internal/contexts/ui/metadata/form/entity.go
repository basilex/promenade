package form

import (
	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/jsonstore"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// FormDefinition represents a metadata-driven UI form definition.
type FormDefinition struct {
	aggregate.BaseAggregate

	FormID      string                                        // Unique form identifier
	EntityType  string                                        // Target entity (customer, order, deal, etc.)
	Name        string                                        // Display name
	Description string                                        // Optional description
	Layout      jsonstore.Field[map[string]any]               // Layout metadata
	Fields      jsonstore.Field[[]map[string]any]             // Field definitions
	Validation  jsonstore.Field[map[string]any]               // Validation rules
	Events      jsonstore.Field[map[string]any]               // Event handlers
	Permissions jsonstore.Field[map[string]any]               // RBAC permissions
	I18n        jsonstore.Field[map[string]map[string]string] // Localization map
	IsActive    bool                                          // Active flag
	TenantID    *uuidv7.UUID                                  // Optional tenant scope
	CreatedBy   *uuidv7.UUID                                  // Optional creator
}

// NewFormDefinition creates a new form definition with required fields.
func NewFormDefinition(formID, entityType, name string, layout map[string]any, fields []map[string]any) (*FormDefinition, error) {
	if formID == "" {
		return nil, ErrFormIDRequired
	}
	if entityType == "" {
		return nil, ErrEntityTypeRequired
	}
	if name == "" {
		return nil, ErrFormNameRequired
	}
	if layout == nil {
		return nil, ErrLayoutRequired
	}
	if fields == nil {
		return nil, ErrFieldsRequired
	}

	return &FormDefinition{
		BaseAggregate: aggregate.NewBaseAggregate(),
		FormID:        formID,
		EntityType:    entityType,
		Name:          name,
		Layout:        jsonstore.NewField(layout),
		Fields:        jsonstore.NewField(fields),
		Validation:    jsonstore.NewNullField[map[string]any](),
		Events:        jsonstore.NewNullField[map[string]any](),
		Permissions:   jsonstore.NewNullField[map[string]any](),
		I18n:          jsonstore.NewNullField[map[string]map[string]string](),
		IsActive:      true,
	}, nil
}

// UpdateMetadata updates form metadata and increments version.
func (f *FormDefinition) UpdateMetadata(
	name, description string,
	layout map[string]any,
	fields []map[string]any,
	validation map[string]any,
	events map[string]any,
	permissions map[string]any,
	i18n map[string]map[string]string,
	isActive bool,
	entityType string,
	tenantID *uuidv7.UUID,
) error {
	if name == "" {
		return ErrFormNameRequired
	}
	if layout == nil {
		return ErrLayoutRequired
	}
	if fields == nil {
		return ErrFieldsRequired
	}
	if entityType == "" {
		return ErrEntityTypeRequired
	}

	f.Name = name
	f.Description = description
	f.EntityType = entityType
	f.IsActive = isActive
	f.TenantID = tenantID

	f.Layout.Set(layout)
	f.Fields.Set(fields)

	if validation == nil {
		f.Validation.SetNull()
	} else {
		f.Validation.Set(validation)
	}

	if events == nil {
		f.Events.SetNull()
	} else {
		f.Events.Set(events)
	}

	if permissions == nil {
		f.Permissions.SetNull()
	} else {
		f.Permissions.Set(permissions)
	}

	if i18n == nil {
		f.I18n.SetNull()
	} else {
		f.I18n.Set(i18n)
	}

	f.IncrementVersion()
	return nil
}

// Validate checks the form definition for required fields.
func (f *FormDefinition) Validate() error {
	if f.FormID == "" {
		return ErrFormIDRequired
	}
	if f.EntityType == "" {
		return ErrEntityTypeRequired
	}
	if f.Name == "" {
		return ErrFormNameRequired
	}
	if f.Layout.IsNull() {
		return ErrLayoutRequired
	}
	if f.Fields.IsNull() {
		return ErrFieldsRequired
	}
	return nil
}
