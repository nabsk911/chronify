package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nabsk911/chronify/internal/db"
	"github.com/nabsk911/chronify/internal/utils"
)

type timelineRequest struct {
	UserID      pgtype.UUID `json:"user_id"`
	Title       string      `json:"title"`
	Description string      `json:"description,omitempty"`
}
type TimelineHandler struct {
	timelineStore *db.Queries
	logger        *log.Logger
}

func NewTimelineHandler(timelineStore *db.Queries, logger *log.Logger) *TimelineHandler {
	return &TimelineHandler{
		timelineStore: timelineStore,
		logger:        logger,
	}
}

// Create
func (th *TimelineHandler) HandleCreateTimeline(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("userID").(string)

	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		th.logger.Printf("Invalid user ID format: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid user ID"})
		return
	}

	var req timelineRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		th.logger.Printf("Failed to decode timeline request: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid request payload!"})
		return
	}
	if req.Title == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Title  is required"})
		return
	}

	// Convert string to pgtype.Text
	timeline, err := th.timelineStore.CreateTimeline(r.Context(), db.CreateTimelineParams{
		UserID:      userID,
		Title:       req.Title,
		Description: pgtype.Text{String: req.Description, Valid: true},
	})
	if err != nil {
		th.logger.Printf("Failed to create user in store: %v", err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				switch pgErr.ConstraintName {
				case "timelines_title_key":
					utils.WriteJSON(w, http.StatusConflict, utils.Envelope{"message": "Timeline with this title already exists"})
					return
				}
			}
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to create timeline"})
		return

	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"data": timeline, "message": "Timeline created successfully"})
}

func (th *TimelineHandler) HandleGetTimelines(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("userID").(string)

	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		th.logger.Printf("Invalid user ID format: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid user ID"})
		return
	}

	timelines, err := th.timelineStore.GetTimelinesByUserId(r.Context(), userID)
	if err != nil {
		th.logger.Printf("Failed to retrieve timeline for user %s: %v", userIDStr, err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to retrieve timeline"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": timelines})
}

func (th *TimelineHandler) HandleGetTimelineById(w http.ResponseWriter, r *http.Request) {
	timelineID, err := utils.ReadIDParam(r, "timelineId")
	if err != nil {
		th.logger.Printf("Invalid timeline ID: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid timeline ID"})
		return
	}
	timeline, err := th.timelineStore.GetTimeLineById(r.Context(), timelineID)
	if err != nil {
		th.logger.Printf("Failed to retrieve timeline: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to retrieve timeline"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": timeline})
}

func (th *TimelineHandler) HandleSearchTimeline(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	userIDStr := r.Context().Value("userID").(string)

	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		th.logger.Printf("Invalid user ID format: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid user ID"})
		return
	}

	timelines, err := th.timelineStore.GetTimelinesByUserIdAndTitle(r.Context(), db.GetTimelinesByUserIdAndTitleParams{
		UserID:  userID,
		Column2: pgtype.Text{String: title, Valid: true},
	})
	if err != nil {
		th.logger.Printf("Failed to retrieve timeline for user %s: %v", userIDStr, err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to retrieve timeline"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": timelines})
}

func (th *TimelineHandler) HandleUpdateTimeline(w http.ResponseWriter, r *http.Request) {
	timelineID, err := utils.ReadIDParam(r, "timelineId")
	if err != nil {
		th.logger.Printf("Invalid timeline ID: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid timeline ID"})
		return
	}
	var req timelineRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		th.logger.Printf("Failed to decode timeline request: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid request payload!"})
		return
	}
	if req.Title == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Title is required"})
		return
	}

	timeline, err := th.timelineStore.UpdateTimeline(r.Context(), db.UpdateTimelineParams{
		ID:          timelineID,
		Title:       req.Title,
		Description: pgtype.Text{String: req.Description, Valid: true},
	})

	if err != nil {
		th.logger.Printf("Failed to update timeline: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to update timeline"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": timeline})
}

func (th *TimelineHandler) HandleDeleteTimeline(w http.ResponseWriter, r *http.Request) {
	timelineID, err := utils.ReadIDParam(r, "timelineId")
	if err != nil {
		th.logger.Printf("Invalid timeline ID: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid timeline ID"})
		return
	}
	err = th.timelineStore.DeleteTimeline(r.Context(), timelineID)
	if err != nil {
		th.logger.Printf("Failed to delete timeline: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to delete timeline"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"message": "Timeline deleted successfully"})
}
