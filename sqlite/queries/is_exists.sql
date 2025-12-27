SELECT EXISTS (
    SELECT 1 FROM pages WHERE owner_id = ? AND url = ?
) AS exists_flag;