// TODO: Apply rate limiting
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nabsk911/chronify/internal/db"
	"github.com/nabsk911/chronify/internal/utils"
	"google.golang.org/genai"
)

type AIEventRequest struct {
	Prompt string `json:"prompt"`
}

type TimelineEventRequest struct {
	Title            string      `json:"title"`
	CardTitle        string      `json:"card_title"`
	CardSubtitle     pgtype.Text `json:"card_subtitle"`
	CardDetailedText pgtype.Text `json:"card_detailed_text"`
}

func (eh *EventHandler) HandleCreateAIEvents(w http.ResponseWriter, r *http.Request) {

	apiKey := os.Getenv("GEMINI_API_KEY")

	timelineID, err := utils.ReadIDParam(r, "timelineId")
	if err != nil {
		eh.logger.Printf("Invalid timeline ID: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid timeline ID"})
		return
	}

	var req AIEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		eh.logger.Printf("Failed to decode request: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid request payload"})
		return
	}

	if req.Prompt == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Prompt is required"})
		return
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		eh.logger.Printf("Failed to create GeminiAI client: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to create GeminiAI client"})
		return
	}

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"title": {
						Type:        genai.TypeString,
						Description: "The main date or time marker for the event, like 'January 2022' 'Week 1', 'Month 2-3' etc.",
					},
					"cardTitle": {
						Type:        genai.TypeString,
						Description: "A short, concise title for the timeline card.",
					},
					"cardSubtitle": {
						Type:        genai.TypeString,
						Description: "A brief, one-sentence subtitle for the event.",
					},
					"cardDetailedText": {
						Type:        genai.TypeString,
						Description: "A detailed, paragraph-length description of the event that occurred.",
					},
				},

				PropertyOrdering: []string{"title", "cardTitle", "cardSubtitle", "cardDetailedText"},
			},
		},
	}
	response, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(
			req.Prompt,
		),
		config,
	)

	if err != nil {
		eh.logger.Printf("Failed to generate content: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to generate content"})
		return
	}

	var timelineEvents []TimelineEventRequest

	if err := json.Unmarshal([]byte(response.Candidates[0].Content.Parts[0].Text), &timelineEvents); err != nil {
		eh.logger.Printf("Failed to decode request: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid request payload"})
		return
	}

	var createParams []db.BulkCreateEventsParams
	for _, event := range timelineEvents {

		createParams = append(createParams, db.BulkCreateEventsParams{
			TimelineID:       timelineID,
			Title:            event.Title,
			CardTitle:        event.CardTitle,
			CardSubtitle:     event.CardSubtitle,
			CardDetailedText: event.CardDetailedText,
		})

	}

	_, err = eh.eventStore.BulkCreateEvents(r.Context(), createParams)
	if err != nil {
		eh.logger.Printf("Failed to create events: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to create events"})
		return
	}

	events, err := eh.eventStore.GetEventsByTimelineId(r.Context(), timelineID)
	if err != nil {
		eh.logger.Printf("Failed to retrieve events: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to retrieve events"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"events": events})

}
