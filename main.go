package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/SaishNaik/simplebank/api"
	db "github.com/SaishNaik/simplebank/db/sqlc"
	_ "github.com/SaishNaik/simplebank/doc/statik"
	"github.com/SaishNaik/simplebank/gapi"
	"github.com/SaishNaik/simplebank/pb"
	"github.com/SaishNaik/simplebank/utils"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"

	"log"
	"net"
	"net/http"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load configurations ", err)
	}

	fmt.Println(config.DBSource)
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db", err)
	}

	runDbMigration(config.MigrationURL, config.DBSource)

	store := db.NewStore(conn)
	go runGatewayServer(config, store)
	runGrpcServer(config, store)
	//runGinServer(config, store)

}

func runGrpcServer(config utils.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer) // kind of documentation for the server

	listener, err := net.Listen("tcp", config.GrpcServerAddress)
	if err != nil {
		log.Fatal("cannot create listener", err)
	}

	log.Printf("grpc server listening on %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start grpc server", err)
	}

}

func runGatewayServer(config utils.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server", err)
	}
	// to make http response to snake case
	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})
	grpcMux := runtime.NewServeMux(jsonOption)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal("cannot start grpc handler server", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFs, err := fs.New()
	if err != nil {
		log.Fatal("cannot create statik fs", err)
	}
	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFs))
	mux.Handle("/swagger/", swaggerHandler)

	listener, err := net.Listen("tcp", config.HttpServerAddress)
	if err != nil {
		log.Fatal("cannot create listener", err)
	}

	log.Printf("http gateway server listening on %s", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatal("cannot start http gateway server", err)
	}

}

func runGinServer(config utils.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server", err)
	}

	err = server.Start(config.HttpServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}

func runDbMigration(migrationUrl, dbSource string) {
	migration, err := migrate.New(migrationUrl, dbSource)
	if err != nil {
		log.Fatal("cannot create migration:", err)
	}
	err = migration.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal("failed to run migrate up:", err)
	}
	log.Println("db migrated successfully")
}
