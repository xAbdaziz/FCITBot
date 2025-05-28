# FCITBot

> **Note**: This project was created during my early days at FCIT (2021-2022) and may not follow the cleanest code practices or best software engineering principles. It serves as a functional WhatsApp bot but could benefit from refactoring and improvements.

A WhatsApp bot designed for FCIT (Faculty of Computing and Information Technology) students and groups. The bot provides various academic services, note management, and group administration features.

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

- Go 1.23 or higher
- PostgreSQL database
- WhatsApp account for the bot

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

3. Set up PostgreSQL databases:
```sql
CREATE DATABASE wadb;
CREATE DATABASE groupnotes;
CREATE DATABASE fcitbotmisc;

-- Connect to fcitbotmisc and create tables:
\c fcitbotmisc;
CREATE TABLE allowance (month INT, year INT);
CREATE TABLE vacations (name VARCHAR(255), date TIMESTAMP, duration VARCHAR(255));
```

4. Configure environment:
```bash
cp config.env.example config.env
# Edit config.env with your database URL and WhatsApp numbers
```

5. Run the bot:
```bash
go run main.go
```

### Quick Deploy (Production)

```bash
# Build the binary
go build -o fcitbot main.go

# Run the bot
./fcitbot
```

### Using Docker

```bash
# Build and run
docker build -t fcitbot .
docker run -d --name fcitbot \
  -e DB_URL="postgres://user:pass@host:5432/" \
  -e OWNER_NUMBER="966591234567@s.whatsapp.net" \
  -e BOT_NUMBER="966551234567@s.whatsapp.net" \
  fcitbot
```

## Configuration

The bot requires the following environment variables in [`config.env`](config.env):

- `DB_URL`: PostgreSQL connection string
- `OWNER_NUMBER`: WhatsApp number of the bot owner (with full format)
- `BOT_NUMBER`: WhatsApp number of the bot account (with full format)

## Database Schema

The bot uses three PostgreSQL databases:
- `wadb`: WhatsApp session storage
- `groupnotes`: Group-specific notes storage
- `fcitbotmisc`: Miscellaneous data (allowance tracking, vacations)

## File Structure

```
├── main.go                 # Main application entry point
├── config.env.example      # Environment configuration template
├── cmds.txt               # Available commands list
├── Dockerfile             # Docker configuration
├── docker-compose.yml     # Docker Compose configuration
├── files/                 # PDF documents for academic resources
│   ├── CS_PLAN.pdf
│   ├── IT_PLAN.pdf
│   ├── IS_PLAN.pdf
│   └── ...
└── lib/
    ├── helper/
    │   └── helper.go      # Utility functions
    └── msgHandler/
        └── msghandler.go  # Message processing logic
```

## Usage

1. Add the bot to your WhatsApp group
2. The bot will automatically create a notes table for the group
3. Use `!الأوامر` to see all available commands
4. Admins can use administrative commands like `!اطرد` and `!منشن الكل`

## Development

### Adding New Commands

1. Add the command logic in [`lib/msgHandler/msghandler.go`](lib/msgHandler/msghandler.go)
2. Update the commands list in [`cmds.txt`](cmds.txt)
3. Add any required helper functions in [`lib/helper/helper.go`](lib/helper/helper.go)
