package telegram

const msgHelp = `I'm a telegram bot offering you to save and keep your pages, that you could read/watch later. 
Then, sometimes I will send a page from your list to remind you.

In order to save the page, just send me a link to it.

Use command /rnd :-: to get a random page from your list

Caution! After I send you a page, the page will be removed from the list!`

// TODO: add method for changing saving logic -- deleting and not deleting.
// TODO: add method for changing language ru/eng.

const msgHello = "Hellooo! :3 \n\n" + msgHelp

const (
	msgUnknownCommand = "Unknown command :/"
	msgNoSavedPages   = "You have no saved pages :("
	msgSaved          = "Saved! ;))"
	msgAlreadyExists  = "You already have this page on your list <3"
)
