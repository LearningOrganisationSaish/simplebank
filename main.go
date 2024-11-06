package main

import (
	"database/sql"
	"github.com/SaishNaik/simplebank/api"
	db "github.com/SaishNaik/simplebank/db/sqlc"
	"github.com/SaishNaik/simplebank/utils"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load configurations ", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db", err)
	}
	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
