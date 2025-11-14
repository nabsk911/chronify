package main

import (
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/nabsk911/chronify/internal/app"
	"github.com/nabsk911/chronify/internal/middleware"
	"github.com/nabsk911/chronify/internal/routes"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	app, err := app.NewApplication()
	if err != nil {
		panic(err)
	}

	defer app.DBConn.Close()
	app.Logger.Println("Starting server on port 8080")

	r := routes.SetupRoutes(app)
	handler := middleware.CORS(r)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server.ListenAndServe()
}
