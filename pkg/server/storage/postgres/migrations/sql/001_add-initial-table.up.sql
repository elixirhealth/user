
CREATE SCHEMA "user";

CREATE TABLE "user".entity (
  row_id SERIAL PRIMARY KEY,
  transaction_period TSTZRANGE NOT NULL DEFAULT tstzrange(NOW(), 'infinity', '[)'),
  user_id VARCHAR NOT NULL,
  entity_id VARCHAR NOT NULL
);

CREATE UNIQUE INDEX entity_user ON "user".entity (entity_id, user_id);