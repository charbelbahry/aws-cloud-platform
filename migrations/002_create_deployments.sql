CREATE TABLE IF NOT EXISTS deployments (
  id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  service_id  UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  version TEXT NOT NULL,
  status  TEXT NOT NULL DEFAULT 'pending'
          CHECK (status IN (
              'pending', 'building', 'running', 'failed', 'rolled_back'
          )),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_deployments_service_id
    ON deployments(service_id);
