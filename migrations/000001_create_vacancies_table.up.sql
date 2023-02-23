CREATE TABLE IF NOT EXISTS vacancies (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    title text NOT NULL,
    company text NOT NULL,
	active boolean NOT NULL DEFAULT true,
    tags text[] NOT NULL,
    version integer NOT NULL DEFAULT 1
);

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";