# Na Rasla Bot

A simple “save now, read later” Telegram bot.

Send a link — it saves it to your personal list. Later you can get a random saved page, list everything, or delete items.

## Features
- Save links (private: just send a link; groups: use `/save@<botname> <url>`)
- `/rnd` — send one random saved page (and remove it from your list)
- `/list` — show saved pages (up to 20)
- `/del` — delete by number or by exact URL
- Uses SQLite for persistent storage

## Commands
- `/help` — show help
- `/save <url>` — save a link (required in groups)
- `/rnd` — send & remove random saved page
- `/list` — show your pages
- `/del` — delete:
  - `/del` (shows list)
  - `/del <number>`
  - `/del <url>`

## Run locally
### 1) Requirements
- Go (1.20+ recommended)
- Telegram bot token (from BotFather)

### 2) Configure
Create `.env`:
```env
TG_BOT_TOKEN=your_token_here
BOT_USERNAME=your_bot_username
STORAGE_PATH=/absolute/path/to/storage.db
```
