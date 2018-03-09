package server

import (
	"context"
	"testing"

	bserver "github.com/elxirhealth/service-base/pkg/server"
	bstorage "github.com/elxirhealth/service-base/pkg/server/storage"
	"github.com/elxirhealth/user/pkg/server/storage"
	api "github.com/elxirhealth/user/pkg/userapi"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	testEntityID = "some entity ID"
	testUserID   = "some user ID"
)

var (
	errTest = errors.New("some test error")
)

func TestNewUser_ok(t *testing.T) {
	config := NewDefaultConfig()
	c, err := newUser(config)
	assert.Nil(t, err)
	assert.Equal(t, config, c.config)
	assert.NotEmpty(t, c.storer)
}

func TestNewUser_err(t *testing.T) {
	badConfigs := map[string]*Config{
		"empty ProjectID": NewDefaultConfig().WithStorage(
			&storage.Parameters{Type: bstorage.DataStore},
		),
	}
	for desc, badConfig := range badConfigs {
		c, err := newUser(badConfig)
		assert.NotNil(t, err, desc)
		assert.Nil(t, c)
	}
}

func TestUser_AddEntity_ok(t *testing.T) {
	u := &User{
		BaseServer: bserver.NewBaseServer(bserver.NewDefaultBaseConfig()),
		storer:     &fixedStorer{},
	}
	rq := &api.AddEntityRequest{
		UserId:   testUserID,
		EntityId: testEntityID,
	}
	_, err := u.AddEntity(context.Background(), rq)
	assert.Nil(t, err)
}

func TestUser_AddEntity_err(t *testing.T) {
	baseServer := bserver.NewBaseServer(bserver.NewDefaultBaseConfig())
	okRq := &api.AddEntityRequest{
		EntityId: testEntityID,
		UserId:   testUserID,
	}
	cases := map[string]struct {
		u        *User
		rq       *api.AddEntityRequest
		expected error
	}{
		"bad rq": {
			u: &User{BaseServer: baseServer},
			rq: &api.AddEntityRequest{
				EntityId: testEntityID,
			},
			expected: api.ErrEmptyUserID,
		},
		"count users err": {
			u: &User{
				BaseServer: baseServer,
				storer: &fixedStorer{
					countUsersErr: errTest,
				},
			},
			rq:       okRq,
			expected: errTest,
		},
		"too many users for entity": {
			u: &User{
				BaseServer: baseServer,
				storer: &fixedStorer{
					countUsersValue: MaxEntityUsers,
				},
			},
			rq:       okRq,
			expected: ErrTooManyEntityUsers,
		},
		"too many entities for user": {
			u: &User{
				BaseServer: baseServer,
				storer: &fixedStorer{
					countEntitiesValue: MaxUserEntities,
				},
			},
			rq:       okRq,
			expected: ErrTooManyUserEntities,
		},
		"add entity err": {
			u: &User{
				BaseServer: baseServer,
				storer: &fixedStorer{
					addEntityErr: errTest,
				},
			},
			rq:       okRq,
			expected: errTest,
		},
	}
	for desc, c := range cases {
		_, err := c.u.AddEntity(context.Background(), c.rq)
		assert.Equal(t, c.expected, err, desc)
	}
}

func TestUser_GetEntities_ok(t *testing.T) {
	u := &User{
		BaseServer: bserver.NewBaseServer(bserver.NewDefaultBaseConfig()),
		storer: &fixedStorer{
			getEntitiesValue: []string{"entity ID 1", "entity ID 2"},
		},
	}
	rq := &api.GetEntitiesRequest{
		UserId: testUserID,
	}
	rp, err := u.GetEntities(context.Background(), rq)
	assert.Nil(t, err)
	assert.NotZero(t, len(rp.EntityIds))
}

func TestUser_GetEntities_err(t *testing.T) {
	baseServer := bserver.NewBaseServer(bserver.NewDefaultBaseConfig())
	cases := map[string]struct {
		u        *User
		rq       *api.GetEntitiesRequest
		expected error
	}{
		"bad rq": {
			u:        &User{BaseServer: baseServer},
			rq:       &api.GetEntitiesRequest{},
			expected: api.ErrEmptyUserID,
		},
		"add entity err": {
			u: &User{
				BaseServer: baseServer,
				storer: &fixedStorer{
					getEntitiesErr: errTest,
				},
			},
			rq: &api.GetEntitiesRequest{
				UserId: testUserID,
			},
			expected: errTest,
		},
	}
	for desc, c := range cases {
		_, err := c.u.GetEntities(context.Background(), c.rq)
		assert.Equal(t, c.expected, err, desc)
	}
}

type fixedStorer struct {
	addEntityErr       error
	getEntitiesValue   []string
	getEntitiesErr     error
	countEntitiesValue int
	countEntitiesErr   error
	countUsersValue    int
	countUsersErr      error
}

func (f *fixedStorer) AddEntity(userID, entityID string) error {
	return f.addEntityErr
}

func (f *fixedStorer) GetEntities(userID string) ([]string, error) {
	return f.getEntitiesValue, f.getEntitiesErr
}

func (f *fixedStorer) CountEntities(userID string) (int, error) {
	return f.countEntitiesValue, f.countEntitiesErr
}

func (f *fixedStorer) CountUsers(entityID string) (int, error) {
	return f.countUsersValue, f.countUsersErr
}
