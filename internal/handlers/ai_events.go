package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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
	TimeMarker       string      `json:"time_marker"`
	Title            string      `json:"title"`
	Subtitle         pgtype.Text `json:"subtitle"`
	Description      pgtype.Text `json:"description"`
	Media            *Media      `json:"media"`
}

func boolPtr(b bool) *bool { return &b }

func toPgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

func extractMediaFields(media *Media) (name, mediaType, url pgtype.Text) {
	if media == nil {
		return pgtype.Text{Valid: false}, pgtype.Text{Valid: false}, pgtype.Text{Valid: false}
	}
	return toPgText(media.Name), toPgText(media.Type), toPgText(media.Source.URL)
}

var apiKey = os.Getenv("GEMINI_API_KEY")

type geminiClient struct {
	client *genai.Client
	ctx    context.Context
}

func newGeminiClient(ctx context.Context) (*geminiClient, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	return &geminiClient{client: client, ctx: ctx}, nil
}

func (gc *geminiClient) generateEvents(prompt string) ([]TimelineEventRequest, error) {
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"time_marker": {
						Type:        genai.TypeString,
						Description: "The main date or time marker for the event, like 'January 2022' 'Week 1', 'Month 2-3' etc.",
					},
					"title": {
						Type:        genai.TypeString,
						Description: "A short, concise title for the timeline card.",
					},
					"media": {
						Type:        genai.TypeObject,
						Nullable:    boolPtr(true),
						Description: "Optional, only return media for relevant events.",
						Properties: map[string]*genai.Schema{
							"name": {
								Type: genai.TypeString,
							},
							"type": {
								Type: genai.TypeString,
								Enum: []string{"IMAGE"},
							},
							"source": {
								Type: genai.TypeObject,
								Properties: map[string]*genai.Schema{
									"url": {
										Type:        genai.TypeString,
										Description: "Give me a valid, direct image URL that actually loads in a browser (no 404).",
									},
								},
							},
						},
					},
					"subtitle": {
						Type:        genai.TypeString,
						Description: "A brief, one-sentence subtitle for the event.",
					},
					"description": {
						Type:        genai.TypeString,
						Description: "A detailed, paragraph-length description of the event that occurred.",
					},
				},
				PropertyOrdering: []string{"time_marker", "title", "media", "subtitle", "description"},
			},
		},
	}

	resp, err := gc.client.Models.GenerateContent(
		gc.ctx,
		"gemini-2.5-flash",
		genai.Text(prompt),
		config,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	var events []TimelineEventRequest
	if err := json.Unmarshal([]byte(resp.Candidates[0].Content.Parts[0].Text), &events); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return events, nil
}

func (eh *EventHandler) HandleCreateAIEvents(w http.ResponseWriter, r *http.Request) {
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

	gc, err := newGeminiClient(r.Context())
	if err != nil {
		eh.logger.Printf("Client creation failed: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to initialize AI service"})
		return
	}

	events, err := gc.generateEvents(req.Prompt)
	if err != nil {
		eh.logger.Printf("Event generation failed: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to generate events"})
		return
	}

	res, err := eh.eventStore.GetMaxPositionByTimelineId(r.Context(), timelineID)
	if err != nil {
		eh.logger.Printf("Failed to get max position: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to determine event position"})
		return
	}

	maxPos := 0
	if res != nil {
		maxPos = int(res.(int32))
	}

	var createParams []db.BulkCreateEventsParams
	for _, event := range events {
		mediaName, mediaType, mediaUrl := extractMediaFields(event.Media)

		maxPos++
		createParams = append(createParams, db.BulkCreateEventsParams{
			TimelineID:       timelineID,
			TimeMarker:       event.TimeMarker,
			Title:            event.Title,
			Subtitle:         event.Subtitle,
			Description:      event.Description,
			MediaName:        mediaName,
			MediaType:        mediaType,
			MediaUrl:         mediaUrl,
			Position: pgtype.Int4{
				Int32: int32(maxPos),
				Valid: true,
			},
		})
	}

	if _, err := eh.eventStore.BulkCreateEvents(r.Context(), createParams); err != nil {
		eh.logger.Printf("Failed to create events: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to create events"})
		return
	}

	savedEvents, err := eh.eventStore.GetEventsByTimelineId(r.Context(), timelineID)
	if err != nil {
		eh.logger.Printf("Failed to retrieve events: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to retrieve events"})
		return
	}

	var responseEvents []any
	for _, ev := range savedEvents {
		responseEvents = append(responseEvents, utils.MapDBEventToFrontend(ev))
	}
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"events": responseEvents})
}

func (eh *EventHandler) HandleUpdateAIEvents(w http.ResponseWriter, r *http.Request) {
	timelineID, err := utils.ReadIDParam(r, "timelineId")
	if err != nil {
		eh.logger.Printf("Invalid timeline ID: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid timeline ID"})
		return
	}

	userID, err := utils.ReadUserIDFromContext(r, "userID")
	if err != nil {
		eh.logger.Printf("Invalid user ID: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"message": "Invalid user ID"})
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

	currentEvents, err := eh.eventStore.GetEventsByTimelineId(r.Context(), timelineID)
	if err != nil {
		eh.logger.Printf("Failed to retrieve events: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to retrieve events"})
		return
	}

	currentEventsJSON, err := json.Marshal(currentEvents)
	if err != nil {
		eh.logger.Printf("Failed to marshal current events: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to process current events"})
		return
	}

	fullPrompt := fmt.Sprintf(`I have an existing timeline with the following data:
%s

The user wants to modify this timeline with this instruction: "%s".

Please output the NEW full list of events based on these changes.
- If the user asks to add an event, insert it in the correct chronological order.
- If the user asks to delete, remove it.
- If the user asks to edit, change the specific fields.
- Keep existing events if they are not affected.`, string(currentEventsJSON), req.Prompt)

	gc, err := newGeminiClient(r.Context())
	if err != nil {
		eh.logger.Printf("Client creation failed: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to initialize AI service"})
		return
	}

	newEvents, err := gc.generateEvents(fullPrompt)
	if err != nil {
		eh.logger.Printf("Event generation failed: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to generate updated events"})
		return
	}

	_, err = eh.eventStore.DeleteEventsByTimelineId(r.Context(), db.DeleteEventsByTimelineIdParams{
		TimelineID: timelineID,
		UserID:     userID,
	})
	if err != nil {
		eh.logger.Printf("Failed to delete events: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to delete events"})
		return
	}

	if len(newEvents) > 0 {
		var createParams []db.BulkCreateEventsParams
		for i, event := range newEvents {
			mediaName, mediaType, mediaUrl := extractMediaFields(event.Media)

			createParams = append(createParams, db.BulkCreateEventsParams{
				TimelineID:       timelineID,
				TimeMarker:       event.TimeMarker,
				Title:            event.Title,
				Subtitle:         event.Subtitle,
				Description:      event.Description,
				MediaName:        mediaName,
				MediaType:        mediaType,
				MediaUrl:         mediaUrl,
				Position: pgtype.Int4{
					Int32: int32(i + 1),
					Valid: true,
				},
			})
		}

		if _, err := eh.eventStore.BulkCreateEvents(r.Context(), createParams); err != nil {
			eh.logger.Printf("Failed to insert new events: %v", err)
			utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to save new events"})
			return
		}
	}

	updatedEvents, err := eh.eventStore.GetEventsByTimelineId(r.Context(), timelineID)
	if err != nil {
		eh.logger.Printf("Failed to retrieve updated events: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"message": "Failed to retrieve updated events"})
		return
	}

	var responseEvents []any
	for _, ev := range updatedEvents {
		responseEvents = append(responseEvents, utils.MapDBEventToFrontend(ev))
	}
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"events": responseEvents})
}
