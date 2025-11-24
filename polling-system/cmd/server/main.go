package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "polling-system/docs"
	"polling-system/internal/config"
	"polling-system/internal/domain/poll"
	"polling-system/internal/domain/user"
	"polling-system/internal/domain/vote"
	api "polling-system/internal/http"
	"polling-system/internal/platform/database"
	jwtpkg "polling-system/internal/platform/jwt"
	"polling-system/internal/repository/postgres"
	"polling-system/internal/worker"
)

// @title           Polling System API
// @version         1.0
// @description     Simple polling platform with JWT auth
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
func main() {
	cfg := config.Load()

	db, err := database.NewPostgres(cfg.DB_DSN)
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	defer db.Close()

	userRepo := postgres.NewUserRepo(db)
	pollRepo := postgres.NewPollRepo(db)
	voteRepo := postgres.NewVoteRepo(db)

	userSvc := user.NewService(userRepo)
	pollSvc := poll.NewService(pollRepo)
	voteSvc := vote.NewService(voteRepo)

	jwtMgr := jwtpkg.NewManager(cfg.JWTSecret)

	voteCh := make(chan worker.VoteEvent, 100)
	statsWorker := worker.NewStatsWorker(voteCh)

	router := api.NewRouter(userSvc, pollSvc, voteSvc, jwtMgr, voteCh)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go statsWorker.Run(ctx)

	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	log.Println("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}

	log.Println("server stopped")
}
