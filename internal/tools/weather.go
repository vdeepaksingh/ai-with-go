package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/openai/openai-go/v2"
)

// WeatherTool provides weather information for specified locations.
type WeatherTool struct{}

// Name returns the tool's identifier.
func (w *WeatherTool) Name() string {
	return "get_weather"
}

// Description returns what the tool does.
func (w *WeatherTool) Description() string {
	return "Get current weather and forecast information for a specified location"
}

// Parameters defines the tool's input schema.
func (w *WeatherTool) Parameters() openai.FunctionParameters {
	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"location": map[string]any{
				"type":        "string",
				"description": "City name, coordinates, or location query",
			},
			"forecast": map[string]any{
				"type":        "boolean",
				"description": "Include forecast information (optional)",
			},
		},
		"required": []string{"location"},
	}
}

type weatherRequest struct {
	Location string `json:"location"`
	Forecast bool   `json:"forecast,omitempty"`
}

type weatherResponse struct {
	Location struct {
		Name    string `json:"name"`
		Country string `json:"country"`
	} `json:"location"`
	Current struct {
		TempC     float64 `json:"temp_c"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
		WindKph  float64 `json:"wind_kph"`
		Humidity int     `json:"humidity"`
		FeelsC   float64 `json:"feelslike_c"`
	} `json:"current"`
	Forecast struct {
		ForecastDay []struct {
			Date string `json:"date"`
			Day  struct {
				MaxTempC  float64 `json:"maxtemp_c"`
				MinTempC  float64 `json:"mintemp_c"`
				Condition struct {
					Text string `json:"text"`
				} `json:"condition"`
			} `json:"day"`
		} `json:"forecastday"`
	} `json:"forecast"`
}

// Execute retrieves weather information for the specified location.
func (w *WeatherTool) Execute(ctx context.Context, args string) (string, error) {
	var req weatherRequest
	if err := json.Unmarshal([]byte(args), &req); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		return "Weather service unavailable: API key not configured", nil
	}

	// Build API URL
	baseURL := "http://api.weatherapi.com/v1/current.json"
	if req.Forecast {
		baseURL = "http://api.weatherapi.com/v1/forecast.json"
	}

	u, _ := url.Parse(baseURL)
	q := u.Query()
	q.Set("key", apiKey)
	q.Set("q", req.Location)
	q.Set("aqi", "no")
	if req.Forecast {
		q.Set("days", "3")
	}
	u.RawQuery = q.Encode()

	// Make HTTP request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(u.String())
	if err != nil {
		return "", fmt.Errorf("weather API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("Weather information unavailable for '%s'", req.Location), nil
	}

	var weather weatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		return "", fmt.Errorf("failed to parse weather response: %w", err)
	}

	// Format response
	var result strings.Builder
	fmt.Fprintf(&result, "Weather in %s, %s:\n", weather.Location.Name, weather.Location.Country)
	fmt.Fprintf(&result, "Temperature: %.1f°C (feels like %.1f°C)\n", weather.Current.TempC, weather.Current.FeelsC)
	fmt.Fprintf(&result, "Condition: %s\n", weather.Current.Condition.Text)
	fmt.Fprintf(&result, "Wind: %.1f km/h\n", weather.Current.WindKph)
	fmt.Fprintf(&result, "Humidity: %d%%", weather.Current.Humidity)

	if req.Forecast && len(weather.Forecast.ForecastDay) > 0 {
		result.WriteString("\n\nForecast:\n")
		for _, day := range weather.Forecast.ForecastDay {
			fmt.Fprintf(&result, "%s: %s, %.1f°C - %.1f°C\n",
				day.Date, day.Day.Condition.Text, day.Day.MinTempC, day.Day.MaxTempC)
		}
	}

	return result.String(), nil
}