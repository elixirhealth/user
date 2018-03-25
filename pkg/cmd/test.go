package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/drausin/libri/libri/common/logging"
	"github.com/drausin/libri/libri/common/parse"
	"github.com/elixirhealth/service-base/pkg/cmd"
	"github.com/elixirhealth/service-base/pkg/server"
	"github.com/elixirhealth/user/pkg/userapi"
	api "github.com/elixirhealth/user/pkg/userapi"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	timeoutFlag = "timeout"

	logEntityID  = "entity_id"
	logUserID    = "user_id"
	logNEntities = "n_entities"
)

func testIO() error {
	rng := rand.New(rand.NewSource(0))
	logger := logging.NewDevLogger(logging.GetLogLevel(viper.GetString(logLevelFlag)))
	timeout := time.Duration(viper.GetInt(timeoutFlag) * 1e9)
	nUserIDs := 4
	nEntityIDs := 32
	maxUserEntities := 8

	userEntities := make(map[string]map[string]struct{})

	clients, err := getClients()
	if err != nil {
		return err
	}

	// add entities
	for c := 0; c < nUserIDs; c++ {
		userID := getUserID(c)
		userEntities[userID] = make(map[string]struct{})
		nEntities := rng.Intn(maxUserEntities) + 1
		for len(userEntities[userID]) < nEntities {
			i := rng.Intn(nEntityIDs)
			entityID := getEntityID(i)
			if _, in := userEntities[userID][entityID]; in {
				continue
			}

			rq := &api.AddEntityRequest{
				UserId:   userID,
				EntityId: entityID,
			}
			client := clients[rng.Int31n(int32(len(clients)))]
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			_, err := client.AddEntity(ctx, rq)
			cancel()
			if err2 := logAddEntityKeysRp(logger, rq, err); err2 != nil {
				return err2
			}
			userEntities[userID][entityID] = struct{}{}
		}
	}

	// get entities
	for c := 0; c < nUserIDs; c++ {
		userID := getUserID(c)
		rq := &api.GetEntitiesRequest{
			UserId: userID,
		}
		client := clients[rng.Int31n(int32(len(clients)))]
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		rp, err := client.GetEntities(ctx, rq)
		cancel()
		if err2 := logGetEntitiesRp(logger, rq, rp, err); err2 != nil {
			return err2
		}
	}

	return nil
}

func logAddEntityKeysRp(logger *zap.Logger, rq *api.AddEntityRequest, err error) error {
	if err != nil {
		logger.Error("adding entity failed", zap.Error(err))
		return err
	}
	logger.Info("added entity",
		zap.String(logEntityID, rq.EntityId),
		zap.String(logUserID, rq.UserId),
	)
	return nil
}

func logGetEntitiesRp(
	logger *zap.Logger, rq *api.GetEntitiesRequest, rp *api.GetEntitiesResponse, err error,
) error {
	if err != nil {
		logger.Error("getting entities failed", zap.Error(err))
		return err
	}
	logger.Info("got user entities",
		zap.String(logUserID, rq.UserId),
		zap.Int(logNEntities, len(rp.EntityIds)),
	)
	return nil
}

func getUserID(i int) string {
	return fmt.Sprintf("User-%d", i)
}

func getEntityID(i int) string {
	return fmt.Sprintf("Entity-%d", i)
}

func getClients() ([]userapi.UserClient, error) {
	addrs, err := parse.Addrs(viper.GetStringSlice(cmd.AddressesFlag))
	if err != nil {
		return nil, err
	}
	dialer := server.NewInsecureDialer()
	clients := make([]userapi.UserClient, len(addrs))
	for i, addr := range addrs {
		conn, err2 := dialer.Dial(addr.String())
		if err != nil {
			return nil, err2
		}
		clients[i] = userapi.NewUserClient(conn)
	}
	return clients, nil
}
