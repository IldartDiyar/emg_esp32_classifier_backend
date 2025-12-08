package dto

import (
	"emg_esp32_classifier_backend/pkg/sessions"
	"emg_esp32_classifier_backend/pkg/utils"
	"strconv"
	"time"
)

type TrainingRaw struct {
	ID         int       `db:"id" json:"id"`
	TrainingID int       `db:"training_id" json:"training_id"`
	DeviceID   int       `db:"device_id" json:"device_id"`
	MovementID int       `db:"movement_id" json:"movement_id"`
	Repetition int       `db:"repetition" json:"repetition"`
	TS         time.Time `db:"ts" json:"timestamp"`
	Raw        []byte    `db:"raw" json:"raw"` // BYTEA
}

type DeviceStatus string

const (
	DeviceStatusIdle      DeviceStatus = "idle"
	DeviceStatusStreaming DeviceStatus = "streaming"
	DeviceStatusReserved  DeviceStatus = "reserved"
)

type Device struct {
	ID       int          `json:"id"`
	Name     string       `json:"name"`
	Status   DeviceStatus `json:"status"`
	LastSeen time.Time    `json:"last_seen"`
}

type Movements struct {
	Movement_id int    `json:"movement_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TrainingSummary struct {
	TrainingID int `json:"training_id"`
	DeviceID   int `json:"device_id"`
	MovementID int `json:"movement_id"`
	Reps       int `json:"reps"`
	Samples    int `json:"samples"`
}

func MapWsToTrainingRaw(rawSlice []int, espTs string, session *sessions.Session) *TrainingRaw {
	rawBytes := utils.IntSliceToBytea(rawSlice)

	tsInt, err := strconv.ParseInt(espTs, 10, 64)
	if err != nil {
		tsInt = time.Now().UnixNano()
	}

	sec := tsInt / 1_000_000_000
	nsec := tsInt % 1_000_000_000

	ts := time.Unix(sec, nsec).UTC()

	return &TrainingRaw{
		TrainingID: session.TrainingID,
		DeviceID:   session.DeviceID,
		MovementID: session.MovementID,
		Repetition: session.Rep,
		TS:         ts,
		Raw:        rawBytes,
	}
}
