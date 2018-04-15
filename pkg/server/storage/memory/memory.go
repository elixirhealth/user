package memory

import (
	"sync"

	"github.com/elixirhealth/user/pkg/server/storage"
	"github.com/elixirhealth/user/pkg/server/storage/datastore"
	api "github.com/elixirhealth/user/pkg/userapi"
	"go.uber.org/zap"
)

type storer struct {
	params       *storage.Parameters
	logger       *zap.Logger
	userEntities []*datastore.UserEntity
	mu           sync.Mutex
}

// New creates a new Storer backed by an in-memory list.
func New(params *storage.Parameters, logger *zap.Logger) storage.Storer {
	return &storer{
		userEntities: make([]*datastore.UserEntity, 0),
		params:       params,
		logger:       logger,
	}
}

func (s *storer) AddEntity(userID, entityID string) error {
	if userID == "" {
		return api.ErrEmptyUserID
	}
	if entityID == "" {
		return api.ErrEmptyEntityID
	}

	// check user-entity association doesn't already exist
	pred := func(ue *datastore.UserEntity) bool {
		return ue.EntityID == entityID && ue.UserID == userID
	}
	if n := s.count(pred); n > 0 {
		return datastore.ErrUserEntityExists
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.userEntities = append(s.userEntities, datastore.NewUserEntity(userID, entityID))
	s.logger.Debug("storer added entity to user", logAddEntityFields(userID, entityID)...)
	return nil
}

func (s *storer) GetEntities(userID string) ([]string, error) {
	if userID == "" {
		return nil, api.ErrEmptyUserID
	}
	entityIDs := make([]string, 0)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, ue := range s.userEntities {
		if ue.UserID == userID {
			entityIDs = append(entityIDs, ue.EntityID)
		}
	}
	s.logger.Debug("storer got entities for user", logGetEntities(userID, entityIDs)...)
	return entityIDs, nil
}

func (s *storer) CountEntities(userID string) (int, error) {
	if userID == "" {
		return 0, api.ErrEmptyUserID
	}
	n := s.count(func(ue *datastore.UserEntity) bool { return ue.UserID == userID })
	s.logger.Debug("storer counted entities for user", logCountEntities(userID, n)...)
	return n, nil
}

func (s *storer) CountUsers(entityID string) (int, error) {
	if entityID == "" {
		return 0, api.ErrEmptyEntityID
	}
	n := s.count(func(ue *datastore.UserEntity) bool { return ue.EntityID == entityID })
	s.logger.Debug("storer counted users for entity", logCountUsers(entityID, n)...)
	return n, nil
}

func (s *storer) Close() error {
	return nil
}

func (s *storer) count(predicate func(ue *datastore.UserEntity) bool) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, ue := range s.userEntities {
		if predicate(ue) {
			n++
		}
	}
	return n
}
