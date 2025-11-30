CREATE TABLE movements (
    movement_id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT
);


CREATE TABLE training_sessions (
    id SERIAL PRIMARY KEY,
    movement_id INT REFERENCES movements(movement_id),
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    repetitions INT DEFAULT 5,
    status TEXT DEFAULT 'running'
);

CREATE TABLE training_repetitions (
    id SERIAL PRIMARY KEY,
    session_id INT REFERENCES training_sessions(id),
    repetition_number INT NOT NULL,
    started_at TIMESTAMP NOT NULL,
    ended_at TIMESTAMP,
    samples_collected INT,
    duration_ms INT
);

CREATE TABLE training_raw (
    id BIGSERIAL PRIMARY KEY,
    repetition_id INT REFERENCES training_repetitions(id),
    timestamp BIGINT NOT NULL,
    raw_data INT[] NOT NULL,
    window_size INT NOT NULL
);


CREATE TABLE training_features (
    id BIGSERIAL PRIMARY KEY,
    repetition_id INT REFERENCES training_repetitions(id),
    timestamp BIGINT NOT NULL,
    mav FLOAT,
    rms FLOAT,
    wl FLOAT,
    zc INT,
    ssc INT
);

CREATE TABLE verification_predictions (
    id SERIAL PRIMARY KEY,
    movement_id INT REFERENCES movements(movement_id),
    predicted_movement INT,
    confidence FLOAT,
    timestamp BIGINT NOT NULL
);

CREATE TABLE verification_feedback (
    id SERIAL PRIMARY KEY,
    prediction_id INT REFERENCES verification_predictions(id),
    user_confirmed BOOLEAN NOT NULL,
    feedback_time TIMESTAMP DEFAULT NOW()
);


CREATE TABLE streaming_predictions (
    id BIGSERIAL PRIMARY KEY,
    gesture_id INT REFERENCES movements(movement_id),
    confidence FLOAT,
    timestamp BIGINT NOT NULL
);


CREATE TABLE system_logs (
    id BIGSERIAL PRIMARY KEY,
    level TEXT,
    message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
