package main

import (
	"FCITBot/lib/msgHandler"
	"database/sql"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"go.mau.fi/whatsmeow/types"

	"context"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/protobuf/proto"
	"os"
	"os/signal"
	"syscall"

	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	waCompanionReg "go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func registerHandler(client *whatsmeow.Client, groupNotes *sql.DB, misc *sql.DB) func(evt interface{}) {
	botNum := os.Getenv("BOT_NUMBER")
	myNum, _ := types.ParseJID(os.Getenv("OWNER_NUMBER"))
	return func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			go msgHandler.Handle(v, client, groupNotes, misc)
			break

		case *events.JoinedGroup:
			if len(v.Participants) > 1 {
				if BotIsAdded(v.Participants, botNum) {
					_, _ = groupNotes.Exec(fmt.Sprintf("CREATE TABLE  %s (\"noteName\" TEXT NOT NULL PRIMARY KEY, \"noteContent\" TEXT NOT NULL , \"created_at\" TIMESTAMP NOT NULL DEFAULT NOW())", pgx.Identifier{v.JID.ToNonAD().String()}.Sanitize()))
					_, _ = client.SendMessage(context.Background(), v.JID.ToNonAD(), &waE2E.Message{Conversation: proto.String("شكرًا لإضافتي الى المجموعة.\nللحصول على قائمة الأوامر اكتب: !الأوامر")})
					_, _ = client.SendMessage(context.Background(), myNum, &waE2E.Message{Conversation: proto.String(v.GroupInfo.Name)})
				}
			}
			break

		case *events.GroupInfo:
			if len(v.Leave) == 1 {
				if v.Leave[0].ToNonAD().String() == botNum {
					_, _ = groupNotes.Exec(fmt.Sprintf("DROP TABLE %s", pgx.Identifier{v.JID.ToNonAD().String()}.Sanitize()))
					return
				}
			}
			break
		}
	}
}

func main() {
	// Load config.env
	_ = godotenv.Load("config.env")

	// Spoof the bot as Windows
	store.SetOSInfo("Windows", store.GetWAVersion())
	store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_DESKTOP.Enum()

	dbLog := waLog.Stdout("Database", "ERROR", true)
	container, err := sqlstore.New(context.Background(), "pgx", os.Getenv("DB_URL")+"wadb", dbLog)
	if err != nil {
		panic(err)
	}
	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// PostgreSQL stuff
	groupNotes, _ := sql.Open("pgx", os.Getenv("DB_URL")+"groupnotes")
	misc, _ := sql.Open("pgx", os.Getenv("DB_URL")+"fcitbotmisc")

	eventHandler := registerHandler(client, groupNotes, misc)
	client.AddEventHandler(eventHandler)

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}

func BotIsAdded(participants []types.GroupParticipant, botNum string) bool {
	for _, participant := range participants {
		if participant.JID.ToNonAD().String() == botNum {
			return true
		}
	}
	return false
}
