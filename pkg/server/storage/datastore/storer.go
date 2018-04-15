package datastore

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
	bstorage "github.com/elixirhealth/service-base/pkg/server/storage"
	"github.com/elixirhealth/user/pkg/server/storage"
	api "github.com/elixirhealth/user/pkg/userapi"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
)

const (
	userEntityKind = "user_entity"

	secsPerDay = int64(3600 * 24 * 24)
)

var (
	// ErrUserEntityExists indicates when a (user ID, entity ID) pair already exists
	ErrUserEntityExists = errors.New("user-entity association already exists")
)

// UserEntity represents an entity-user association.
type UserEntity struct {
	UserID       string    `datastore:"user_id"`
	EntityID     string    `datastore:"entity_id"`
	Removed      bool      `datastore:"removed"`
	ModifiedDate int32     `datastore:"modified_date"`
	ModifiedTime time.Time `datastore:"modified_time,noindex"`
	AddedTime    time.Time `datastore:"added_time,noindex"`
	RemovedTime  time.Time `datastore:"removed_time,noindex"`
}

type storer struct {
	params *storage.Parameters
	client bstorage.DatastoreClient
	iter   bstorage.DatastoreIterator
	logger *zap.Logger
}

// New creates a new Storer backed by a GCP DataStore instance.
func New(
	gcpProjectID string, params *storage.Parameters, logger *zap.Logger,
) (storage.Storer, error) {
	client, err := datastore.NewClient(context.Background(), gcpProjectID)
	if err != nil {
		return nil, err
	}
	return &storer{
		params: params,
		client: &bstorage.DatastoreClientImpl{Inner: client},
		iter:   &bstorage.DatastoreIteratorImpl{},
		logger: logger,
	}, nil
}

func (s *storer) AddEntity(userID, entityID string) error {
	if userID == "" {
		return api.ErrEmptyUserID
	}
	if entityID == "" {
		return api.ErrEmptyEntityID
	}

	// check user-entity association doesn't already exist
	if n, err := s.countUserEntities(userID, entityID); err != nil {
		return err
	} else if n > 0 {
		return ErrUserEntityExists
	}

	key := datastore.IncompleteKey(userEntityKind, nil)
	ue := NewUserEntity(userID, entityID)
	ctx, cancel := context.WithTimeout(context.Background(), s.params.AddQueryTimeout)
	defer cancel()
	if _, err := s.client.Put(ctx, key, ue); err != nil {
		return err
	}
	s.logger.Debug("storer added entity to user", logAddEntityFields(userID, entityID)...)
	return nil
}

func (s *storer) GetEntities(userID string) ([]string, error) {
	if userID == "" {
		return nil, api.ErrEmptyUserID
	}
	q := getEntitiesQuery(userID)
	ctx, cancel := context.WithTimeout(context.Background(), s.params.GetQueryTimeout)
	defer cancel()
	iter := s.client.Run(ctx, q)
	s.iter.Init(iter)
	entityIDs := make([]string, 0)
	for {
		ue := &UserEntity{}
		if _, err := s.iter.Next(ue); err == iterator.Done {
			// no more results
			break
		} else if err != nil {
			return nil, err
		}
		entityIDs = append(entityIDs, ue.EntityID)
	}
	s.logger.Debug("storer got entities for user", logGetEntities(userID, entityIDs)...)
	return entityIDs, nil
}

func (s *storer) CountEntities(userID string) (int, error) {
	if userID == "" {
		return 0, api.ErrEmptyUserID
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.params.GetQueryTimeout)
	defer cancel()
	n, err := s.client.Count(ctx, getEntitiesQuery(userID))
	if err != nil {
		return 0, err
	}
	s.logger.Debug("storer counted entities for user", logCountEntities(userID, n)...)
	return n, nil
}

func (s *storer) CountUsers(entityID string) (int, error) {
	if entityID == "" {
		return 0, api.ErrEmptyEntityID
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.params.GetQueryTimeout)
	defer cancel()
	n, err := s.client.Count(ctx, getUsersQuery(entityID))
	if err != nil {
		return 0, err
	}
	s.logger.Debug("storer counted users for entity", logCountUsers(entityID, n)...)
	return n, nil
}

func (s *storer) countUserEntities(userID, entityID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.params.CountQueryTimeout)
	defer cancel()
	q := getEntitiesQuery(userID).Filter("entity_id = ", entityID)
	n, err := s.client.Count(ctx, q)
	if err != nil {
		return 0, err
	}
	s.logger.Debug("storer counted user entities", logCountUserEntities(userID, entityID, n)...)
	return n, nil
}

func (s *storer) Close() error {
	return nil
}

// NewUserEntity creates a new *UserEntity from the given user and entity ID.
func NewUserEntity(userID, entityID string) *UserEntity {
	now := time.Now()
	return &UserEntity{
		UserID:       userID,
		EntityID:     entityID,
		Removed:      false,
		ModifiedDate: int32(now.Unix() / secsPerDay),
		ModifiedTime: now,
		AddedTime:    now,
	}
}

func getEntitiesQuery(userID string) *datastore.Query {
	return datastore.NewQuery(userEntityKind).
		Filter("user_id = ", userID).
		Filter("removed = ", false)
}

func getUsersQuery(entityID string) *datastore.Query {
	return datastore.NewQuery(userEntityKind).
		Filter("entity_id = ", entityID).
		Filter("removed = ", false)
}
