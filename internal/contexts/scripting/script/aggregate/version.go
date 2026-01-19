package aggregate

import (
	"time"

	scripterrors "github.com/basilex/promenade/internal/contexts/scripting/script"
	"github.com/basilex/promenade/pkg/jsonstore"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ScriptVersion represents an immutable snapshot of a script version.
type ScriptVersion struct {
	ID        uuidv7.UUID
	ScriptID  uuidv7.UUID
	Version   int
	Code      string
	Metadata  jsonstore.Field[map[string]string]
	ChangeLog string
	CreatedBy *uuidv7.UUID
	CreatedAt time.Time
}

// NewScriptVersion creates a new script version snapshot.
func NewScriptVersion(s *Script, changeLog string, createdBy *uuidv7.UUID) (*ScriptVersion, error) {
	if s == nil {
		return nil, scripterrors.ErrScriptNotFound
	}
	if s.GetID() == uuidv7.Nil {
		return nil, scripterrors.ErrScriptExecutionScriptIDNil
	}
	if s.Version < 1 {
		return nil, scripterrors.ErrScriptVersionInvalid
	}
	if s.Code == "" {
		return nil, scripterrors.ErrScriptCodeEmpty
	}

	return &ScriptVersion{
		ID:        uuidv7.New(),
		ScriptID:  s.GetID(),
		Version:   s.Version,
		Code:      s.Code,
		Metadata:  s.Metadata,
		ChangeLog: changeLog,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
	}, nil
}

// Validate ensures the version snapshot is valid.
func (v *ScriptVersion) Validate() error {
	if v.ScriptID == uuidv7.Nil {
		return scripterrors.ErrScriptExecutionScriptIDNil
	}
	if v.Version < 1 {
		return scripterrors.ErrScriptVersionInvalid
	}
	if v.Code == "" {
		return scripterrors.ErrScriptCodeEmpty
	}
	return nil
}
