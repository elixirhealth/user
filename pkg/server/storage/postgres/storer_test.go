package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	sq "github.com/Masterminds/squirrel"
	errors2 "github.com/drausin/libri/libri/common/errors"
	"github.com/drausin/libri/libri/common/logging"
	bstorage "github.com/elixirhealth/service-base/pkg/server/storage"
	"github.com/elixirhealth/user/pkg/server/storage"
	"github.com/elixirhealth/user/pkg/server/storage/postgres/migrations"
	api "github.com/elixirhealth/user/pkg/userapi"
	"github.com/mattes/migrate/source/go-bindata"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

var (
	errTest = errors.New("test error")
)

func setUpPostgresTest() (string, func() error) {
	dbURL, cleanup, err := bstorage.StartTestPostgres()
	errors2.MaybePanic(err)
	as := bindata.Resource(migrations.AssetNames(), migrations.Asset)
	logger := &bstorage.LogLogger{}
	m := bstorage.NewBindataMigrator(dbURL, as, logger)
	errors2.MaybePanic(m.Up())
	return dbURL, func() error {
		if err := m.Down(); err != nil {
			return err
		}
		return cleanup()
	}
}

func TestStorer_AddGetCount(t *testing.T) {
	dbURL, tearDown := setUpPostgresTest()
	defer func() {
		err := tearDown()
		assert.Nil(t, err)
	}()

	params := storage.NewDefaultParameters()
	params.Type = bstorage.Postgres
	lg := logging.NewDevLogger(zapcore.DebugLevel) // zap.NewNop()
	userID1, userID2 := "user ID 1", "user ID 2"
	entityID1, entityID2, entityID3 := "entity ID 1", "entity ID 2", "entity ID 3"

	s, err := New(dbURL, params, lg)
	assert.Nil(t, err)

	err = s.AddEntity(userID1, entityID1)
	assert.Nil(t, err)
	err = s.AddEntity(userID1, entityID2)
	assert.Nil(t, err)
	err = s.AddEntity(userID2, entityID3)
	assert.Nil(t, err)

	entityIDs, err := s.GetEntities(userID1)
	assert.Nil(t, err)
	assert.Equal(t, []string{entityID1, entityID2}, entityIDs)

	nEntities, err := s.CountEntities(userID1)
	assert.Nil(t, err)
	assert.Equal(t, 2, nEntities)

	nEntities, err = s.CountUsers(entityID2)
	assert.Nil(t, err)
	assert.Equal(t, 1, nEntities)

	nUsers, err := s.CountUsers(entityID1)
	assert.Nil(t, err)
	assert.Equal(t, 1, nUsers)
}

func TestStorer_AddEntity_err(t *testing.T) {
	userID, entityID := "some user ID", "some entity ID"
	params := storage.NewDefaultParameters()
	params.Type = bstorage.Postgres
	lg := logging.NewDevLogger(zapcore.DebugLevel) // zap.NewNop()

	cases := map[string]struct {
		userID   string
		entityID string
		s        *storer
		expected error
	}{
		"bad user ID": {
			userID:   "",
			entityID: entityID,
			s:        &storer{},
			expected: api.ErrEmptyUserID,
		},
		"bad entity ID": {
			userID:   userID,
			entityID: "",
			s:        &storer{},
			expected: api.ErrEmptyEntityID,
		},
		"insert err": {
			userID:   userID,
			entityID: entityID,
			s: &storer{
				params: params,
				logger: lg,
				qr: &fixedQuerier{
					insertErr: errTest,
				},
			},
			expected: errTest,
		},
	}
	for desc, c := range cases {
		err := c.s.AddEntity(c.userID, c.entityID)
		assert.Equal(t, c.expected, err, desc)
	}
}

func TestStorer_GetEntities_err(t *testing.T) {
	userID := "some user ID"
	params := storage.NewDefaultParameters()
	params.Type = bstorage.Postgres
	lg := logging.NewDevLogger(zapcore.DebugLevel) // zap.NewNop()

	cases := map[string]struct {
		userID   string
		s        *storer
		expected error
	}{
		"bad user ID": {
			userID:   "",
			s:        &storer{},
			expected: api.ErrEmptyUserID,
		},
		"select err": {
			userID: userID,
			s: &storer{
				params: params,
				logger: lg,
				qr: &fixedQuerier{
					selectErr: errTest,
				},
			},
			expected: errTest,
		},
		"scan err": {
			userID: userID,
			s: &storer{
				params: params,
				logger: lg,
				qr: &fixedQuerier{
					selectResult: &fixedRowScanner{
						next:    true,
						scanErr: errTest,
					},
				},
			},
			expected: errTest,
		},
		"err err": {
			userID: userID,
			s: &storer{
				params: params,
				logger: lg,
				qr: &fixedQuerier{
					selectResult: &fixedRowScanner{
						errErr: errTest,
					},
				},
			},
			expected: errTest,
		},
	}
	for desc, c := range cases {
		entityIDs, err := c.s.GetEntities(c.userID)
		assert.Equal(t, c.expected, err, desc)
		assert.Nil(t, entityIDs)
	}
}

func TestStorer_CountUsers_err(t *testing.T) {
	userID := "some user ID"
	params := storage.NewDefaultParameters()
	params.Type = bstorage.Postgres
	lg := logging.NewDevLogger(zapcore.DebugLevel) // zap.NewNop()

	cases := map[string]struct {
		userID   string
		s        *storer
		expected error
	}{
		"bad user ID": {
			userID:   "",
			s:        &storer{},
			expected: api.ErrEmptyUserID,
		},
		"select row scan err": {
			userID: userID,
			s: &storer{
				params: params,
				logger: lg,
				qr: &fixedQuerier{
					selectRowResult: &fixedRowScanner{
						scanErr: errTest,
					},
				},
			},
			expected: errTest,
		},
	}
	for desc, c := range cases {
		count, err := c.s.CountEntities(c.userID)
		assert.Equal(t, c.expected, err, desc)
		assert.Zero(t, count, desc)
	}
}

func TestStorer_CountEntities_err(t *testing.T) {
	entityID := "some entity ID"
	params := storage.NewDefaultParameters()
	params.Type = bstorage.Postgres
	lg := logging.NewDevLogger(zapcore.DebugLevel) // zap.NewNop()

	cases := map[string]struct {
		entityID string
		s        *storer
		expected error
	}{
		"bad entity ID": {
			entityID: "",
			s:        &storer{},
			expected: api.ErrEmptyEntityID,
		},
		"select row scan err": {
			entityID: entityID,
			s: &storer{
				params: params,
				logger: lg,
				qr: &fixedQuerier{
					selectRowResult: &fixedRowScanner{
						scanErr: errTest,
					},
				},
			},
			expected: errTest,
		},
	}
	for desc, c := range cases {
		count, err := c.s.CountUsers(c.entityID)
		assert.Equal(t, c.expected, err, desc)
		assert.Zero(t, count, desc)
	}
}

type fixedQuerier struct {
	selectResult    bstorage.QueryRows
	selectErr       error
	selectRowResult sq.RowScanner
	insertResult    sql.Result
	insertErr       error
}

func (f *fixedQuerier) SelectQueryContext(
	ctx context.Context, b sq.SelectBuilder,
) (bstorage.QueryRows, error) {
	return f.selectResult, f.selectErr
}

func (f *fixedQuerier) SelectQueryRowContext(
	ctx context.Context, b sq.SelectBuilder,
) sq.RowScanner {
	return f.selectRowResult
}

func (f *fixedQuerier) InsertExecContext(
	ctx context.Context, b sq.InsertBuilder,
) (sql.Result, error) {
	return f.insertResult, f.insertErr
}

func (f *fixedQuerier) UpdateExecContext(
	ctx context.Context, b sq.UpdateBuilder,
) (sql.Result, error) {
	panic("implement me")
}

func (f *fixedQuerier) DeleteExecContext(
	ctx context.Context, b sq.DeleteBuilder,
) (sql.Result, error) {
	panic("implement me")
}

type fixedRowScanner struct {
	next    bool
	scanErr error
	errErr  error
}

func (f *fixedRowScanner) Next() bool {
	return f.next
}

func (f *fixedRowScanner) Close() error {
	panic("implement me")
}

func (f *fixedRowScanner) Err() error {
	return f.errErr
}

func (f *fixedRowScanner) Scan(...interface{}) error {
	return f.scanErr
}
