package postgres

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	errors2 "github.com/drausin/libri/libri/common/errors"
	bstorage "github.com/elixirhealth/service-base/pkg/server/storage"
	"github.com/elixirhealth/user/pkg/server"
	"github.com/elixirhealth/user/pkg/server/storage"
	api "github.com/elixirhealth/user/pkg/userapi"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	userSchema  = `"user"`
	entityTable = "entity"

	userIDCol   = "user_id"
	entityIDCol = "entity_id"

	count = "COUNT(*)"
)

var (
	psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	fqEntityTable = userSchema + "." + entityTable

	errEmptyDBUrl            = errors.New("empty DB URL")
	errUnexpectedStorageType = errors.New("unexpected storage type")
)

type storer struct {
	params  *storage.Parameters
	db      *sql.DB
	dbCache sq.DBProxyContext
	qr      bstorage.Querier
	logger  *zap.Logger
}

// New creates a new storage.Storer backed by a Postgres DB at the given dbURL.
func New(dbURL string, params *storage.Parameters, logger *zap.Logger) (storage.Storer, error) {
	if dbURL == "" {
		return nil, errEmptyDBUrl
	}
	if params.Type != bstorage.Postgres {
		return nil, errUnexpectedStorageType
	}
	db, err := sql.Open("postgres", dbURL)
	errors2.MaybePanic(err)
	return &storer{
		params:  params,
		db:      db,
		dbCache: sq.NewStmtCacher(db),
		qr:      bstorage.NewQuerier(),
		logger:  logger,
	}, nil
}

func (s *storer) AddEntity(userID, entityID string) error {
	if userID == "" {
		return api.ErrEmptyUserID
	}
	if entityID == "" {
		return api.ErrEmptyEntityID
	}
	q := psql.RunWith(s.dbCache).
		Insert(fqEntityTable).
		SetMap(getSQLValues(userID, entityID))
	s.logger.Debug("adding entity", logUserEntityID(userID, entityID)...)
	ctx, cancel := context.WithTimeout(context.Background(), s.params.AddQueryTimeout)
	defer cancel()
	_, err := s.qr.InsertExecContext(ctx, q)
	if err != nil {
		return err
	}
	s.logger.Debug("added entity", logUserEntityID(userID, entityID)...)
	return nil
}

func (s *storer) GetEntities(userID string) ([]string, error) {
	if userID == "" {
		return nil, api.ErrEmptyUserID
	}
	cols, _, _ := prepEntityScan()
	q := psql.RunWith(s.dbCache).
		Select(cols...).
		From(fqEntityTable).
		Where(sq.Eq{userIDCol: userID})
	s.logger.Debug("getting entities", logGettingEntities(q, userID)...)
	ctx, cancel := context.WithTimeout(context.Background(), s.params.GetQueryTimeout)
	defer cancel()
	rows, err := s.qr.SelectQueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	entityIDs := make([]string, 0, server.MaxUserEntities)
	for rows.Next() {
		_, dest, create := prepEntityScan()
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		entityIDs = append(entityIDs, create())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	s.logger.Debug("got entities", logGotEntities(userID, entityIDs)...)
	return entityIDs, nil
}

func (s *storer) CountEntities(userID string) (int, error) {
	if userID == "" {
		return 0, api.ErrEmptyUserID
	}
	return s.count(sq.Eq{userIDCol: userID}, "counted entities", zap.String(logUserID, userID))
}

func (s *storer) CountUsers(entityID string) (int, error) {
	if entityID == "" {
		return 0, api.ErrEmptyEntityID
	}
	return s.count(sq.Eq{entityIDCol: entityID}, "counted users",
		zap.String(logEntityID, entityID))
}

func (s storer) count(pred interface{}, logMsg string, fields ...zapcore.Field) (int, error) {
	q := psql.RunWith(s.dbCache).
		Select(count).
		From(fqEntityTable).
		Where(pred)
	ctx, cancel := context.WithTimeout(context.Background(), s.params.GetQueryTimeout)
	defer cancel()
	row := s.qr.SelectQueryRowContext(ctx, q)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	fields = append(fields, zap.Int(logCount, count))
	s.logger.Debug(logMsg, fields...)
	return count, nil
}

func getSQLValues(userID, entityID string) map[string]interface{} {
	return map[string]interface{}{
		userIDCol:   userID,
		entityIDCol: entityID,
	}
}

func prepEntityScan() ([]string, []interface{}, func() string) {
	var entityID string
	cols, dests := bstorage.SplitColDests(0, []*bstorage.ColDest{
		{entityIDCol, &entityID},
	})
	return cols, dests, func() string {
		entityID = *dests[0].(*string)
		return entityID
	}
}
