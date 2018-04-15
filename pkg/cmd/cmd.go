package cmd

import (
	"errors"
	"fmt"
	"log"

	cerrors "github.com/drausin/libri/libri/common/errors"
	"github.com/drausin/libri/libri/common/logging"
	"github.com/elixirhealth/service-base/pkg/cmd"
	bserver "github.com/elixirhealth/service-base/pkg/server"
	bstorage "github.com/elixirhealth/service-base/pkg/server/storage"
	"github.com/elixirhealth/user/pkg/server"
	"github.com/elixirhealth/user/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	serviceNameLower = "user"
	serviceNameCamel = "User"
	envVarPrefix     = "USER"

	logLevelFlag        = "logLevel"
	storageMemoryFlag   = "storageMemory"
	dbURLFlag           = "dbURL"
	dbPasswordFlag      = "dbPassword"
	storagePostgresFlag = "storagePostgres"
)

var (
	errMultipleStorageTypes = errors.New("multiple storage types specified")
	errNoStorageType        = errors.New("no storage type specified")

	rootCmd = &cobra.Command{
		Short: "operate a User server",
	}
)

func init() {
	rootCmd.PersistentFlags().String(logLevelFlag, bserver.DefaultLogLevel.String(),
		"log level")

	cmd.Start(serviceNameLower, serviceNameCamel, rootCmd, version.Current, start,
		func(flags *pflag.FlagSet) {
			flags.Bool(storageMemoryFlag, false, "use in-memory storage")
			flags.Bool(storagePostgresFlag, false, "use Postgres DB storage")
			flags.String(dbURLFlag, "", "Postgres DB URL, including username")
			flags.String(dbPasswordFlag, "", "DB user's password")
		})

	testCmd := cmd.Test(serviceNameLower, rootCmd)
	cmd.TestHealth(serviceNameLower, testCmd)
	cmd.TestIO(serviceNameLower, testCmd, testIO, func(flags *pflag.FlagSet) {
		// add additional test flags here if needed
	})

	cmd.Version(serviceNameLower, rootCmd, version.Current)

	// bind viper flags
	viper.SetEnvPrefix(envVarPrefix) // look for env vars with prefix
	viper.AutomaticEnv()             // read in environment variables that match
	cerrors.MaybePanic(viper.BindPFlags(rootCmd.Flags()))
}

// Execute runs the root user command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func start() error {
	config, err := getUserConfig()
	if err != nil {
		return err
	}
	return server.Start(config, make(chan *server.User, 1))
}

func getUserConfig() (*server.Config, error) {
	c := server.NewDefaultConfig()
	c.WithServerPort(uint(viper.GetInt(cmd.ServerPortFlag))).
		WithMetricsPort(uint(viper.GetInt(cmd.MetricsPortFlag))).
		WithProfilerPort(uint(viper.GetInt(cmd.ProfilerPortFlag))).
		WithLogLevel(logging.GetLogLevel(viper.GetString(logLevelFlag))).
		WithProfile(viper.GetBool(cmd.ProfileFlag))
	st, err := getStorageType()
	if err != nil {
		return nil, err
	}
	c.Storage.Type = st
	c.DBUrl = getDBUrl()
	return c, nil
}

func getDBUrl() string {
	dbURL := viper.GetString(dbURLFlag)
	if dbPass := viper.GetString(dbPasswordFlag); dbPass != "" {
		// append pw to URL args
		return fmt.Sprintf("%s&password=%s", dbURL, dbPass)
	}
	return dbURL
}

func getStorageType() (bstorage.Type, error) {
	if viper.GetBool(storageMemoryFlag) && viper.GetBool(storagePostgresFlag) {
		return bstorage.Unspecified, errMultipleStorageTypes
	}
	if viper.GetBool(storageMemoryFlag) {
		return bstorage.Memory, nil
	}
	if viper.GetBool(storagePostgresFlag) {
		return bstorage.Postgres, nil
	}
	return bstorage.Unspecified, errNoStorageType
}
