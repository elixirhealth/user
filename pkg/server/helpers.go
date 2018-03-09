package server

import (
	"errors"

	bstorage "github.com/elxirhealth/service-base/pkg/server/storage"
	"github.com/elxirhealth/user/pkg/server/storage"
	"go.uber.org/zap"
)

var (
	// ErrInvalidStorageType indicates when a storage type is not expected.
	ErrInvalidStorageType = errors.New("invalid storage type")
)

func getStorer(config *Config, logger *zap.Logger) (storage.Storer, error) {
	switch config.Storage.Type {
	case bstorage.Memory:
		return storage.NewMemory(config.Storage, logger), nil
	case bstorage.DataStore:
		return storage.NewDatastore(config.GCPProjectID, config.Storage, logger)
	default:
		return nil, ErrInvalidStorageType
	}
}
