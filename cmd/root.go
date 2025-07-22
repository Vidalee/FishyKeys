package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/Vidalee/FishyKeys/internal/db"
	"github.com/Vidalee/FishyKeys/internal/migration"
	"github.com/Vidalee/FishyKeys/internal/server"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	pgUser     string
	pgPass     string
	pgAddress  string
	pgDatabase string
	serverAddr string
	serverPort int
)

var rootCmd = &cobra.Command{
	Use:   "fishykeys",
	Short: "FishyKeys server",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		user := getConfigValue("pg.user", pgUser)
		pass := getConfigValue("pg.pass", pgPass)
		addr := getConfigValue("pg.address", pgAddress)
		dbName := getConfigValue("pg.database", pgDatabase)

		dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, pass, addr, dbName)

		if err := db.Init(ctx, dsn); err != nil {
			log.Fatalf("failed to init db: %v", err)
		}
		defer db.Pool().Close()

		if err := migration.RunMigrations(db.Pool()); err != nil {
			log.Fatalf("failed to run migrations: %v", err)
		}

		serverAddress := getConfigValue("server.address", serverAddr)
		serverPort := getConfigPort("server.port", serverPort)

		goaServer, grpcServer := server.NewServers(db.Pool())

		httpServer := &http.Server{
			Addr:    fmt.Sprintf("%s:%s", serverAddress, serverPort),
			Handler: goaServer,
		}

		// Start HTTP server
		go func() {
			log.Printf("Starting HTTP server on %s:%s", serverAddress, serverPort)
			if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("failed to start HTTP server: %v", err)
			}
		}()

		// Start gRPC server
		go func() {
			lis, err := net.Listen("tcp", ":8090")
			if err != nil {
				log.Fatalf("failed to listen: %v", err)
			}
			log.Printf("gRPC server listening on :8090")
			if err := grpcServer.Serve(lis); err != nil {
				log.Fatalf("failed to serve: %v", err)
			}
		}()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutdown signal received")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		} else {
			log.Println("HTTP server shutdown gracefully")
		}

		log.Println("Shutting down gRPC server...")
		grpcServer.GracefulStop()
	},
}

func getConfigValue(key string, flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	return viper.GetString(key)
}

func getConfigPort(key string, flagValue int) string {
	if flagValue != 0 {
		return fmt.Sprintf("%d", flagValue)
	}
	return viper.GetString(key)
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("FISHYKEYS")

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			log.Printf("Error reading config file: %v", err)
		}
	}

	rootCmd.Flags().StringVar(&pgUser, "pg-user", "", "PostgreSQL username")
	rootCmd.Flags().StringVar(&pgPass, "pg-pass", "", "PostgreSQL password")
	rootCmd.Flags().StringVar(&pgAddress, "pg-address", "", "PostgreSQL address")
	rootCmd.Flags().StringVar(&pgDatabase, "pg-database", "", "PostgreSQL database name")

	rootCmd.Flags().StringVar(&serverAddr, "server-address", "", "Server address to listen on")
	rootCmd.Flags().IntVar(&serverPort, "server-port", 0, "Server port to listen on")
}
