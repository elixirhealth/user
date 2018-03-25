package server

import (
	api "github.com/elixirhealth/user/pkg/userapi"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	logStorage   = "storage"
	logEntityID  = "entity_id"
	logUserID    = "user_id"
	logNEntities = "n_entities"
)

func logAddEntityRq(rq *api.AddEntityRequest) []zapcore.Field {
	return []zapcore.Field{
		zap.String(logUserID, rq.UserId),
		zap.String(logEntityID, rq.EntityId),
	}
}

func logGetEntityRp(rq *api.GetEntitiesRequest, rp *api.GetEntitiesResponse) []zapcore.Field {
	return []zapcore.Field{
		zap.String(logUserID, rq.UserId),
		zap.Int(logNEntities, len(rp.EntityIds)),
	}
}
