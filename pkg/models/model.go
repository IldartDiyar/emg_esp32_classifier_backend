package models

import (
	"errors"
)

type Event string

var ErrMovementNotAllowed = errors.New("movement is not allowed: previous movement is not finished")

const DefaultDurationOfTraining = 5

const (

	// Frontend to backend
	EventStartTraining  Event = "start_training"
	EventStartStreaming Event = "start_streaming"
	EventStopTraining   Event = "stop"

	// Backend to esp
	EventESPStartRawStream Event = "raw_stream"
	EventESPStopRawStream  Event = "stop_raw_stream"

	// Esp to backend
	HandShake            Event = "handshake"
	EventRawStreamBegin  Event = "raw_stream_begin" // first packet
	EventRawStreamInProc Event = "raw_stream_in_process"
	EventRawStreamFinish Event = "raw_stream_finish"

	// backend to frontend
	EventTrainingStarted   Event = "training_started"
	EventTrainingRawData   Event = "training_raw_data"
	EventTrainingCompleted Event = "start_training_completed"
	EventStreamingData     Event = "streaming_data"
)

type WsBackendToFrontend struct {
	Event      Event       `json:"event"`
	DeviceID   int         `json:"device_id"`
	MovementID int         `json:"movement_id,omitempty"`
	Rep        int         `json:"rep,omitempty"`
	Message    string      `json:"message"`
	Raw        []RawSample `json:"raw,omitempty"`
	ClassID    int         `json:"class_id,omitempty"`
	ClassName  string      `json:"class_name,omitempty"`
	Prob       []float64   `json:"prob,omitempty"`
}

type RawSample struct {
	Timestamp string `json:"timestamp"`
	Raw       []int  `json:"raw"`
}

type WsFrontendToBackend struct {
	Event      Event `json:"event"`
	DeviceID   int   `json:"device_id"`
	MovementID int   `json:"movement_id,omitempty"`
	Rep        int   `json:"rep,omitempty"`
}

type WsEspToBackend struct {
	Event      Event  `json:"event"`
	DeviceName string `json:"device_name"`
	Timestamp  string `json:"timestamp"`
	Raw        []int  `json:"raw"`
}

type WsBackendToEsp struct {
	Event      Event `json:"event"`
	Duration   int   `json:"duration"`
	ServerTime int64 `json:"server_time"`
}
