package cmd

import (
	"testing"

	"github.com/elixirhealth/service-base/pkg/cmd"
	bstorage "github.com/elixirhealth/service-base/pkg/server/storage"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestGetUserConfig(t *testing.T) {
	serverPort := uint(1234)
	metricsPort := uint(5678)
	profilerPort := uint(9012)
	logLevel := zapcore.DebugLevel.String()
	profile := true
	dbURL := "some DB URL"
	storageInMemory := false
	storagePostgres := true

	viper.Set(cmd.ServerPortFlag, serverPort)
	viper.Set(cmd.MetricsPortFlag, metricsPort)
	viper.Set(cmd.ProfilerPortFlag, profilerPort)
	viper.Set(cmd.LogLevelFlag, logLevel)
	viper.Set(cmd.ProfileFlag, profile)
	viper.Set(dbURLFlag, dbURL)
	viper.Set(storageMemoryFlag, storageInMemory)
	viper.Set(storagePostgresFlag, storagePostgres)

	c, err := getUserConfig()
	assert.Nil(t, err)
	assert.Equal(t, serverPort, c.ServerPort)
	assert.Equal(t, metricsPort, c.MetricsPort)
	assert.Equal(t, profilerPort, c.ProfilerPort)
	assert.Equal(t, logLevel, c.LogLevel.String())
	assert.Equal(t, profile, c.Profile)
	assert.Equal(t, dbURL, c.DBUrl)
	assert.Equal(t, bstorage.Postgres, c.Storage.Type)
}
