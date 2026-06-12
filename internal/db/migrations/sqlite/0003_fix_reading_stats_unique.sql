-- migrate:up
CREATE UNIQUE INDEX IF NOT EXISTS idx_reading_stats_date ON reading_stats(date);
