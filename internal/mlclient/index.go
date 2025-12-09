package mlclient

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type PredictRequest struct {
	Features []float64 `json:"features"`
}

type PredictResponse struct {
	ClassID       int       `json:"class_id"`
	ClassName     string    `json:"class_name"`
	Probabilities []float64 `json:"probabilities"`
}

type Client struct {
	URL        string
	HTTPClient *http.Client
}

func New(url string) *Client {
	return &Client{
		URL: url,
		HTTPClient: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

func (c *Client) Predict(features []float64) (*PredictResponse, error) {
	reqPayload := PredictRequest{Features: features}

	body, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Post(c.URL+"/predict", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out PredictResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return &out, nil
}
