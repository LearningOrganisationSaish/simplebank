package gapi

import (
	"fmt"
	db "github.com/SaishNaik/simplebank/db/sqlc"
	"github.com/SaishNaik/simplebank/pb"
	"github.com/SaishNaik/simplebank/token"
	"github.com/SaishNaik/simplebank/utils"
	"github.com/SaishNaik/simplebank/worker"
)

type Server struct {
	config          utils.Config
	store           db.Store
	taskDistributor worker.TaskDistributor
	tokenMaker      token.Maker
	pb.UnimplementedSimpleBankServer
}

func NewServer(config utils.Config, store db.Store, taskDistributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:          config,
		store:           store,
		tokenMaker:      tokenMaker,
		taskDistributor: taskDistributor,
	}
	return server, nil
}
