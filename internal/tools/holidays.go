package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/openai/openai-go/v2"
)

// LoadCalendar loads calendar events from the specified URL.
func LoadCalendar(ctx context.Context, link string) ([]*ics.VEvent, error) {
	slog.InfoContext(ctx, "Loading calendar", "link", link)

	cal, err := ics.ParseCalendarFromUrl(link, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse calendar: %w", err)
	}

	return cal.Events(), nil
}

// HolidaysTool provides information about public holidays.
type HolidaysTool struct{}

// Name returns the tool's identifier.
func (h *HolidaysTool) Name() string {
	return "get_holidays"
}

// Description returns what the tool does.
func (h *HolidaysTool) Description() string {
	return "Gets local bank and public holidays. Each line is a single holiday in the format 'YYYY-MM-DD: Holiday Name'."
}

// Parameters defines the tool's input schema.
func (h *HolidaysTool) Parameters() openai.FunctionParameters {
	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"before_date": map[string]string{
				"type":        "string",
				"description": "Optional date in RFC3339 format to get holidays before this date. If not provided, all holidays will be returned.",
			},
			"after_date": map[string]string{
				"type":        "string",
				"description": "Optional date in RFC3339 format to get holidays after this date. If not provided, all holidays will be returned.",
			},
			"max_count": map[string]string{
				"type":        "integer",
				"description": "Optional maximum number of holidays to return. If not provided, all holidays will be returned.",
			},
		},
	}
}

// Execute retrieves holiday information based on the provided criteria.
func (h *HolidaysTool) Execute(ctx context.Context, args string) (string, error) {
	link := "https://www.officeholidays.com/ics/spain/catalonia"
	if v := os.Getenv("HOLIDAY_CALENDAR_LINK"); v != "" {
		link = v
	}

	events, err := LoadCalendar(ctx, link)
	if err != nil {
		return "failed to load holiday events", nil
	}

	var payload struct {
		BeforeDate time.Time `json:"before_date,omitempty"`
		AfterDate  time.Time `json:"after_date,omitempty"`
		MaxCount   int       `json:"max_count,omitempty"`
	}

	if err := json.Unmarshal([]byte(args), &payload); err != nil {
		return "failed to parse tool call arguments: " + err.Error(), nil
	}

	var holidays []string
	for _, event := range events {
		date, err := event.GetAllDayStartAt()
		if err != nil {
			continue
		}

		if payload.MaxCount > 0 && len(holidays) >= payload.MaxCount {
			break
		}

		if !payload.BeforeDate.IsZero() && date.After(payload.BeforeDate) {
			continue
		}

		if !payload.AfterDate.IsZero() && date.Before(payload.AfterDate) {
			continue
		}

		holidays = append(holidays, date.Format(time.DateOnly)+": "+event.GetProperty(ics.ComponentPropertySummary).Value)
	}

	return strings.Join(holidays, "\n"), nil
}