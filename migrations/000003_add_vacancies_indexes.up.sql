CREATE INDEX IF NOT EXISTS vacancies_title_idx ON vacancies USING GIN (to_tsvector('simple', title));
CREATE INDEX IF NOT EXISTS vacancies_tags_idx ON vacancies USING GIN (tags)