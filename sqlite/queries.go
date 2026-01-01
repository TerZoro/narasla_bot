package sqlite

import (
	"embed"
	"log"
)

//go:embed queries/*.sql
var queriesFS embed.FS

func mustSQL(name string) string {
	sb, err := queriesFS.ReadFile("queries/" + name)
	if err != nil {
		log.Fatalf("missing sql file: %s: %v", name, err)
	}

	return string(sb)
}

var (
	qSave        = mustSQL("save.sql")
	qPickRandom  = mustSQL("pick_random.sql")
	qRemove      = mustSQL("remove.sql")
	qIsExists    = mustSQL("is_exists.sql")
	qRemoveByUrl = mustSQL("remove_by_url.sql")
	qList        = mustSQL("list.sql")
	qCount       = mustSQL("count.sql")

	qListEnabledUsers = mustSQL("list_enabled_users.sql")
	qUpdateLastSendAt = mustSQL("update_last_send_at.sql")
	qUpdateUserInfo   = mustSQL("update_user_info.sql")
	qUpdateEnabled    = mustSQL("update_enabled.sql")
	qGetUserInfo      = mustSQL("get_user_info.sql")

	qInit = mustSQL("init.sql")
)
