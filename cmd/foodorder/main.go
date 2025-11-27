package main

import (
	"backend-challenge/config"
	"backend-challenge/internal/db"
	"backend-challenge/internal/generated/openapi"
	"backend-challenge/internal/services"
	"database/sql"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func setupDB(db string) *sql.DB {
	log.Printf("DB Path: %s", db)
	d, err := sql.Open("sqlite3", db)
	if err != nil {
		log.Fatalf("sql.Open failed: %v", err)
	}
	return d
}

func main() {
	log.Printf("Server started")
	config := config.GetConfig()
	conn := setupDB(config.Db)
	defer conn.Close()
	product_dao := db.NewProductDao(conn)
	order_dao := db.NewOrderDao(conn)

	OrderAPIService := services.NewOrderAPIService(order_dao, product_dao, config.CouponBase, config.CouponMin)
	OrderAPIController := openapi.NewOrderAPIController(OrderAPIService)

	ProductAPIService := services.NewProductAPIService(product_dao)
	ProductAPIController := openapi.NewProductAPIController(ProductAPIService)

	router := openapi.NewRouter(OrderAPIController, ProductAPIController)

	log.Fatal(http.ListenAndServe(":8080", router))
}
