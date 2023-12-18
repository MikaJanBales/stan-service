package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MikaJanBales/stan-service/go-service/internal/config"
	cache "github.com/MikaJanBales/stan-service/go-service/internal/redis"
	"github.com/MikaJanBales/stan-service/go-service/internal/service"
	"github.com/MikaJanBales/stan-service/go-service/internal/storage"
	transport "github.com/MikaJanBales/stan-service/go-service/internal/transport/http"
	redisClient "github.com/MikaJanBales/stan-service/go-service/pkg/redis"

	"github.com/go-redis/redis"
	"github.com/jackc/pgx/v4"
	"github.com/nats-io/stan.go"
	"github.com/rs/zerolog/log"
)

func main() {
	// Getting a config from environment variables
	cfg := config.NewConfig()
	// Initialization logger
	log.Logger = config.Logger()

	// Connecting db
	connStr := connectionString(
		cfg.DbConfig.Host,
		cfg.DbConfig.Database,
		cfg.DbConfig.User,
		cfg.DbConfig.Password,
	)
	pgConn := newDbConnection(connStr)

	// Connecting NATS
	sc, err := stan.Connect(cfg.NATS.ClusterName, cfg.NATS.ClientID)
	if err != nil {
		log.Error().Err(err)
	}

	rClient := redis.NewClient(&redis.Options{
		Addr:       cfg.Redis.Address,
		MaxRetries: cfg.Redis.MaxRetries,
	})

	cClient := redisClient.New(rClient)

	cacheClient := cache.NewCacheClient(cClient, log.Logger)
	strg := storage.New(log.Logger, pgConn, cacheClient)
	svc := service.New(log.Logger, sc, strg)

	// Initialization transport
	ts := transport.New(svc)
	router := ts.InitRouter()

	go func() {
		if err = svc.ReceiveMessage(cfg.NATS.SubjectName); err != nil {
			log.Fatal().Err(err)
		}
	}()

	cors := transport.CorsSettings()

	server := &http.Server{
		Addr:         cfg.ServerConfig.Host,
		Handler:      cors.Handler(router),
		ReadTimeout:  http.DefaultClient.Timeout,
		WriteTimeout: http.DefaultClient.Timeout,
	}

	log.Info().Msgf("starting server at addr: %s", server.Addr)
	go func() {
		if err = server.ListenAndServe(); err != nil {
			log.Fatal().Err(err).Send()
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	log.Info().Msg("http server shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Send()
	}
}

func connectionString(host, db, user, password string) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		user,
		password,
		host,
		db,
	)
}

func newDbConnection(connStr string) *pgx.Conn {
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get db connection")
	}

	return conn
}
