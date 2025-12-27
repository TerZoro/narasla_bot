package telegram

import (
	"context"
	"strings"
)

func (p *Processor) initHandlers() {
	p.handlers = map[string]handler{
		SaveCmd:   p.hSave,
		RndCmd:    p.hRand,
		HelpCmd:   p.hHelp,
		StartCmd:  p.hStart,
		DeleteCmd: p.hDel,
		ListCmd:   p.hList,
	}
}

func (p *Processor) middleHandler(ctx context.Context, cmd, arg string, m Meta) error {
	h, ok := p.handlers[cmd]
	if !ok {
		return p.tg.SendMessage(ctx, m.Chat.ID, m.UserID, msgUnknownCommand)
	}

	return h(ctx, arg, m)
}

func (p *Processor) hSave(ctx context.Context, arg string, m Meta) error {
	arg = strings.TrimSpace(arg)
	if arg == "" || !isAddCmd(arg) {
		return p.tg.SendMessage(ctx, m.Chat.ID, m.UserID, msgIncorrectSave)
	}

	return p.savePage(ctx, m.Chat.ID, m.UserID, arg, m.Username)
}

func (p *Processor) hRand(ctx context.Context, arg string, m Meta) error {
	return p.sendRandom(ctx, m.Chat.ID, m.UserID)
}

func (p *Processor) hHelp(ctx context.Context, arg string, m Meta) error {
	return p.tg.SendMessage(ctx, m.Chat.ID, m.UserID, msgHelp)
}

func (p *Processor) hStart(ctx context.Context, arg string, m Meta) error {
	return p.tg.SendMessage(ctx, m.Chat.ID, m.UserID, msgHello)
}

func (p *Processor) hDel(ctx context.Context, arg string, m Meta) error {
	arg = strings.TrimSpace(arg)
	return p.removePage(ctx, m.Chat.ID, m.UserID, m.Username, arg)
}

func (p *Processor) hList(ctx context.Context, arg string, m Meta) error {
	return p.sendList(ctx, m.Chat.ID, m.UserID, m.Username)
}
