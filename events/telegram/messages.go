package telegram

const msgHelp = `I'm a simple “save now, read later” bot.

How to save:
• In private chat: just send me a link — I'll save it.
• In group chats: use /save@na_raslabot <link> (so I don't react to random messages).

Commands:
• /help — show this message
• /save <url> — save a link (required in groups)
• /rnd — send one random saved page and remove it from your list
• /del — delete a page:
  - /del            (show your list)
  - /del <number>   (delete by number from the list)
  - /del <url>      (delete by exact link)
• /list — show your saved pages (up to 20)

Note:
After /rnd, the sent page is deleted from your list (so you won't get repeats).`

// TODO: add method for changing saving logic -- deleting and not deleting.
// TODO: add method for changing language ru/eng.

const msgHello = "Hellooo! :3\n\n" + msgHelp

const (
	msgUnknownCommand     = "Unknown command."
	msgNoSavedPages       = "You have no saved pages."
	msgSaved              = "Saved!"
	msgAlreadyExists      = "You already have this page on your list."
	msgDeleted            = "Page was deleted."
	msgIncorrectDeleteArg = "Usage: /del or /del <number> or /del <url>"
	msgIncorrectSave      = "Usage: /save <url>"
)
