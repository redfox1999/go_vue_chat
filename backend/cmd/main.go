package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/config"
	"backend/handler"
	"backend/repository"
	"backend/router"
	"backend/service"
	"backend/websocket"
)

func main() {
	config.InitLogger()
	config.LoadEnv()
	config.InitSecurity()

	err := config.InitDB()
	if err != nil {
		config.Logger.Fatal().Err(err).Msg("Failed to initialize database")
	}

	userRepo := repository.NewUserRepository(config.DB, config.Logger)
	userService := service.NewUserService(userRepo, config.Logger)
	userHandler := handler.NewUserHandler(userService, config.Logger)

	wsManager := websocket.NewManager(config.Logger)
	go wsManager.Run()
	wsHandler := handler.NewWebSocketHandler(wsManager, config.Logger)

	r := router.NewRouter(userHandler, wsHandler, config.Logger)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		config.Logger.Info().Msgf("Server is running on http://localhost:%s", port)
		config.Logger.Info().Msgf("WebSocket endpoint available at ws://localhost:%s/ws", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			config.Logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	config.Logger.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		config.Logger.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	if err := config.DB.Close(); err != nil {
		config.Logger.Error().Err(err).Msg("Failed to close database connection")
	}

	config.Logger.Info().Msg("Server shutdown completed")
}
