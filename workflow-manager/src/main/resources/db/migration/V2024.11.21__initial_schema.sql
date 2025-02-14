CREATE TABLE IF NOT EXISTS workflows (
    id SERIAL PRIMARY KEY,
    workflow_name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS args (
    id SERIAL PRIMARY KEY,
    workflow_id INT NOT NULL,
    arg_key TEXT NOT NULL,
    default_value TEXT,
    UNIQUE (workflow_id, arg_key),
    CONSTRAINT fk_args_workflows FOREIGN KEY (workflow_id) REFERENCES workflows(id)
);

CREATE TABLE IF NOT EXISTS steps (
    id SERIAL PRIMARY KEY,
    workflow_id INT NOT NULL,
    step_name TEXT NOT NULL,
    step_order INT NOT NULL,
    service TEXT,
    task TEXT,
    CONSTRAINT fk_steps_workflows FOREIGN KEY (workflow_id) REFERENCES workflows(id)
);

CREATE TABLE IF NOT EXISTS step_inputs (
    id SERIAL PRIMARY KEY,
    step_id INT NOT NULL,
    input_key TEXT NOT NULL,
    input_value TEXT NOT NULL,
    UNIQUE (step_id, input_key),
    CONSTRAINT fk_step_inputs_steps FOREIGN KEY (step_id) REFERENCES steps(id)
);