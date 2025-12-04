CREATE TABLE devices (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE,
    status TEXT CHECK (status IN ('idle', 'streaming', 'disconnected', 'reserved')),
    last_seen TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE movements (
    movement_id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT
);

CREATE TABLE training (
    id SERIAL PRIMARY KEY,
    device_id INT NOT NULL REFERENCES devices(id),
    movement_id INT NOT NULL REFERENCES movements(movement_id),
    repetition INT NOT NULL,
    finished BOOLEAN NOT NULL DEFAULT false,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE training_raw (
    id BIGSERIAL PRIMARY KEY,
    training_id INTEGER NOT NULL id,
    device_id INTEGER NOT NULL REFERENCES devices(id),
    movement_id INTEGER NOT NULL REFERENCES movements(movement_id),
    repetition INTEGER NOT NULL,
    ts TIMESTAMPTZ NOT NULL,
    raw BYTEA NOT NULL
);


INSERT INTO movements (movement_id, name, description) VALUES
                                                           (1, 'Rest', 'Open/relaxed hand, no intentional tension'),
                                                           (2, 'Fist', 'Strong full hand grip'),
                                                           (3, 'Wrist Flexion', 'Bending wrist down with moderate tension');