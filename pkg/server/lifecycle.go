package server

import (
	errors2 "github.com/drausin/libri/libri/common/errors"
	bstorage "github.com/elixirhealth/service-base/pkg/server/storage"
	"github.com/elixirhealth/user/pkg/server/storage/postgres/migrations"
	api "github.com/elixirhealth/user/pkg/userapi"
	bindata "github.com/mattes/migrate/source/go-bindata"
	"google.golang.org/grpc"
)

// Start starts the server and eviction routines.
func Start(config *Config, up chan *User) error {
	c, err := newUser(config)
	if err != nil {
		return err
	}

	if err := c.maybeMigrateDB(); err != nil {
		return err
	}

	registerServer := func(s *grpc.Server) { api.RegisterUserServer(s, c) }
	return c.Serve(registerServer, func() { up <- c })
}

// StopServer handles cleanup involved in closing down the server.
func (u *User) StopServer() {
	u.BaseServer.StopServer()
	err := u.storer.Close()
	errors2.MaybePanic(err)
}

func (u *User) maybeMigrateDB() error {
	if u.config.Storage.Type != bstorage.Postgres {
		return nil
	}

	m := bstorage.NewBindataMigrator(
		u.config.DBUrl,
		bindata.Resource(migrations.AssetNames(), migrations.Asset),
		&bstorage.ZapLogger{Logger: u.Logger},
	)
	return m.Up()
}
