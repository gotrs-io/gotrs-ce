package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gotrs-io/gotrs/internal/auth"
	"github.com/gotrs-io/gotrs/internal/config"
	"github.com/gotrs-io/gotrs/internal/database"
	authpb "github.com/gotrs-io/gotrs/proto/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize query optimizer
	optimizer := database.NewQueryOptimizer(db, database.DefaultOptimizerConfig())

	// Initialize auth service
	authService := auth.NewService(db, optimizer, cfg.Auth)

	// Create gRPC server
	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(1024*1024*4), // 4MB
		grpc.MaxSendMsgSize(1024*1024*4), // 4MB
	)

	// Register services
	authpb.RegisterAuthServiceServer(server, authService)
	
	// Register health service
	healthServer := health.NewServer()
	healthServer.SetServingStatus("auth", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	// Enable reflection for development
	if cfg.Environment == "development" {
		reflection.Register(server)
	}

	// Start server
	port := cfg.Services.Auth.Port
	if port == "" {
		port = "50001"
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	log.Printf("Auth service starting on port %s", port)

	// Start server in a goroutine
	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down auth service...")

	// Graceful shutdown
	server.GracefulStop()
	log.Println("Auth service stopped")
}