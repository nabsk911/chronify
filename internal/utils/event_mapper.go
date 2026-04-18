package utils

import "github.com/nabsk911/chronify/internal/db"

func MapDBEventToFrontend(dbEvent db.Event) map[string]any {
	return map[string]any{
		"id":                 dbEvent.ID,
		"time_marker":        dbEvent.TimeMarker,
		"title":              dbEvent.Title,
		"subtitle":           dbEvent.Subtitle.String,
		"description":        dbEvent.Description.String,
		"media": map[string]any{
			"name": dbEvent.MediaName.String,
			"type": dbEvent.MediaType.String,
			"source": map[string]string{
				"url": dbEvent.MediaUrl.String,
			},
		},
	}
}
