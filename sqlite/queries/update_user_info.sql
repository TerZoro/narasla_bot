INSERT INTO users(owner_id, chat_id, user_name, last_send_at) 
VALUES (?, ?, ?, strftime('%s', 'now'))
ON CONFLICT(owner_id) DO UPDATE SET 
    chat_id = excluded.chat_id,
    user_name = excluded.user_name;