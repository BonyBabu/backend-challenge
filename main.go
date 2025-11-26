package main

import (
	"backend-challenge/internal/db"
	"backend-challenge/internal/generated/openapi"
	"backend-challenge/internal/services"
	"database/sql"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

func setupDB() *sql.DB {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Failed to Setup DB")
	}
	currentFileDir := filepath.Dir(filename)
	currentFileDir = filepath.ToSlash(currentFileDir)
	dbPath := filepath.Join(currentFileDir, "db.sqlite3")
	log.Printf("DB Path: %s", dbPath)
	d, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("sql.Open failed: %v", err)
	}
	return d
}

func main() {
	log.Printf("Server started")
	conn := setupDB()
	defer conn.Close()
	product_dao := db.NewProductDao(conn)
	order_dao := db.NewOrderDao(conn)

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatalf("Failed to get current file path")
	}
	couponbase1 := path.Join(path.Dir(filename), "/couponbase/couponbase1")
	couponbase2 := path.Join(path.Dir(filename), "/couponbase/couponbase2")
	couponbase3 := path.Join(path.Dir(filename), "/couponbase/couponbase3")

	OrderAPIService := services.NewOrderAPIService(order_dao, product_dao, []string{couponbase1, couponbase2, couponbase3})
	OrderAPIController := openapi.NewOrderAPIController(OrderAPIService)

	ProductAPIService := services.NewProductAPIService(product_dao)
	ProductAPIController := openapi.NewProductAPIController(ProductAPIService)

	router := openapi.NewRouter(OrderAPIController, ProductAPIController)

	log.Fatal(http.ListenAndServe(":8080", router))
}
