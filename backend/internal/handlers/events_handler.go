package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nabsk911/chronify/internal/db"
	"github.com/nabsk911/chronify/internal/utils"
)

type UpsertEventRequest struct {
	ID               *pgtype.UUID `json:"id,omitempty"`
	Title            string       `json:"title"`
	CardTitle        string       `json:"card_title"`
	CardSubtitle     pgtype.Text  `json:"card_subtitle"`
	CardDetailedText pgtype.Text  `json:"card_detailed_text"`
}

type EventHandler struct {
	eventStore *db.Queries
	logger     *log.Logger
}

func NewEventHandler(eventStore *db.Queries, logger *log.Logger) *EventHandler {
	return &EventHandler{
		eventStore: eventStore,
		logger:     logger,
	}
}

func (eh *EventHandler) HandleGetEventsByTimelineId(w http.ResponseWriter, r *http.Request) {
	timelineID, err := utils.ReadIDParam(r, "timelineId")
	if err != nil {
		eh.logger.Printf("Invalid timeline ID: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid timeline ID"})
		return
	}

	events, err := eh.eventStore.GetEventsByTimelineId(r.Context(), timelineID)
	if err != nil {
		eh.logger.Printf("Failed to retrieve events: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to retrieve events"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"events": events})
}

func (eh *EventHandler) HandleUpsertEvents(w http.ResponseWriter, r *http.Request) {
	timelineID, err := utils.ReadIDParam(r, "timelineId")
	if err != nil {
		eh.logger.Printf("Invalid timeline ID: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid timeline ID"})
		return
	}

	var req []UpsertEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		eh.logger.Printf("Failed to decode request: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid request payload"})
		return
	}

	if len(req) == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "No events provided"})
		return
	}

	var createParams []db.BulkCreateEventsParams
	var updateParams []db.BulkUpdateEventsParams

	for _, e := range req {
		if e.ID == nil || !e.ID.Valid {
			// New event → create
			createParams = append(createParams, db.BulkCreateEventsParams{
				TimelineID:       timelineID,
				Title:            e.Title,
				CardTitle:        e.CardTitle,
				CardSubtitle:     e.CardSubtitle,
				CardDetailedText: e.CardDetailedText,
			})
		} else {
			// Existing event → update
			updateParams = append(updateParams, db.BulkUpdateEventsParams{
				ID:               *e.ID,
				Title:            e.Title,
				CardTitle:        e.CardTitle,
				CardSubtitle:     e.CardSubtitle,
				CardDetailedText: e.CardDetailedText,
			})
		}
	}

	ctx := r.Context()

	// Bulk create new events
	if len(createParams) > 0 {
		_, err := eh.eventStore.BulkCreateEvents(ctx, createParams)
		if err != nil {
			eh.logger.Printf("Failed to create events: %v", err)
			utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to create events"})
			return
		}
	}

	// Bulk update existing events
	if len(updateParams) > 0 {
		results := eh.eventStore.BulkUpdateEvents(ctx, updateParams)
		defer results.Close()
		results.Exec(func(i int, err error) {
			if err != nil {
				eh.logger.Printf("Failed to update event: %v", err)
				utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to update one or more events"})
				return
			}
		})
	}

	// Return full list for the timeline
	events, err := eh.eventStore.GetEventsByTimelineId(ctx, timelineID)
	if err != nil {
		eh.logger.Printf("Failed to retrieve events: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to retrieve events"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"events": events})
}

func (eh *EventHandler) HandleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	timelineID, err := utils.ReadIDParam(r, "timelineId")
	if err != nil {
		eh.logger.Printf("Invalid timeline ID: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid timeline ID"})
		return
	}

	eventID, err := utils.ReadIDParam(r, "eventId")
	if err != nil {
		eh.logger.Printf("Invalid event ID: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid event ID"})
		return
	}

	err = eh.eventStore.DeleteEvent(r.Context(), eventID)
	if err != nil {
		eh.logger.Printf("Failed to delete event: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to delete event"})
		return
	}

	events, err := eh.eventStore.GetEventsByTimelineId(r.Context(), timelineID)
	if err != nil {
		eh.logger.Printf("Failed to retrieve events: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to retrieve events"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"message": "Event deleted successfully", "events": events})
}
