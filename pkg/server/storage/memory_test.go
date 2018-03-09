package storage

import (
	"testing"

	api "github.com/elxirhealth/user/pkg/userapi"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMemoryStorer_AddEntity_ok(t *testing.T) {
	s := NewMemory(NewDefaultParameters(), zap.NewNop())

	userID, entityID := "some user", "some entity"
	err := s.AddEntity(userID, entityID)
	assert.Nil(t, err)

	putValue := s.(*memoryStorer).userEntities[0]
	assert.Equal(t, userID, putValue.UserID)
	assert.Equal(t, entityID, putValue.EntityID)
	assert.False(t, putValue.Removed)
	assert.NotZero(t, putValue.ModifiedDate)
	assert.NotZero(t, putValue.ModifiedTime)
	assert.NotZero(t, putValue.AddedTime)
	assert.Zero(t, putValue.RemovedTime)
}

func TestMemoryStorer_AddEntity_err(t *testing.T) {
	params := NewDefaultParameters()
	lg := zap.NewNop()
	userID, entityID := "some user", "some entity"
	cases := map[string]struct {
		s        Storer
		userID   string
		entityID string
		expected error
	}{
		"empty user ID": {
			s:        NewMemory(params, lg),
			userID:   "",
			entityID: entityID,
			expected: api.ErrEmptyUserID,
		},
		"empty entity ID": {
			s:        NewMemory(params, lg),
			userID:   userID,
			entityID: "",
			expected: api.ErrEmptyEntityID,
		},
		"non-zero user entities count": {
			s: &memoryStorer{
				params: params,
				logger: lg,
				userEntities: []*UserEntity{
					{UserID: userID, EntityID: entityID},
				},
			},
			userID:   userID,
			entityID: entityID,
			expected: ErrUserEntityExists,
		},
	}
	for desc, c := range cases {
		err := c.s.AddEntity(c.userID, c.entityID)
		assert.Equal(t, c.expected, err, desc)
	}
}

func TestMemoryStorer_GetEntities_ok(t *testing.T) {
	params := NewDefaultParameters()
	lg := zap.NewNop()

	userID := "some user ID"
	entityID1, entityID2 := "some entity ID", "another entity ID"
	s := &memoryStorer{
		params: params,
		logger: lg,
		userEntities: []*UserEntity{
			{UserID: userID, EntityID: entityID1},
			{UserID: userID, EntityID: entityID2},
		},
	}

	entityIDs, err := s.GetEntities(userID)
	assert.Nil(t, err)
	assert.Equal(t, []string{entityID1, entityID2}, entityIDs)
}

func TestMemoryStorer_GetEntities_err(t *testing.T) {
	s := NewMemory(NewDefaultParameters(), zap.NewNop())
	entityIDs, err := s.GetEntities("")
	assert.Equal(t, api.ErrEmptyUserID, err)
	assert.Nil(t, entityIDs)
}

func TestMemoryStorer_CountEntities_ok(t *testing.T) {
	params := NewDefaultParameters()
	lg := zap.NewNop()

	userID1, userID2 := "some user ID", "another user ID"
	entityID1, entityID2 := "some entity ID", "another entity ID"
	s := &memoryStorer{
		params: params,
		logger: lg,
		userEntities: []*UserEntity{
			{UserID: userID1, EntityID: entityID1},
			{UserID: userID1, EntityID: entityID2},
			{UserID: userID2, EntityID: entityID1},
		},
	}

	n, err := s.CountEntities(userID1)
	assert.Nil(t, err)
	assert.Equal(t, 2, n)

	n, err = s.CountEntities(userID2)
	assert.Nil(t, err)
	assert.Equal(t, 1, n)
}

func TestMemoryStorer_CountEntities_err(t *testing.T) {
	s := NewMemory(NewDefaultParameters(), zap.NewNop())
	n, err := s.CountEntities("")
	assert.Equal(t, api.ErrEmptyUserID, err)
	assert.Zero(t, n)
}

func TestMemoryStorer_CountUsers_ok(t *testing.T) {
	params := NewDefaultParameters()
	lg := zap.NewNop()

	userID1, userID2 := "some user ID", "another user ID"
	entityID1, entityID2 := "some entity ID", "another entity ID"
	s := &memoryStorer{
		params: params,
		logger: lg,
		userEntities: []*UserEntity{
			{UserID: userID1, EntityID: entityID1},
			{UserID: userID1, EntityID: entityID2},
			{UserID: userID2, EntityID: entityID1},
		},
	}

	n, err := s.CountUsers(entityID1)
	assert.Nil(t, err)
	assert.Equal(t, 2, n)

	n, err = s.CountUsers(entityID2)
	assert.Nil(t, err)
	assert.Equal(t, 1, n)
}

func TestMemoryStorer_CountUsers_err(t *testing.T) {
	s := NewMemory(NewDefaultParameters(), zap.NewNop())
	n, err := s.CountUsers("")
	assert.Equal(t, api.ErrEmptyEntityID, err)
	assert.Zero(t, n)
}
