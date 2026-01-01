SELECT timezone, enabled, send_hour, send_minute, last_send_at FROM users
WHERE owner_id = ? LIMIT 1;