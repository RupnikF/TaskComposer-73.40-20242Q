CREATE TABLE IF NOT EXISTS services (
    name TEXT PRIMARY KEY,
    broker_url TEXT,
    input_topic TEXT,
    output_topic TEXT,
    last_heartbeat TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tasks (
    service_name TEXT,
    name TEXT,
    UNIQUE (service_name, name),
    CONSTRAINT service_fk FOREIGN KEY (service_name) REFERENCES services
);

CREATE TABLE IF NOT EXISTS task_inputs (
    service_name TEXT,
    name TEXT,
    key TEXT,
    type TEXT,
    CONSTRAINT task_output_fk FOREIGN KEY (service_name, name) REFERENCES tasks
);

CREATE TABLE IF NOT EXISTS task_outputs (
    service_name TEXT,
    name TEXT,
    key TEXT,
    type TEXT,
    CONSTRAINT task_output_fk FOREIGN KEY (service_name, name) REFERENCES tasks
);
