package storage

import (
	"context"
	"testing"

	"cloud.google.com/go/datastore"
	api "github.com/elixirhealth/user/pkg/userapi"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
)

var (
	errTest = errors.New("some test error")
)

func TestDatastoreStorer_AddEntity_ok(t *testing.T) {
	params := NewDefaultParameters()
	lg := zap.NewNop()
	client := &fixedDatastoreClient{
		putValues: make([]*UserEntity, 0),
	}
	s := &datastoreStorer{
		params: params,
		client: client,
		logger: lg,
	}

	userID, entityID := "some user", "some entity"
	err := s.AddEntity(userID, entityID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(client.putValues))

	putValue := client.putValues[0]
	assert.Equal(t, userID, putValue.UserID)
	assert.Equal(t, entityID, putValue.EntityID)
	assert.False(t, putValue.Removed)
	assert.NotZero(t, putValue.ModifiedDate)
	assert.NotZero(t, putValue.ModifiedTime)
	assert.NotZero(t, putValue.AddedTime)
	assert.Zero(t, putValue.RemovedTime)
}

func TestDatastoreStorer_AddEntity_err(t *testing.T) {
	params := NewDefaultParameters()
	lg := zap.NewNop()
	userID, entityID := "some user", "some entity"
	cases := map[string]struct {
		s        *datastoreStorer
		userID   string
		entityID string
		expected error
	}{
		"empty user ID": {
			s:        &datastoreStorer{params: params, logger: lg},
			userID:   "",
			entityID: entityID,
			expected: api.ErrEmptyUserID,
		},
		"empty entity ID": {
			s:        &datastoreStorer{params: params, logger: lg},
			userID:   userID,
			entityID: "",
			expected: api.ErrEmptyEntityID,
		},
		"user entities count err": {
			s: &datastoreStorer{
				params: params,
				logger: lg,
				client: &fixedDatastoreClient{
					countErr: errTest,
				},
			},
			userID:   userID,
			entityID: entityID,
			expected: errTest,
		},
		"non-zero user entities count": {
			s: &datastoreStorer{
				params: params,
				logger: lg,
				client: &fixedDatastoreClient{
					countValue: 1,
				},
			},
			userID:   userID,
			entityID: entityID,
			expected: ErrUserEntityExists,
		},
		"put err": {
			s: &datastoreStorer{
				params: params,
				logger: lg,
				client: &fixedDatastoreClient{
					putErr: errTest,
				},
			},
			userID:   userID,
			entityID: entityID,
			expected: errTest,
		},
	}
	for desc, c := range cases {
		err := c.s.AddEntity(c.userID, c.entityID)
		assert.Equal(t, c.expected, err, desc)
	}
}

func TestDatastoreStorer_GetEntities_ok(t *testing.T) {
	params := NewDefaultParameters()
	lg := zap.NewNop()

	userID := "some user ID"
	entityID1, entityID2 := "some entity ID", "another entity ID"
	getResult := []*UserEntity{
		{UserID: userID, EntityID: entityID1},
		{UserID: userID, EntityID: entityID2},
	}
	s := &datastoreStorer{
		params: params,
		client: &fixedDatastoreClient{},
		iter: &fixedDatastoreIter{
			keys: []*datastore.Key{
				datastore.IDKey(userEntityKind, 0, nil),
				datastore.IDKey(userEntityKind, 1, nil),
			},
			values: getResult,
		},
		logger: lg,
	}

	entityIDs, err := s.GetEntities(userID)
	assert.Nil(t, err)
	assert.Equal(t, []string{entityID1, entityID2}, entityIDs)
}

func TestDatastoreStorer_GetEntities_err(t *testing.T) {
	params := NewDefaultParameters()
	lg := zap.NewNop()
	userID := "some user"
	cases := map[string]struct {
		s        *datastoreStorer
		userID   string
		expected error
	}{
		"empty user ID": {
			s:        &datastoreStorer{params: params, logger: lg},
			userID:   "",
			expected: api.ErrEmptyUserID,
		},
		"iter next err": {
			s: &datastoreStorer{
				params: params,
				logger: lg,
				client: &fixedDatastoreClient{},
				iter: &fixedDatastoreIter{
					err: errTest,
				},
			},
			userID:   userID,
			expected: errTest,
		},
	}
	for desc, c := range cases {
		entityIDs, err := c.s.GetEntities(c.userID)
		assert.Equal(t, c.expected, err, desc)
		assert.Nil(t, entityIDs)
	}
}

func TestDatastoreStorer_CountEntities_ok(t *testing.T) {
	params := NewDefaultParameters()
	lg := zap.NewNop()
	client := &fixedDatastoreClient{
		countValue: 2,
	}
	s := &datastoreStorer{
		params: params,
		client: client,
		logger: lg,
	}

	n, err := s.CountEntities("some user ID")
	assert.Nil(t, err)
	assert.Equal(t, client.countValue, n)
}

func TestDatastoreStorer_CountEntities_err(t *testing.T) {
	params := NewDefaultParameters()
	lg := zap.NewNop()
	userID := "some user"
	cases := map[string]struct {
		s        *datastoreStorer
		userID   string
		expected error
	}{
		"empty user ID": {
			s:        &datastoreStorer{params: params, logger: lg},
			userID:   "",
			expected: api.ErrEmptyUserID,
		},
		"client count err": {
			s: &datastoreStorer{
				params: params,
				logger: lg,
				client: &fixedDatastoreClient{
					countErr: errTest,
				},
			},
			userID:   userID,
			expected: errTest,
		},
	}
	for desc, c := range cases {
		n, err := c.s.CountEntities(c.userID)
		assert.Equal(t, c.expected, err, desc)
		assert.Zero(t, n)
	}
}

func TestDatastoreStorer_CountUsers_ok(t *testing.T) {
	params := NewDefaultParameters()
	lg := zap.NewNop()
	client := &fixedDatastoreClient{
		countValue: 2,
	}
	s := &datastoreStorer{
		params: params,
		client: client,
		logger: lg,
	}

	n, err := s.CountUsers("some entity ID")
	assert.Nil(t, err)
	assert.Equal(t, client.countValue, n)
}

func TestDatastoreStorer_CountUsers_err(t *testing.T) {
	params := NewDefaultParameters()
	lg := zap.NewNop()
	entityID := "some entity"
	cases := map[string]struct {
		s        *datastoreStorer
		entityID string
		expected error
	}{
		"empty entity ID": {
			s:        &datastoreStorer{params: params, logger: lg},
			entityID: "",
			expected: api.ErrEmptyEntityID,
		},
		"client count err": {
			s: &datastoreStorer{
				params: params,
				logger: lg,
				client: &fixedDatastoreClient{
					countErr: errTest,
				},
			},
			entityID: entityID,
			expected: errTest,
		},
	}
	for desc, c := range cases {
		n, err := c.s.CountUsers(c.entityID)
		assert.Equal(t, c.expected, err, desc)
		assert.Zero(t, n)
	}
}

type fixedDatastoreClient struct {
	putValues  []*UserEntity
	putErr     error
	countValue int
	countErr   error
}

func (f *fixedDatastoreClient) Put(
	ctx context.Context, key *datastore.Key, value interface{},
) (*datastore.Key, error) {
	f.putValues = append(f.putValues, value.(*UserEntity))
	return nil, f.putErr
}

func (f *fixedDatastoreClient) PutMulti(
	context.Context, []*datastore.Key, interface{},
) ([]*datastore.Key, error) {
	panic("implement me")
}

func (f *fixedDatastoreClient) Get(
	ctx context.Context, key *datastore.Key, dest interface{},
) error {
	panic("implement me")
}

func (f *fixedDatastoreClient) GetMulti(
	ctx context.Context, keys []*datastore.Key, dst interface{},
) error {
	panic("implement me")
}

func (f *fixedDatastoreClient) Delete(ctx context.Context, keys []*datastore.Key) error {
	panic("implement me")
}

func (f *fixedDatastoreClient) Count(ctx context.Context, q *datastore.Query) (int, error) {
	return f.countValue, f.countErr
}

func (f *fixedDatastoreClient) Run(ctx context.Context, q *datastore.Query) *datastore.Iterator {
	return nil
}

type fixedDatastoreIter struct {
	err    error
	keys   []*datastore.Key
	values []*UserEntity
	offset int
}

func (f *fixedDatastoreIter) Init(iter *datastore.Iterator) {}

func (f *fixedDatastoreIter) Next(dst interface{}) (*datastore.Key, error) {
	if f.err != nil {
		return nil, f.err
	}
	defer func() { f.offset++ }()
	if f.offset == len(f.values) {
		return nil, iterator.Done
	}
	v := f.values[f.offset]
	dst.(*UserEntity).EntityID = v.EntityID
	dst.(*UserEntity).UserID = v.UserID
	return f.keys[f.offset], nil
}
