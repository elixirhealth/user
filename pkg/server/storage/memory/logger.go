package memory

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	logEntityID  = "entity_id"
	logUserID    = "user_id"
	logNEntities = "n_entities"
	logCount     = "count"
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
