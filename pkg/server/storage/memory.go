package storage

import (
	"sync"

	api "github.com/elixirhealth/user/pkg/userapi"
	"go.uber.org/zap"
)

type memoryStorer struct {
	params       *Parameters
	logger       *zap.Logger
	userEntities []*UserEntity
	mu           sync.Mutex
}

// NewMemory creates a new Storer backed by an in-memory list.
func NewMemory(params *Parameters, logger *zap.Logger) Storer {
	return &memoryStorer{
		userEntities: make([]*UserEntity, 0),
		params:       params,
		logger:       logger,
	}
}

func (s *memoryStorer) AddEntity(userID, entityID string) error {
	if userID == "" {
		return api.ErrEmptyUserID
	}
	if entityID == "" {
		return api.ErrEmptyEntityID
	}

	// check user-entity association doesn't already exist
	pred := func(ue *UserEntity) bool {
		return ue.EntityID == entityID && ue.UserID == userID
	}
	if n := s.count(pred); n > 0 {
		return ErrUserEntityExists
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.userEntities = append(s.userEntities, newUserEntity(userID, entityID))
	s.logger.Debug("storer added entity to user", logAddEntityFields(userID, entityID)...)
	return nil
}

func (s *memoryStorer) GetEntities(userID string) ([]string, error) {
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

func (s *memoryStorer) CountEntities(userID string) (int, error) {
	if userID == "" {
		return 0, api.ErrEmptyUserID
	}
	n := s.count(func(ue *UserEntity) bool { return ue.UserID == userID })
	s.logger.Debug("storer counted entities for user", logCountEntities(userID, n)...)
	return n, nil
}

func (s *memoryStorer) CountUsers(entityID string) (int, error) {
	if entityID == "" {
		return 0, api.ErrEmptyEntityID
	}
	n := s.count(func(ue *UserEntity) bool { return ue.EntityID == entityID })
	s.logger.Debug("storer counted users for entity", logCountUsers(entityID, n)...)
	return n, nil
}

func (s *memoryStorer) count(predicate func(ue *UserEntity) bool) int {
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
