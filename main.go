package main

import (
	"context"
	"fmt"
	"forum/interface/routing"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func runServer(addr string) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	sugarLogger := logger.Sugar()

	err := godotenv.Load(".env")
	if err != nil {
		sugarLogger.Fatal("Could not load .env file", zap.String("error", err.Error()))
	}

	dbPrefix := os.Getenv("DB_PREFIX")

	connectionString := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s",
		os.Getenv(dbPrefix+"_DB_USER"), os.Getenv(dbPrefix+"_DB_PASSWORD"), os.Getenv(dbPrefix+"_DB_HOST"),
		os.Getenv(dbPrefix+"_DB_PORT"), os.Getenv(dbPrefix+"_DB_NAME"))

	fmt.Println(connectionString)

	conn, err := pgxpool.Connect(context.Background(), connectionString)
	if err != nil {
		sugarLogger.Fatal("Could not connect to database. Closing...", zap.String("error", err.Error()))
		return
	}

	defer conn.Close()
	fmt.Println("Successfully connected to database")
	r := routing.CreateRouter(conn, logger)

	allowedOrigins := make([]string, 0)

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
	})

	handler := c.Handler(r)
	fmt.Printf("Starting server at localhost%s\n", addr)

	switch os.Getenv("HTTPS_ON") {
	case "true":
		log.Fatal(http.ListenAndServeTLS(addr, "cert.pem", "key.pem", handler))
	case "false":
		log.Fatal(http.ListenAndServe(addr, handler))
	}
}

func main() {
	runServer(":5000")
}
