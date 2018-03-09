package server

import (
	"github.com/elxirhealth/service-base/pkg/server"
	"github.com/elxirhealth/user/pkg/server/storage"
	api "github.com/elxirhealth/user/pkg/userapi"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

const (
	// MaxUserEntities is the maximum number of entities a user can be associated with.
	MaxUserEntities = 16

	// MaxEntityUsers is the maximum number of users that can be associated with a single
	// entity.
	MaxEntityUsers = 256
)

var (
	// ErrTooManyUserEntities indicates when associating another entity ID with a given user ID
	// would put the total associated entities above the max value (MaxUserEntities).
	ErrTooManyUserEntities = errors.New("too many associated entities for user ID")

	// ErrTooManyEntityUsers indicates when associating another user ID with the given entity ID
	// would put the total associated users above the max value (MaxEntityUsers).
	ErrTooManyEntityUsers = errors.New("too many associated users for entity ID")
)

// User implements the UserServer interface.
type User struct {
	*server.BaseServer
	config *Config

	storer storage.Storer
}

// newUser creates a new UserServer from the given config.
func newUser(config *Config) (*User, error) {
	baseServer := server.NewBaseServer(config.BaseConfig)
	storer, err := getStorer(config, baseServer.Logger)
	if err != nil {
		return nil, err
	}
	return &User{
		BaseServer: baseServer,
		config:     config,
		storer:     storer,
	}, nil
}

// AddEntity associates an entity ID with the given user ID.
func (u *User) AddEntity(
	ctx context.Context, rq *api.AddEntityRequest,
) (*api.AddEntityResponse, error) {
	u.Logger.Debug("received add entity request", logAddEntityRq(rq)...)
	if err := api.ValidateAddEntityRequest(rq); err != nil {
		return nil, err
	}
	if n, err := u.storer.CountUsers(rq.EntityId); err != nil {
		return nil, err
	} else if n+1 > MaxEntityUsers {
		return nil, ErrTooManyEntityUsers
	}
	if n, err := u.storer.CountEntities(rq.UserId); err != nil {
		return nil, err
	} else if n+1 > MaxUserEntities {
		return nil, ErrTooManyUserEntities
	}
	if err := u.storer.AddEntity(rq.UserId, rq.EntityId); err != nil {
		return nil, err
	}
	u.Logger.Info("added entity to user", logAddEntityRq(rq)...)
	return &api.AddEntityResponse{}, nil
}

// GetEntities gets the associated entity IDs for the given user ID.
func (u *User) GetEntities(
	ctx context.Context, rq *api.GetEntitiesRequest,
) (*api.GetEntitiesResponse, error) {
	u.Logger.Debug("received get entities request", zap.String(logUserID, rq.UserId))
	if err := api.ValidateGetEntitiesRequest(rq); err != nil {
		return nil, err
	}
	entityIDs, err := u.storer.GetEntities(rq.UserId)
	if err != nil {
		return nil, err
	}
	rp := &api.GetEntitiesResponse{EntityIds: entityIDs}
	u.Logger.Info("got entities for user", logGetEntityRp(rq, rp)...)
	return rp, nil
}
