package main

import (
	"FCITBot/lib/msgHandler"
	"FCITBot/models"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/mdp/qrterminal/v3"
	"google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow"
	waCompanionReg "go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func registerHandler(client *whatsmeow.Client, gormDB *gorm.DB) func(evt interface{}) {
	botNum := os.Getenv("BOT_NUMBER")
	myNum, _ := types.ParseJID(os.Getenv("OWNER_NUMBER"))
	return func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			go msgHandler.Handle(v, client, gormDB)
			break

		case *events.JoinedGroup:
			if len(v.Participants) > 1 {
				if BotIsAdded(v.Participants, botNum) {
					client.SendMessage(context.Background(), v.JID.ToNonAD(), &waE2E.Message{Conversation: proto.String("شكرًا لإضافتي الى المجموعة.\nللحصول على قائمة الأوامر اكتب: !الأوامر")})
					client.SendMessage(context.Background(), myNum, &waE2E.Message{Conversation: proto.String(v.GroupInfo.Name)})
				}
			}
			break

		case *events.GroupInfo:
			if len(v.Leave) == 1 {
				if v.Leave[0].ToNonAD().String() == botNum {
					gormDB.Delete(&models.GroupsNotes{}, "group_id = ?", v.JID.ToNonAD().String())
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
	ctx := context.Background()
	container, err := sqlstore.New(ctx, "sqlite3", "file:whatsmeow.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// GORM
	gormDB, err := gorm.Open(sqlite.Open("file:fcitbot.db?_foreign_keys=on&journal_mode=WAL"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// Migrate models
	gormDB.AutoMigrate(&models.GroupsNotes{}, &models.Allowance{})

	eventHandler := registerHandler(client, gormDB)
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
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
	db, _ := gormDB.DB()
	db.Close()
}

func BotIsAdded(participants []types.GroupParticipant, botNum string) bool {
	for _, participant := range participants {
		if participant.PhoneNumber.ToNonAD().String() == botNum {
			return true
		}
	}
	return false
}
