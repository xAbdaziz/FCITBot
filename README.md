# FCITBot

## Features

### 📚 Academic Services
- **Study Plans**: Get study plans for CS, IT, and IS majors
- **Academic Calendar**: Access current academic calendar
- **Transfer Conditions**: Information about transferring to FCIT
- **Major Differences**: Comparison between different majors
- **Tracks**: Available tracks in the faculty
- **Elective Courses**: List of elective courses for each major
- **Classrooms**: Links to classroom schedules
- **Allowance Tracker**: Check remaining time until next allowance
- **Schedule Tools**: Links to BetterKAU and FCIT Groups websites

### 💬 Group Management
- **Kick Members**: Remove members from groups (admin only)
- **Mention All**: Mention all group members (admin only)
- **Report Messages**: Report inappropriate messages to admins
- **Notes System**: Save, retrieve, and delete group-specific notes

### 🤖 Bot Commands
All commands start with `!` (exclamation mark):

- `!الأوامر` - Show all available commands
- `!خطة [CS/IT/IS]` - Get study plan for specified major
- `!التقويم الأكاديمي` - Get academic calendar
- `!شروط التحويل` - Get transfer conditions
- `!الفرق بين التخصصات` - Get differences between majors
- `!المسارات` - Get faculty tracks
- `!المواد الاختيارية` - Get elective courses
- `!المكافأة` - Check time remaining until next allowance
- `!القاعات` - Get classroom links
- `!الجدول` - Get BetterKAU link
- `!القروبات` - Get FCIT Groups link
- `!احفظ [name]` - Save a note (admin only)
- `!هات [name]` - Retrieve a note
- `!احذف [name]` - Delete a note (admin only)
- `!الملاحظات` - List all saved notes
- `!اطرد` - Kick a member (admin only)
- `!منشن الكل` - Mention all members (admin only)
- `!تبليغ` - Report a message to admins
- `!اقتراحات` - Contact developer

## Prerequisites

- Go 1.25 or higher
- A WhatsApp account for the bot
- *(Optional)* PostgreSQL database — SQLite is used by default

## Installation

### Quick Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd FCITBot
```

2. Install dependencies:
```bash
go mod download
```

3. Configure environment:
```bash
cp config.env.example config.env
# Edit config.env with your WhatsApp owner number
# Optionally set DB_URL to use PostgreSQL instead of SQLite
```

4. Run the bot:
```bash
go run main.go
```

On first run, a QR code will be printed to the terminal — scan it with WhatsApp to log in. Session data is persisted locally so subsequent runs connect automatically.

### Quick Deploy (Production)

```bash
# Build the binary
go build -o fcitbot main.go

# Run the bot
./fcitbot
```

### Using Docker Compose (Recommended)

The easiest way to run the bot in production. On first run, enable `tty` and `stdin_open` in `docker-compose.yml` to scan the QR code, then disable them again.

```bash
# Pull and start
docker compose up -d
```

The `docker-compose.yml` uses the pre-built image from `ghcr.io/xabdaziz/fcitbot:latest` and mounts a local `./fcitbot-data` volume for persistent session and database storage.

### Using Docker (Manual)

```bash
# Build and run
docker build -t fcitbot .
docker run -d --name fcitbot \
  -v ./fcitbot-data:/app/data \
  -e OWNER_NUMBER="966591234567@s.whatsapp.net" \
  fcitbot
```

## Configuration

Copy [`config.env.example`](config.env.example) to `config.env` and set the following:

| Variable | Required | Description |
|---|---|---|
| `OWNER_NUMBER` | ✅ Yes | WhatsApp JID of the bot owner (e.g. `966591234567@s.whatsapp.net`) |
| `DB_URL` | ❌ Optional | PostgreSQL connection string. If omitted, SQLite is used. |

## Database

The bot supports two database backends, selected automatically based on whether `DB_URL` is set:

- **SQLite** *(default)*: No setup needed. Data is stored in `data/whatsmeow.db` (WhatsApp session) and `data/fcitbot.db` (bot data).
- **PostgreSQL**: Set `DB_URL` to a valid connection string (e.g. `postgres://user:pass@host:5432/dbname`).

Database schema is managed automatically via GORM's `AutoMigrate` — no manual SQL setup required.

## File Structure

```
├── main.go                    # Application entry point; connects to WhatsApp and dispatches events
├── config.env.example         # Environment configuration template
├── Dockerfile                 # Docker build configuration
├── docker-compose.yml         # Docker Compose configuration (uses pre-built image)
├── go.mod / go.sum            # Go module files
├── models/
│   └── groupsnotes.go         # GORM model for group notes
├── files/                     # PDF documents served by the bot
│   ├── CALENDAR.pdf
│   ├── CS_PLAN.pdf
│   ├── IT_PLAN.pdf
│   ├── IS_PLAN.pdf
│   ├── DIFFERENCE_BETWEEN_MAJORS.pdf
│   ├── ELECTIVE_COURSES.pdf
│   ├── FCIT_TRACKS.pdf
│   └── TRANSFERRING_CONDITIONS.pdf
└── lib/
    ├── helper/
    │   └── helper.go          # Utility / helper functions (Bot struct, send helpers)
    └── msgHandler/
        ├── msghandler.go      # Core handler: command registry, routing, MessageContext
        ├── cmd_allowance.go
        ├── cmd_broadcast.go
        ├── cmd_calendar.go
        ├── cmd_elective_courses.go
        ├── cmd_groups.go
        ├── cmd_kick.go
        ├── cmd_major_differences.go
        ├── cmd_mention_all.go
        ├── cmd_notes.go
        ├── cmd_plan.go
        ├── cmd_report.go
        ├── cmd_schedule.go
        ├── cmd_tracks.go
        └── cmd_transfer.go
```

## Usage

1. Add the bot to your WhatsApp group
2. The bot will automatically set up its database tables on startup
3. Use `!الأوامر` to see all available commands
4. Admins can use administrative commands like `!اطرد` and `!منشن الكل`

## Development

### Adding New Commands

Commands are registered using the `RegisterCommand` function in [`lib/msgHandler/msghandler.go`](lib/msgHandler/msghandler.go). Each command is a self-contained `cmd_*.go` file.

1. Create a new file `lib/msgHandler/cmd_mycommand.go`
2. Register your command in an `init()` function:

```go
package msgHandler

func init() {
    RegisterCommand(Command{
        Name:        "!mycommand",
        Description: "Description of my command",
        Handler:     handleMyCommand,
    })
}

func handleMyCommand(mc *MessageContext) {
    mc.HelperLib.ReplyText("Hello!")
}
```

3. Add any required PDF files to the [`files/`](files/) directory
4. Add any required helper functions to [`lib/helper/helper.go`](lib/helper/helper.go)
