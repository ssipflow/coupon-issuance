package main

import (
	"github.com/joho/godotenv"
	"github.com/ssipflow/coupon-issuance/gen/proto/coupon/v1/couponv1connect"
	"github.com/ssipflow/coupon-issuance/internal/handler"
	"github.com/ssipflow/coupon-issuance/internal/repo"
	"github.com/ssipflow/coupon-issuance/internal/service"
	"github.com/ssipflow/coupon-issuance/internal/task"
	"log"
	"net/http"
	"os"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found")
	}

	redisClient := repo.NewRedisClient()
	db := repo.NewRepository()
	couponService := service.NewCouponService(redisClient, db)

	asynqWorker := task.NewAsynqWorker(db, redisClient)
	go func() {
		asynqWorker.Start()
	}()

	couponHandler := handler.NewCouponHandler(couponService)
	path, svcHandler := couponv1connect.NewCouponServiceHandler(couponHandler)

	mux := http.NewServeMux()
	mux.Handle(path, svcHandler)

	addr := ":" + os.Getenv("PORT")
	log.Println("Listening on " + addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
