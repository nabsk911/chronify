package app

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nabsk911/chronify/internal/db"
	"github.com/nabsk911/chronify/internal/handlers"
)

type Application struct {
	DB              *db.Queries
	DBConn          *pgxpool.Pool
	Logger          *log.Logger
	UserHandler     *handlers.UserHandler
	TimelineHandler *handlers.TimelineHandler
	EventHandler    *handlers.EventHandler
}

func NewApplication() (*Application, error) {

	dbURL := os.Getenv("DB_URL")

	conn, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, err
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	queries := db.New(conn)

	return &Application{
		DB:              queries,
		DBConn:          conn,
		Logger:          logger,
		UserHandler:     handlers.NewUserHandler(queries, logger),
		TimelineHandler: handlers.NewTimelineHandler(queries, logger),
		EventHandler:    handlers.NewEventHandler(queries, logger),
	}, nil
}
