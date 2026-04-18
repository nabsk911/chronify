package handlers

import (
	"log"
	"net/http"

	"github.com/nabsk911/chronify/internal/db"
	"github.com/nabsk911/chronify/internal/utils"
)

type BookmarkHandler struct {
	bookmarkStore *db.Queries
	logger        *log.Logger
}

func NewBookmarkHandler(bookmarkStore *db.Queries, logger *log.Logger) *BookmarkHandler {
	return &BookmarkHandler{
		bookmarkStore: bookmarkStore,
		logger:        logger,
	}
}

func (bh *BookmarkHandler) HandleGetBookmarks(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.ReadUserIDFromContext(r, "userID")
	if err != nil {
		bh.logger.Printf("Invalid user ID: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid user ID"})
		return
	}

	timelines, err := bh.bookmarkStore.GetBookmarkedTimelinesByUserId(r.Context(), userID)
	if err != nil {
		bh.logger.Printf("Failed to retrieve timeline %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to retrieve timeline"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": timelines})

}

func (bh *BookmarkHandler) HandleAddBookmark(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.ReadUserIDFromContext(r, "userID")
	if err != nil {
		bh.logger.Printf("Invalid user ID: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid user ID"})
		return
	}

	timelineID, err := utils.ReadIDParam(r, "timelineId")
	if err != nil {
		bh.logger.Printf("Invalid timeline ID: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid timeline ID"})
		return
	}

	err = bh.bookmarkStore.AddBookmark(r.Context(), db.AddBookmarkParams{
		UserID:     userID,
		TimelineID: timelineID,
	})
	if err != nil {
		bh.logger.Printf("Failed to add bookmark: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to add bookmark"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"message": "Bookmark added successfully"})

}

func (bh *BookmarkHandler) HandleRemoveBookmark(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.ReadUserIDFromContext(r, "userID")
	if err != nil {
		bh.logger.Printf("Invalid user ID: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid user ID"})
		return
	}

	timelineID, err := utils.ReadIDParam(r, "timelineId")
	if err != nil {
		bh.logger.Printf("Invalid timeline ID: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid timeline ID"})
		return
	}

	err = bh.bookmarkStore.RemoveBookmark(r.Context(), db.RemoveBookmarkParams{
		UserID:     userID,
		TimelineID: timelineID,
	})
	if err != nil {
		bh.logger.Printf("Failed to add bookmark: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to remove bookmark"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"message": "Bookmark removed successfully"})
}
