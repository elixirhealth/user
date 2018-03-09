package storage

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	logType            = "type"
	logEntityID        = "entity_id"
	logUserID          = "user_id"
	logNEntities       = "n_entities"
	logCount           = "count"
	logAddQueryTimeout = "add_query_timeout"
	logGetQueryTimeout = "get_query_timeout"
)

func logAddEntityFields(userID, entityID string) []zapcore.Field {
	return []zapcore.Field{
		zap.String(logEntityID, entityID),
		zap.String(logUserID, userID),
	}
}

func logGetEntities(userID string, entityIDs []string) []zapcore.Field {
	return []zapcore.Field{
		zap.String(logUserID, userID),
		zap.Int(logNEntities, len(entityIDs)),
	}
}

func logCountEntities(userID string, n int) []zapcore.Field {
	return []zapcore.Field{
		zap.String(logUserID, userID),
		zap.Int(logCount, n),
	}
}

func logCountUsers(entityID string, n int) []zapcore.Field {
	return []zapcore.Field{
		zap.String(logEntityID, entityID),
		zap.Int(logCount, n),
	}
}

func logCountUserEntities(userID, entityID string, n int) []zapcore.Field {
	return []zapcore.Field{
		zap.String(logUserID, userID),
		zap.String(logEntityID, entityID),
		zap.Int(logCount, n),
	}
}
