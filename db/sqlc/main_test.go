package db

import (
	"context"
	"github.com/SaishNaik/simplebank/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"testing"
)

var testStore Store

func TestMain(m *testing.M) {
	var err error
	config, err := utils.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load configurations", err)
	}

	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db", err)
	}
	testStore = NewStore(connPool)
	os.Exit(m.Run())

}
