SELECT EXISTS (
    SELECT 1 FROM pages WHERE url = ? AND user_name = ?
) AS exists_flag;