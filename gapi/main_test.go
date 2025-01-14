package gapi

import (
	db "github.com/SaishNaik/simplebank/db/sqlc"
	"github.com/SaishNaik/simplebank/utils"
	"github.com/SaishNaik/simplebank/worker"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func NewTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor) *Server {
	config := utils.Config{
		TokenSymmetricKey:   utils.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store, taskDistributor)
	require.NoError(t, err)
	return server
}
