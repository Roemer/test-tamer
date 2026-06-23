----------
-- Functions
----------

-- Function to automatically set the updated_at column to the current timestamp whenever a row is updated
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

----------
-- Things related to projects
----------

CREATE TABLE projects (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL,
    CONSTRAINT projects_name_uniq UNIQUE (name),
    CONSTRAINT projects_public_id_uniq UNIQUE (public_id)
);

CREATE TRIGGER projects_set_updated_at
BEFORE UPDATE ON projects
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
