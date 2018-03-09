package storage

import (
	"time"

	bstorage "github.com/elxirhealth/service-base/pkg/server/storage"
	"go.uber.org/zap/zapcore"
)

const (
// MaxUserEntities is the maximum number of entities a user can be associated with.
//MaxUserEntities = 16

// MaxEntityUsers is the maximum number of users that can be associated with a single
// entity.
//MaxEntityUsers = 256
)

var (
	// DefaultType is the default storage type.
	DefaultType = bstorage.Memory

	// DefaultQueryTimeout is the default timeout for DataStore queries.
	DefaultQueryTimeout = 1 * time.Second
)

// Storer stores and retrieves user attributes.
type Storer interface {
	AddEntity(userID, entityID string) error
	GetEntities(userID string) ([]string, error)
	CountEntities(userID string) (int, error)
	CountUsers(entityID string) (int, error)
}

// Parameters defines the parameters of the Storer.
type Parameters struct {
	Type              bstorage.Type
	AddQueryTimeout   time.Duration
	GetQueryTimeout   time.Duration
	CountQueryTimeout time.Duration
}

// NewDefaultParameters returns a *Parameters object with default values.
func NewDefaultParameters() *Parameters {
	return &Parameters{
		Type:              DefaultType,
		AddQueryTimeout:   DefaultQueryTimeout,
		GetQueryTimeout:   DefaultQueryTimeout,
		CountQueryTimeout: DefaultQueryTimeout,
	}
}

// MarshalLogObject writes the parameters to the given object encoder.
func (p *Parameters) MarshalLogObject(oe zapcore.ObjectEncoder) error {
	oe.AddString(logType, p.Type.String())
	oe.AddDuration(logAddQueryTimeout, p.AddQueryTimeout)
	oe.AddDuration(logGetQueryTimeout, p.GetQueryTimeout)
	return nil
}
