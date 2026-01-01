SELECT owner_id, chat_id, user_name, timezone, send_hour, send_minute, last_send_at
FROM users WHERE enabled = 1;