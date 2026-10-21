-- The asynchronous Pub/Sub login flow was replaced by a direct call from
-- profiles.Login to auth.Login, so the correlation table is no longer used.
DROP TABLE IF EXISTS login_results;
