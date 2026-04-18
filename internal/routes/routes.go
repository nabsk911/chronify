package routes

import (
	"net/http"

	"github.com/nabsk911/chronify/internal/app"
	"github.com/nabsk911/chronify/internal/middleware"
)

func SetupRoutes(app *app.Application) *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("POST /register", app.UserHandler.HandleRegister)
	router.HandleFunc("POST /login", app.UserHandler.HandleLogin)
	router.HandleFunc("GET /timelines", middleware.Authentication(app.TimelineHandler.HandleGetTimelines))
	router.HandleFunc("GET /timelines/{timelineId}", middleware.Authentication(app.TimelineHandler.HandleGetTimelineById))
	router.HandleFunc("GET /timelines/search", middleware.Authentication(app.TimelineHandler.HandleSearchTimeline))
	router.HandleFunc("GET /timelines/public", middleware.Authentication(app.TimelineHandler.HandleGetPublicTimelines))
	router.HandleFunc("GET /timelines/public/search", middleware.Authentication(app.TimelineHandler.HandleSearchPublicTimeline))
	router.HandleFunc("POST /timelines", middleware.Authentication(app.TimelineHandler.HandleCreateTimeline))
	router.HandleFunc("PUT /timelines/{timelineId}", middleware.Authentication(app.TimelineHandler.HandleUpdateTimeline))
	router.HandleFunc("DELETE /timelines/{timelineId}", middleware.Authentication(app.TimelineHandler.HandleDeleteTimeline))
	router.HandleFunc("GET /timelines/{timelineId}/events", middleware.Authentication(app.EventHandler.HandleGetEventsByTimelineId))
	router.HandleFunc("POST /timelines/{timelineId}/events", middleware.Authentication(app.EventHandler.HandleUpsertEvents))
	router.HandleFunc("POST /timelines/{timelineId}/aievents", middleware.LimitMiddleware(middleware.Authentication(app.EventHandler.HandleCreateAIEvents)))
	router.HandleFunc("PUT /timelines/{timelineId}/aievents", middleware.LimitMiddleware(middleware.Authentication(app.EventHandler.HandleUpdateAIEvents)))
	router.HandleFunc("DELETE /timelines/{timelineId}/events/{eventId}", middleware.Authentication(app.EventHandler.HandleDeleteEvent))
	router.HandleFunc("GET /bookmarks", middleware.Authentication(app.BookmarkHandler.HandleGetBookmarks))
	router.HandleFunc("DELETE /bookmarks/{timelineId}", middleware.Authentication(app.BookmarkHandler.HandleRemoveBookmark))
	router.HandleFunc("POST /bookmarks/{timelineId}", middleware.Authentication(app.BookmarkHandler.HandleAddBookmark))
	return router
}
