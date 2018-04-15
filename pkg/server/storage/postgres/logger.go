package postgres

import (
	sq "github.com/Masterminds/squirrel"
	errors2 "github.com/drausin/libri/libri/common/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	logEntityID  = "entity_id"
	logUserID    = "user_id"
	logNEntities = "n_entities"
	logSQL       = "sql"
	logArgs      = "args"
	logCount     = "count"
)

func logUserEntityID(userID, entityID string) []zapcore.Field {
	return []zapcore.Field{
		zap.String(logEntityID, entityID),
		zap.String(logUserID, userID),
	}
}

func logGettingEntities(q sq.SelectBuilder, userID string) []zapcore.Field {
	qSQL, args, err := q.ToSql()
	errors2.MaybePanic(err)
	return []zapcore.Field{
		zap.String(logUserID, userID),
		zap.String(logSQL, qSQL),
		zap.Array(logArgs, queryArgs(args)),
	}
}

func logGotEntities(userID string, entityIDs []string) []zapcore.Field {
	return []zapcore.Field{
		zap.String(logUserID, userID),
		zap.Int(logNEntities, len(entityIDs)),
	}
}

type queryArgs []interface{}

func (qas queryArgs) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, qa := range qas {
		switch val := qa.(type) {
		case string:
			enc.AppendString(val)
		default:
			if err := enc.AppendReflected(qa); err != nil {
				return err
			}
		}
	}
	return nil
}
