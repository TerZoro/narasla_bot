package telegram

const msgHelp = `I'm a simple “save now, read later” bot.

Send me any link — I'll save it to your personal list.

Commands:
• /help — show this message
• /rnd  — send one random saved page (and remove it from your list)
• /del  — delete a page: /del <number> or /del <url> or just /del to show list
• /list — show your saved pages (up to 20)

Note:
After /rnd, the sent page is deleted from your list (so you won't get repeats).`

// TODO: add method for changing saving logic -- deleting and not deleting.
// TODO: add method for changing language ru/eng.

const msgHello = "Hellooo! :3\n\n" + msgHelp

const (
	msgUnknownCommand     = "Unknown command :/"
	msgNoSavedPages       = "You have no saved pages :("
	msgSaved              = "Saved! ;))"
	msgAlreadyExists      = "You already have this page on your list <3"
	msgDeleted            = "Page was deleted 0_0"
	msgIncorrectDeleteArg = "Usage: /del or /del <number> or /del <url>"
)
