/* pgmigrate-encoding: utf-8 */

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE token
(
    id         SERIAL PRIMARY KEY,
    user_id    UUID                     NOT NULL,
    hash       TEXT                     NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(user_id)
);
