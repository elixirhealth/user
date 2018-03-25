package server

import (
	api "github.com/elixirhealth/user/pkg/userapi"
	"google.golang.org/grpc"
)

// Start starts the server and eviction routines.
func Start(config *Config, up chan *User) error {
	c, err := newUser(config)
	if err != nil {
		return err
	}

	// start User aux routines
	// TODO add go x.auxRoutine() or delete comment

	registerServer := func(s *grpc.Server) { api.RegisterUserServer(s, c) }
	return c.Serve(registerServer, func() { up <- c })
}
