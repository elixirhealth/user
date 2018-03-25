package client

import (
	api "github.com/elixirhealth/user/pkg/userapi"
	"google.golang.org/grpc"
)

// NewInsecure returns a new UserClient without any TLS on the connection.
func NewInsecure(address string) (api.UserClient, error) {
	cc, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return api.NewUserClient(cc), nil
}
