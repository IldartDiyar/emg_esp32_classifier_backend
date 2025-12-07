package svc

import (
	"bytes"
	"context"
	"emg_esp32_classifier_backend/internal/repo"
	"emg_esp32_classifier_backend/pkg/cerrors"
	"emg_esp32_classifier_backend/pkg/dto"
	"emg_esp32_classifier_backend/pkg/models"
	"emg_esp32_classifier_backend/pkg/sessions"
	"emg_esp32_classifier_backend/pkg/utils"
	"encoding/csv"
	"log"
	"strconv"
	"time"
)

type Service struct {
	repo    repo.Repository
	session *sessions.SessionManager
}

func NewService(repo repo.Repository) *Service {
	return &Service{
		repo:    repo,
		session: sessions.NewSessionManager(),
	}
}

// from frontend
func (s *Service) WSStartTraining(ctx context.Context, msg models.WsFrontendToBackend) (*models.WsBackendToEsp, error) {
	if msg.Rep > 5 {
		return nil, cerrors.ErrIncorrectRep
	}

	if _, err := s.repo.GetMovementsById(ctx, msg.MovementID); err != nil {
		return nil, err
	}

	dev, err := s.repo.GetDeviceById(ctx, msg.DeviceID)
	if err != nil {
		return nil, err
	}

	if dev.Status == dto.DeviceStatusStreaming {
		return nil, cerrors.ErrDeviceBusy
	}

	ss, exists := s.session.Get(msg.DeviceID)

	if !exists {
		if msg.Rep != 1 {
			return nil, cerrors.ErrIncorrectRep
		}

		var tID int
		tID, err = s.repo.CreateTraining(ctx, msg.DeviceID, msg.MovementID, msg.Rep)
		if err != nil {
			return nil, err
		}

		ss = &sessions.Session{
			TrainingID: tID,
			Rep:        msg.Rep,
			MovementID: msg.MovementID,
			DeviceID:   msg.DeviceID,
		}

		s.session.Set(msg.DeviceID, ss)
	} else {
		if ss.MovementID != msg.MovementID {
			return nil, cerrors.ErrMovementNotAllowed
		}

		if msg.Rep != ss.Rep+1 {
			return nil, cerrors.ErrIncorrectRep
		}

		s.session.Update(msg.DeviceID, func(sx *sessions.Session) {
			sx.Rep = msg.Rep
			sx.TrainingID = ss.TrainingID
		})
	}

	return &models.WsBackendToEsp{
		Event:      models.EventESPStartRawStream,
		Duration:   models.DefaultDurationOfTraining,
		ServerTime: time.Now().UnixMilli(),
	}, nil
}

// from esp
func (s *Service) WSRawStream(ctx context.Context, msg models.WsEspToBackend, deviceId int) (*models.WsBackendToFrontend, error) {
	ss, ex := s.session.Get(deviceId)
	if !ex {
		return nil, cerrors.ErrSomethingWentWrong
	}

	var event models.Event

	raw := []models.RawSample{}

	switch msg.Event {
	case models.EventRawStreamBegin:
		event = models.EventTrainingStarted
		if err := s.repo.UpdateDeviceStatus(ctx, deviceId, dto.DeviceStatusStreaming); err != nil {
			log.Printf("[RawStream][EventRawStreamBegin][UpdateDeviceStatus]: %v\n", err)
		}
		raw = nil
	case models.EventRawStreamInProc:
		event = models.EventTrainingRawData
		err := s.repo.InsertTrainingRaw(ctx, dto.MapWsToTrainingRaw(msg.Raw, msg.Timestamp, ss))
		if err != nil {
			return nil, err
		}

		raw, err = s.repo.SelectTrainingRawSamples(ctx, ss.TrainingID, ss.DeviceID)
		if err != nil {
			return nil, err
		}

		if err = s.repo.UpdateDeviceStatus(ctx, deviceId, dto.DeviceStatusStreaming); err != nil {
			log.Printf("[RawStream][EventRawStreamBegin][UpdateDeviceStatus]: %v\n", err)
		}

	case models.EventRawStreamFinish:
		event = models.EventTrainingCompleted
		if ss.Rep == 5 {
			defer s.session.Delete(deviceId)
			if err := s.repo.UpdateDeviceStatus(ctx, deviceId, dto.DeviceStatusIdle); err != nil {
				log.Printf("[RawStream][EventRawStreamFinish][UpdateDeviceStatus]: %v", err)
			}

			if err := s.repo.DeleteTraining(ctx, ss.TrainingID); err != nil {
				log.Printf("[RawStream][EventRawStreamFinish][DeleteTraining]: %v\n", err)
			}
		}

		if ss.Rep < 5 {
			if err := s.repo.UpdateDeviceStatus(ctx, deviceId, dto.DeviceStatusReserved); err != nil {
				log.Printf("[RawStream][EventRawStreamBegin][UpdateDeviceStatus]: %v\n", err)
			}
		}

		raw = nil
	}

	return &models.WsBackendToFrontend{
		Event:      event,
		DeviceID:   deviceId,
		MovementID: ss.MovementID,
		Rep:        ss.Rep,
		Raw:        raw,
	}, nil
}

func (s *Service) RegisterDevice(ctx context.Context, deviceName string) (int, error) {
	dev, err := s.repo.InsertDevice(ctx, deviceName)
	if err != nil {
		return 0, err
	}

	return dev.ID, nil
}

func (s *Service) ReserveDevice(ctx context.Context, deviceId int) error {
	dev, err := s.repo.GetDeviceById(ctx, deviceId)

	if err != nil {
		return err
	}

	if dev.Status == dto.DeviceStatusStreaming {
		return cerrors.ErrDeviceBusy
	}

	if err = s.repo.UpdateDeviceStatus(ctx, deviceId, dto.DeviceStatusReserved); err != nil {
		return err
	}

	return nil

}

func (s *Service) GetDeviceList(ctx context.Context) ([]dto.Device, error) {
	dev, err := s.repo.ListDevices(ctx)
	if err != nil {
		return nil, err
	}

	return dev, nil
}

func (s *Service) WSStartStreaming(ctx context.Context, msg models.WsFrontendToBackend) (*models.WsBackendToEsp, error) {
	dev, err := s.repo.GetDeviceById(ctx, msg.DeviceID)
	if err != nil {
		return nil, err
	}

	if dev.Status == dto.DeviceStatusStreaming {
		return nil, cerrors.ErrDeviceBusy
	}

	if err := s.repo.UpdateDeviceStatus(ctx, msg.DeviceID, dto.DeviceStatusStreaming); err != nil {
		return nil, err
	}

	return &models.WsBackendToEsp{
		Event:      models.EventESPStartRawStream, // можно завести отдельный EventESPStartLiveStream
		Duration:   models.DefaultDurationOfTraining * 60,
		ServerTime: time.Now().UnixMilli(),
	}, nil
}

func (s *Service) WSStopStreaming(ctx context.Context, deviceID int) (*models.WsBackendToEsp, error) {
	if err := s.repo.UpdateDeviceStatus(ctx, deviceID, dto.DeviceStatusIdle); err != nil {
		return nil, err
	}

	return &models.WsBackendToEsp{
		Event: models.EventESPStopRawStream,
	}, nil
}

func (s *Service) GetMovements(ctx context.Context) ([]dto.Movements, error) {
	movs, err := s.repo.GetMovements(ctx)
	if err != nil {
		return nil, err
	}
	return movs, nil
}

func (s *Service) GetTrainingRawCSV(ctx context.Context) ([]byte, error) {
	// Получаем TrainingRaw записи (не RawSample!)
	rows, err := s.repo.GetAllRawData(ctx)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	header := []string{
		"id",
		"training_id",
		"device_id",
		"movement_id",
		"repetition",
		"timestamp",
		"raw",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	for _, r := range rows {
		rawDecoded := utils.DecodeRawBytes(r.Raw)

		row := []string{
			strconv.Itoa(r.ID),
			strconv.Itoa(r.TrainingID),
			strconv.Itoa(r.DeviceID),
			strconv.Itoa(r.MovementID),
			strconv.Itoa(r.Repetition),
			r.TS.Format(time.RFC3339),
			utils.IntSliceToString(rawDecoded),
		}

		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
