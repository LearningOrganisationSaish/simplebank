package main

import (
	"context"
	"errors"
	"github.com/SaishNaik/simplebank/api"
	db "github.com/SaishNaik/simplebank/db/sqlc"
	_ "github.com/SaishNaik/simplebank/doc/statik"
	"github.com/SaishNaik/simplebank/gapi"
	"github.com/SaishNaik/simplebank/mail"
	"github.com/SaishNaik/simplebank/pb"
	"github.com/SaishNaik/simplebank/utils"
	"github.com/SaishNaik/simplebank/worker"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakyll/statik/fs"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var interruptSignals = []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGINT}

func main() {

	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load configurations ")
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()

	connPool, err := pgxpool.New(ctx, config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	runDbMigration(config.MigrationURL, config.DBSource)

	store := db.NewStore(connPool)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	waitGroup, ctx := errgroup.WithContext(ctx)

	runTaskProcessor(ctx, waitGroup, config, redisOpt, store)
	runGatewayServer(ctx, waitGroup, config, store, taskDistributor)
	runGrpcServer(ctx, waitGroup, config, store, taskDistributor)

	err = waitGroup.Wait()
	if err != nil {
		log.Fatal().Err(err).Msg("wait group exited with error")
	}
	//runGinServer(config, store)

}

func runTaskProcessor(ctx context.Context, waitGroup *errgroup.Group, config utils.Config, redisOpt asynq.RedisClientOpt, store db.Store) {
	mailer := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer)
	log.Info().Msg("start task processor")
	err := taskProcessor.Start()
	if err != nil {
		log.Error().Err(err).Msg("cannot start task processor")
	}

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("Graceful shutdown task processor")
		taskProcessor.Shutdown()
		log.Info().Msg("Shutdown of task processor completed")
		return nil
	})

}

func runGrpcServer(ctx context.Context, waitGroup *errgroup.Group, config utils.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)

	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer) // kind of documentation for the server

	listener, err := net.Listen("tcp", config.GrpcServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}

	waitGroup.Go(func() error {
		log.Info().Msgf("grpc server listening on %s", listener.Addr().String())
		err = grpcServer.Serve(listener)
		if errors.Is(err, grpc.ErrServerStopped) {
			return nil
		}
		if err != nil {
			log.Error().Err(err).Msg("grpc server failed to serve")
		}
		return err
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("graceful shutdown grpc server")
		grpcServer.GracefulStop()
		log.Info().Msg("grpc server is stopped")
		return nil
	})

}

func runGatewayServer(ctx context.Context, waitGroup *errgroup.Group, config utils.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
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

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start grpc handler server")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFs, err := fs.New()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create statik fs")
	}
	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFs))
	mux.Handle("/swagger/", swaggerHandler)

	c := cors.New(cors.Options{
		//AllowedOrigins: config.AllowedOrigins,
	})
	corsHandler := c.Handler(gapi.HttpLogger(mux))
	httpServer := &http.Server{
		Handler: corsHandler,
		Addr:    config.HttpServerAddress,
	}

	waitGroup.Go(func() error {
		log.Info().Msgf("http gateway server listening on %s", httpServer.Addr)
		err = httpServer.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		if err != nil {
			log.Error().Err(err).Msg("http gateway server failed to serve")
		}
		return err
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("graceful shutdown http gateway server")
		err := httpServer.Shutdown(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("failed to shutdown http gateway server")
			return err
		}
		log.Info().Msg("http gateway server is stopped")
		return nil
	})

}

func runGinServer(config utils.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}

	err = server.Start(config.HttpServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server:")
	}
}

func runDbMigration(migrationUrl, dbSource string) {
	migration, err := migrate.New(migrationUrl, dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create migration:")
	}
	err = migration.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal().Err(err).Msg("failed to run migrate up:")
	}
	log.Info().Msg("db migrated successfully")
}
