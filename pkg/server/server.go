package server

import (
	"github.com/elxirhealth/service-base/pkg/server"
	"github.com/elxirhealth/user/pkg/server/storage"
)

// User implements the UserServer interface.
type User struct {
	*server.BaseServer
	config *Config

	storer storage.Storer
	// TODO maybe add other things here
}

// newUser creates a new UserServer from the given config.
func newUser(config *Config) (*User, error) {
	baseServer := server.NewBaseServer(config.BaseConfig)
	storer, err := getStorer(config, baseServer.Logger)
	if err != nil {
		return nil, err
	}
	// TODO maybe add other init

	return &User{
		BaseServer: baseServer,
		config:     config,
		storer:     storer,
		// TODO maybe add other things
	}, nil
}

// TODO implement userapi.User endpoints
