package server

import (
	"github.com/drausin/libri/libri/common/errors"
	"github.com/elixirhealth/service-base/pkg/server"
	"github.com/elixirhealth/user/pkg/server/storage"
	"go.uber.org/zap/zapcore"
)

// Config is the config for a User instance.
type Config struct {
	*server.BaseConfig
	Storage      *storage.Parameters
	GCPProjectID string
}

// NewDefaultConfig create a new config instance with default values.
func NewDefaultConfig() *Config {
	config := &Config{
		BaseConfig: server.NewDefaultBaseConfig(),
	}
	return config.
		WithDefaultStorage()
}

// MarshalLogObject writes the config to the given object encoder.
func (c *Config) MarshalLogObject(oe zapcore.ObjectEncoder) error {
	err := c.BaseConfig.MarshalLogObject(oe)
	errors.MaybePanic(err) // should never happen
	err = oe.AddObject(logStorage, c.Storage)
	errors.MaybePanic(err) // should never happen
	return nil
}

// WithStorage sets the cache parameters to the given value or the defaults if it is nil.
func (c *Config) WithStorage(p *storage.Parameters) *Config {
	if p == nil {
		return c.WithDefaultStorage()
	}
	c.Storage = p
	return c
}

// WithDefaultStorage set the Cache parameters to their default values.
func (c *Config) WithDefaultStorage() *Config {
	c.Storage = storage.NewDefaultParameters()
	return c
}

// WithGCPProjectID sets the GCP ProjectID to the given value.
func (c *Config) WithGCPProjectID(id string) *Config {
	c.GCPProjectID = id
	return c
}
