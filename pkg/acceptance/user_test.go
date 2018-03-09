// +build acceptance

package acceptance

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"testing"
	"time"

	"github.com/drausin/libri/libri/common/errors"
	bstorage "github.com/elxirhealth/service-base/pkg/server/storage"
	"github.com/elxirhealth/user/pkg/server"
	"github.com/elxirhealth/user/pkg/server/storage"
	api "github.com/elxirhealth/user/pkg/userapi"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

type parameters struct {
	nUsers       uint
	gcpProjectID string
	logLevel     zapcore.Level
	timeout      time.Duration

	nUserIDs        uint
	nEntityIDs      uint
	maxUserEntities uint
}

type state struct {
	users         []*server.User
	userClients   []api.UserClient
	dataDir       string
	datastoreProc *os.Process
	rng           *rand.Rand

	userEntities map[string]map[string]struct{}
}

func (st *state) randClient() api.UserClient {
	return st.userClients[st.rng.Intn(len(st.userClients))]
}

func TestAcceptance(t *testing.T) {
	params := &parameters{
		nUsers:       3,
		gcpProjectID: "dummy-acceptance-id",
		logLevel:     zapcore.InfoLevel,
		timeout:      1 * time.Second,

		nUserIDs:        16,
		nEntityIDs:      64,
		maxUserEntities: 8,
	}
	st := setUp(params)

	// add a bunch of (user, entity) pairs
	testAddEntity(t, params, st)

	// get entities for each user
	testGetEntities(t, params, st)

	tearDown(st)
}

func testAddEntity(t *testing.T, params *parameters, st *state) {
	for c := 0; c < int(params.nUserIDs); c++ {
		userID := getUserID(c)
		st.userEntities[userID] = make(map[string]struct{})
		nEntities := st.rng.Intn(int(params.maxUserEntities)) + 1
		for len(st.userEntities[userID]) < nEntities {
			i := st.rng.Intn(int(params.nEntityIDs))
			entityID := getEntityID(i)
			if _, in := st.userEntities[userID][entityID]; in {
				continue
			}

			rq := &api.AddEntityRequest{
				UserId:   userID,
				EntityId: entityID,
			}
			ctx, cancel := context.WithTimeout(context.Background(), params.timeout)
			_, err := st.randClient().AddEntity(ctx, rq)
			cancel()
			assert.Nil(t, err)

			st.userEntities[userID][entityID] = struct{}{}
		}
	}
}

func testGetEntities(t *testing.T, params *parameters, st *state) {
	for c := 0; c < int(params.nUserIDs); c++ {
		userID := getUserID(c)
		rq := &api.GetEntitiesRequest{
			UserId: userID,
		}
		ctx, cancel := context.WithTimeout(context.Background(), params.timeout)
		rp, err := st.randClient().GetEntities(ctx, rq)
		cancel()
		assert.Nil(t, err)

		rpEntityIDs := make(map[string]struct{})
		for _, entityID := range rp.EntityIds {
			rpEntityIDs[entityID] = struct{}{}
		}
		assert.Equal(t, st.userEntities[userID], rpEntityIDs)
	}
}

func getUserID(i int) string {
	return fmt.Sprintf("User-%d", i)
}

func getEntityID(i int) string {
	return fmt.Sprintf("Entity-%d", i)
}

func setUp(params *parameters) *state {
	rng := rand.New(rand.NewSource(0))

	dataDir, err := ioutil.TempDir("", "user-datastore-test")
	errors.MaybePanic(err)
	datastoreProc := bstorage.StartDatastoreEmulator(dataDir)

	time.Sleep(5 * time.Second)
	st := &state{
		rng:           rng,
		dataDir:       dataDir,
		datastoreProc: datastoreProc,
		userEntities:  make(map[string]map[string]struct{}),
	}
	createAndStartUsers(params, st)
	return st
}

func createAndStartUsers(params *parameters, st *state) {
	configs, addrs := newUserConfigs(params)
	users := make([]*server.User, params.nUsers)
	userClients := make([]api.UserClient, params.nUsers)
	up := make(chan *server.User, 1)

	for i := uint(0); i < params.nUsers; i++ {
		go func() {
			err := server.Start(configs[i], up)
			errors.MaybePanic(err)
		}()

		// wait for server to come up
		users[i] = <-up

		// set up client to it
		conn, err := grpc.Dial(addrs[i].String(), grpc.WithInsecure())
		errors.MaybePanic(err)
		userClients[i] = api.NewUserClient(conn)
	}

	st.users = users
	st.userClients = userClients
}

func newUserConfigs(params *parameters) ([]*server.Config, []*net.TCPAddr) {
	startPort := uint(10100)
	configs := make([]*server.Config, params.nUsers)
	addrs := make([]*net.TCPAddr, params.nUsers)

	// set eviction params to ensure that evictions actually happen during test
	storageParams := storage.NewDefaultParameters()
	storageParams.Type = bstorage.DataStore

	for i := uint(0); i < params.nUsers; i++ {
		serverPort, metricsPort := startPort+i*10, startPort+i*10+1
		configs[i] = server.NewDefaultConfig().
			WithStorage(storageParams).
			WithGCPProjectID(params.gcpProjectID)
		configs[i].WithServerPort(uint(serverPort)).
			WithMetricsPort(uint(metricsPort)).
			WithLogLevel(params.logLevel)
		addrs[i] = &net.TCPAddr{IP: net.ParseIP("localhost"), Port: int(serverPort)}
	}
	return configs, addrs
}

func tearDown(st *state) {
	for _, c := range st.users {
		c.StopServer()
	}
	bstorage.StopDatastoreEmulator(st.datastoreProc)
	err := os.RemoveAll(st.dataDir)
	errors.MaybePanic(err)
}
