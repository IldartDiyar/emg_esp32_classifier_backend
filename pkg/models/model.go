package models

import "encoding/json"

type Event string

const (
	// RAW (ESP32 → Backend)
	EventRawStream Event = "raw_stream"

	// Training (ESP32 → Backend)
	EventTrainingRepStarted  Event = "training_repetition_started"
	EventTrainingRepFinished Event = "training_repetition_finished"
	EventTrainingCompleted   Event = "training_completed"

	// Frontend → Backend (HTTP actions)
	EventStartTraining     Event = "start_training"
	EventStopTraining      Event = "stop_training"
	EventStartVerification Event = "start_verification"
	EventStopVerification  Event = "stop_verification"
	EventStartStreaming    Event = "start_streaming"
	EventStopStreaming     Event = "stop_streaming"

	// Backend → Frontend
	EventTrainingStatus         Event = "training_status"
	EventVerificationPrediction Event = "verification_prediction"
	EventGesturePrediction      Event = "gesture_prediction"

	// Feedback (Frontend → Backend)
	EventVerificationFeedback Event = "verification_feedback"
)

type EventMessage interface {
	GetEvent() Event
}

type FrontendAction struct {
	Event      Event `json:"event"`
	MovementID int   `json:"movement_id,omitempty"`
}

func (m FrontendAction) GetEvent() Event { return m.Event }

type ESPCommand struct {
	Event      Event `json:"event"`
	MovementID int   `json:"movement_id,omitempty"`
}

func (m ESPCommand) GetEvent() Event { return m.Event }

type RawStream struct {
	Event     Event `json:"event"` // raw_stream
	Timestamp int64 `json:"timestamp"`
	Raw       []int `json:"raw"`
}

func (m RawStream) GetEvent() Event { return m.Event }

type TrainingRepetitionStarted struct {
	Event      Event `json:"event"`
	MovementID int   `json:"movement_id"`
	Repetition int   `json:"repetition"`
}

func (m TrainingRepetitionStarted) GetEvent() Event { return m.Event }

type TrainingRepetitionFinished struct {
	Event            Event `json:"event"`
	MovementID       int   `json:"movement_id"`
	Repetition       int   `json:"repetition"`
	DurationMs       int   `json:"duration_ms"`
	SamplesCollected int   `json:"samples_collected"`
	Timestamp        int64 `json:"timestamp"`
}

func (m TrainingRepetitionFinished) GetEvent() Event { return m.Event }

type TrainingCompleted struct {
	Event       Event `json:"event"`
	MovementID  int   `json:"movement_id"`
	Repetitions int   `json:"repetitions"`
	Timestamp   int64 `json:"timestamp"`
}

func (m TrainingCompleted) GetEvent() Event { return m.Event }

type TrainingStatus struct {
	Event       Event `json:"event"`
	MovementID  int   `json:"movement_id"`
	Repetition  int   `json:"repetition"`
	SecondsLeft int   `json:"seconds_left"`
}

func (m TrainingStatus) GetEvent() Event { return m.Event }

type VerificationPrediction struct {
	Event             Event   `json:"event"`
	PredictedMovement int     `json:"predicted_movement"`
	Confidence        float64 `json:"confidence"`
	Timestamp         int64   `json:"timestamp"`
}

func (m VerificationPrediction) GetEvent() Event { return m.Event }

type GesturePrediction struct {
	Event       Event   `json:"event"`
	GestureID   int     `json:"gesture_id"`
	GestureName string  `json:"gesture_name"`
	Confidence  float64 `json:"confidence"`
	Timestamp   int64   `json:"timestamp"`
}

func (m GesturePrediction) GetEvent() Event { return m.Event }

type VerificationFeedback struct {
	Event         Event `json:"event"`
	MovementID    int   `json:"movement_id"`
	UserConfirmed bool  `json:"user_confirmed"`
}

func (m VerificationFeedback) GetEvent() Event { return m.Event }

// just wrapper
type WSMessage struct {
	Event     Event           `json:"event"`
	Timestamp int64           `json:"timestamp,omitempty"`
	Raw       []int           `json:"raw,omitempty"`
	Body      json.RawMessage `json:"body,omitempty"`
}

func (m WSMessage) GetEvent() Event { return m.Event }
