package main

import (
	"context"
	"database/sql"
	"flash-sale-reservation/internal/reservation"
	"log"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"

	apphttp "flash-sale-reservation/internal/http"
	"flash-sale-reservation/internal/outbox"
	"flash-sale-reservation/internal/product"
)

func main() {
	ctx := context.Background()

	// ---------- PostgreSQL ----------
	dsn := "postgres://postgres:postgres@127.0.0.1:5433/reservation_db?sslmode=disable"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("PostgreSQL connected")

	// ---------- Redis ----------
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal(err)
	}

	log.Println("Redis connected")

	// ---------- Products ----------
	productRepo := product.NewRepository(db)
	productService := product.NewService(productRepo)

	// ---------- Reservations ----------
	reservationRepo := reservation.NewRepository(db)
	outboxRepo := outbox.NewRepository(db)

	reservationService := reservation.NewService(
		reservationRepo,
		productRepo,
		outboxRepo,
		rdb,
	)

	// ---------- HTTP ----------
	router := apphttp.NewRouter(
		productService,
		reservationService,
	)

	log.Println("HTTP server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
