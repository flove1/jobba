CREATE TABLE IF NOT EXISTS subscribers (
    id bigserial PRIMARY KEY,
	userId bigserial references users(id),
	tag text,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);