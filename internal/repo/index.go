package repo

import (
	"context"
	"database/sql"
	"emg_esp32_classifier_backend/pkg/cerrors"
	"emg_esp32_classifier_backend/pkg/models"
	"emg_esp32_classifier_backend/pkg/utils"
	"errors"
	"time"

	"emg_esp32_classifier_backend/pkg/dto"
)

var ErrNotIdle = errors.New("device is not idle")

type Repository interface {
	GetMovements(ctx context.Context) ([]dto.Movements, error)
	GetMovementsById(ctx context.Context, MovementID int) (*dto.Movements, error)

	ListDevices(ctx context.Context) ([]dto.Device, error)
	GetDeviceById(ctx context.Context, DeviceID int) (*dto.Device, error)
	UpdateDeviceStatus(ctx context.Context, deviceID int, status dto.DeviceStatus) error
	InsertDevice(ctx context.Context, name string) (*dto.Device, error)

	CreateTraining(ctx context.Context, deviceID, movementID, rep int) (int, error)
	UpdateTrainingRepetition(ctx context.Context, trainingID, rep int) error
	MarkTrainingFinished(ctx context.Context, trainingID int) error
	DeleteTraining(ctx context.Context, trainingID int) error

	InsertTrainingRaw(ctx context.Context, tr *dto.TrainingRaw) error
	SelectTrainingRawSamples(ctx context.Context, trainingID, deviceID int) ([]models.RawSample, error)
	GetAllRawData(ctx context.Context) ([]dto.TrainingRaw, error)
}

type pgRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) GetMovements(ctx context.Context) ([]dto.Movements, error) {
	const q = ` SELECT movement_id, name, description FROM movements`

	var movements []dto.Movements

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var movement dto.Movements
		err := rows.Scan(&movement.Movement_id, &movement.Name, &movement.Description)
		if err != nil {
			return nil, err
		}
		movements = append(movements, movement)
	}

	return movements, nil
}
func (r *pgRepository) GetMovementsById(ctx context.Context, MovementID int) (*dto.Movements, error) {
	const q = `SELECT movement_id, name, description FROM movements WHERE movement_id = $1`

	var movement dto.Movements

	err := r.db.QueryRowContext(ctx, q, MovementID).Scan(
		&movement.Movement_id,
		&movement.Name,
		&movement.Description,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, cerrors.ErrNotFound
		}
		return nil, err
	}

	return &movement, nil
}

// ---- Devices ----

func (r *pgRepository) ListDevices(ctx context.Context) ([]dto.Device, error) {
	const q = `
	SELECT id, name, status, last_seen
	FROM devices
	ORDER BY id;`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []dto.Device
	for rows.Next() {
		var d dto.Device
		if err := rows.Scan(&d.ID, &d.Name, &d.Status, &d.LastSeen); err != nil {
			return nil, err
		}

		res = append(res, d)
	}
	return res, rows.Err()
}

func (r *pgRepository) GetDeviceById(ctx context.Context, DeviceID int) (*dto.Device, error) {
	const q = `SELECT id, name, status, last_seen FROM devices WHERE id = $1`

	var d dto.Device

	err := r.db.QueryRowContext(ctx, q, DeviceID).Scan(&d.ID, &d.Name, &d.Status, &d.LastSeen)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, cerrors.ErrNotFound
		}
		return nil, err
	}

	return &d, nil
}

func (r *pgRepository) UpdateDeviceStatus(ctx context.Context, deviceID int, status dto.DeviceStatus) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const q = `
	UPDATE 
	    devices
	SET 
	    status = $2, last_seen = NOW()
	WHERE 
	    id = $1;
`
	_, err = tx.ExecContext(ctx, q, deviceID, string(status))
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}

	return err
}

func (r *pgRepository) InsertDevice(ctx context.Context, name string) (*dto.Device, error) {
	const q = `
		INSERT INTO devices (name, status)
		VALUES ($1, 'idle')
		RETURNING id, name, status, last_seen
	`

	var d dto.Device
	err := r.db.QueryRowContext(ctx, q, name).
		Scan(&d.ID, &d.Name, &d.Status, &d.LastSeen)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

// ---- Training ----
func (r *pgRepository) CreateTraining(ctx context.Context, deviceID, movementID, rep int) (int, error) {
	const q = `
	INSERT INTO training 
	    (device_id, movement_id, repetition)
	VALUES 
	    ($1, $2, $3)
	RETURNING id;
	`
	var id int
	if err := r.db.QueryRowContext(ctx, q, deviceID, movementID, rep).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *pgRepository) UpdateTrainingRepetition(ctx context.Context, trainingID, rep int) error {
	const q = `
	UPDATE 
	    training
	SET 
	    repetition = $2
	WHERE 
	    id = $1;
	`
	_, err := r.db.ExecContext(ctx, q, trainingID, rep)
	return err
}

func (r *pgRepository) MarkTrainingFinished(ctx context.Context, trainingID int) error {
	const q = `
	UPDATE 
	    training
	SET 
	    finished = true
	WHERE 
	    id = $1;
	`
	_, err := r.db.ExecContext(ctx, q, trainingID)
	return err
}

func (r *pgRepository) DeleteTraining(ctx context.Context, trainingID int) error {
	const q = `
	DELETE FROM 
        training 
    WHERE 
        id = $1;`
	_, err := r.db.ExecContext(ctx, q, trainingID)
	return err
}

// ---- Training Raw ----

func (r *pgRepository) InsertTrainingRaw(ctx context.Context, tr *dto.TrainingRaw) error {
	const q = `
	INSERT INTO training_raw (training_id, device_id, movement_id, repetition, ts, raw)
	VALUES ($1, $2, $3, $4, $5, $6);
	`
	_, err := r.db.ExecContext(
		ctx,
		q,
		tr.TrainingID,
		tr.DeviceID,
		tr.MovementID,
		tr.Repetition,
		tr.TS,
		tr.Raw,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *pgRepository) SelectTrainingRawSamples(ctx context.Context, trainingID, deviceID int) ([]models.RawSample, error) {
	const q = `
	SELECT ts, raw
	FROM training_raw
	WHERE training_id = $1 AND device_id = $2
	ORDER BY ts;
	`

	rows, err := r.db.QueryContext(ctx, q, trainingID, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.RawSample

	for rows.Next() {
		var (
			ts  time.Time
			raw []byte
		)

		if err := rows.Scan(&ts, &raw); err != nil {
			return nil, err
		}

		result = append(result, models.RawSample{
			Timestamp: ts.Format(time.RFC3339),
			Raw:       utils.ByteaToIntSlice(raw),
		})
	}

	return result, nil
}

func (r *pgRepository) GetAllRawData(ctx context.Context) ([]dto.TrainingRaw, error) {
	q := `
SELECT 
    id,
    training_id,
    device_id,
    movement_id,
    repetition,
    ts,
    raw
FROM training_raw
ORDER BY id;
`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dto.TrainingRaw

	for rows.Next() {
		var tr dto.TrainingRaw

		if err = rows.Scan(
			&tr.ID,
			&tr.TrainingID,
			&tr.DeviceID,
			&tr.MovementID,
			&tr.Repetition,
			&tr.TS,
			&tr.Raw,
		); err != nil {
			return nil, err
		}

		result = append(result, tr)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
