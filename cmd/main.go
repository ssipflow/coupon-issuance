package main

import (
	"context"
	"errors"
	"github.com/joho/godotenv"
	"github.com/ssipflow/coupon-issuance/gen/proto/coupon/v1/couponv1connect"
	"github.com/ssipflow/coupon-issuance/internal/handler"
	"github.com/ssipflow/coupon-issuance/internal/infra"
	"github.com/ssipflow/coupon-issuance/internal/repo"
	"github.com/ssipflow/coupon-issuance/internal/service"
	"github.com/ssipflow/coupon-issuance/internal/task"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	redisClient := infra.NewRedisClient()
	db := infra.NewMySql()

	// AutoMigrate
	if os.Getenv("IS_PRIMARY") == "true" {
		if err := db.AutoMigrate(); err != nil {
			log.Fatalf("failed to auto migrate: %v", err)
		}
	}

	couponRepository := repo.NewCouponRepository(db.GetDB())
	couponService := service.NewCouponService(redisClient, couponRepository)
	asynqWorker := task.NewAsynqWorker(couponRepository, redisClient)

	go asynqWorker.Start()

	couponHandler := handler.NewCouponHandler(couponService)
	path, svcHandler := couponv1connect.NewCouponServiceHandler(couponHandler)

	mux := http.NewServeMux()
	mux.Handle(path, svcHandler)

	server := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: mux,
	}

	go func() {
		log.Printf("Listening on port %v\n", os.Getenv("PORT"))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	if os.Getenv("IS_PRIMARY") == "true" {
		// Drop tables - this is optional and should be used with caution
		if err := db.DropTables(); err != nil {
			log.Fatalf("failed to drop tables: %v\n", err)
		}
		// Flush Redis - this is optional and should be used with caution
		if err := redisClient.FlushAll(shutdownCtx); err != nil {
			log.Fatalf("failed to flush Redis: %v\n", err)
		}
	}

	log.Println("Gracefully shutting down server...")
}
