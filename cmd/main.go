package main

import (
	"github.com/joho/godotenv"
	"github.com/ssipflow/coupon-issuance/gen/proto/coupon/v1/couponv1connect"
	"github.com/ssipflow/coupon-issuance/internal/handler"
	"github.com/ssipflow/coupon-issuance/internal/repo"
	"github.com/ssipflow/coupon-issuance/internal/task"
	"log"
	"net/http"
	"os"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found")
	}

	redis := repo.NewRedisClient()
	db := repo.NewMySQLDB()

	go func() {
		task.StartWorker(redis, db)
	}()

	mux := http.NewServeMux()
	couponHandler := handler.NewCouponHandler(redis, db)
	path, svcHandler := couponv1connect.NewCouponServiceHandler(couponHandler)
	mux.Handle(path, svcHandler)

	addr := ":" + os.Getenv("PORT")
	log.Println("Listening on " + addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
