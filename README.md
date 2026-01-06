# Na Rasla Bot

A simple “save now, read later” Telegram bot.

Send a link — it saves it to your personal list. Later you can get a random saved page, list everything, or delete items.

## Features
- Save links (private: just send a link; groups: use `/save@<botname> <url>`)
- `/rnd` — send one random saved page (and remove it from your list)
- `/list` — show saved pages (up to 20)
- `/del` — delete by number or by exact URL
- `/autopush` — enable/disable daily auto-send
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
- `/autopush` — daily auto-send control:
  - `/autopush on`
  - `/autopush off`
  - `/autopush status`
  - `/autopush` (toggle)

## Auto-send (daily)
- When **autopush is enabled**, the bot sends **one page per day** at `12:00` until `23:00` (Asia/Almaty) and removes it from your list.
- For now, the schedule is global (no per-user time settings yet).
- Current implementation checks users on a scheduler tick (currently **every minute**). This is OK for now; it can be optimized later to sleep until the next planned send time.

## Run locally
### 1) Requirements
- Go (1.20+ recommended)
- Telegram bot token (from BotFather)
- Telegram bot name (from BotFather)
- (Optional) Docker + Docker Compose

### 2) Configure
Create `.env`:
```env
TG_BOT_TOKEN=your_token_here
BOT_USERNAME=your_bot_username
STORAGE_PATH=/absolute/path/to/storage.db
```
### 3) Option A: Run with Go (Binary)
- Build: ```go build -o bin/na_raslabot```
- Run: ```bin/na_raslabot```
- Stop: Press **Ctrl + C**

### 4) Option B: Run with Docker Compose
- Build: ```docker compose up --build -d```
- View logs: ```docker compose logs -f```
- Stop logs View: Press **Ctrl + C**
- Stop Container: ```docker compose down```
- Stop and remove volumes (this will delete SQLite database inside the volume): ```docker compose down -v```