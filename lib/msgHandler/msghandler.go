package msgHandler

import (
	"FCITBot/lib/helper"

	"context"
	"os"
	"strings"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	"gorm.io/gorm"
)

type MessageContext struct {
	Message         *events.Message
	Client          *whatsmeow.Client
	GormDB          *gorm.DB
	HelperLib       *helper.Bot
	Chat            types.JID
	Author          string
	BotNum          string
	QuotedMsg       *waE2E.Message
	QuotedMsgText   string
	QuotedMsgAuthor string
	Owner           types.JID
	MsgContentSplit []string
	Ctx             context.Context
}

type Command struct {
	Name        string
	Description string
	Handler     func(*MessageContext)
}

var commands []Command
var commandMap map[string]func(*MessageContext)

func RegisterCommand(cmd Command) {
	commands = append(commands, cmd)
}

func init() {
	// Build commandMap after all init() functions have registered their commands.
	// We use a deferred build approach — the map is built on first Handle() call.
}

func buildCommandMap() {
	commandMap = make(map[string]func(*MessageContext))
	for _, cmd := range commands {
		commandMap[cmd.Name] = cmd.Handler
	}
}

func Handle(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB) {
	if commandMap == nil {
		buildCommandMap()
	}

	ctx := context.Background()
	helperLib := helper.BotContext(client, message, gormDB)
	if message.Info.IsFromMe {
		return
	}
	if !message.Info.IsGroup {
		helperLib.ReplyText("يا هلا، معاك بوت الحاسبات\nرجاءً تحدث معي في القروب فقط")
		return
	}

	ownerJID, _ := types.ParseJID(os.Getenv("OWNER_NUMBER"))
	owner, err := client.Store.LIDs.GetLIDForPN(ctx, ownerJID)
	if err != nil {
		println("Error getting owner number, some commands might not work as expected", err)
		return
	}
	botNum := client.Store.GetLID().ToNonAD().String()

	msgContent := helperLib.GetCMD()
	msgContentSplit := strings.Split(msgContent, " ")
	quotedMsgContext := message.Message.ExtendedTextMessage.GetContextInfo()
	quotedMsg := quotedMsgContext.GetQuotedMessage()
	quotedMsgText := quotedMsg.GetConversation()
	quotedMsgAuthor := quotedMsgContext.GetParticipant()
	chat := message.Info.Chat.ToNonAD()
	author := message.Info.Sender.ToNonAD().String()

	// Create message context
	mc := &MessageContext{
		Message:         message,
		Client:          client,
		GormDB:          gormDB,
		HelperLib:       helperLib,
		Chat:            chat,
		Author:          author,
		BotNum:          botNum,
		QuotedMsg:       quotedMsg,
		QuotedMsgText:   quotedMsgText,
		QuotedMsgAuthor: quotedMsgAuthor,
		Owner:           owner.ToNonAD(),
		MsgContentSplit: msgContentSplit,
		Ctx:             ctx,
	}

	// Try exact command match first
	if handler, exists := commandMap[msgContent]; exists {
		handler(mc)
		return
	}

	// Try prefix match for commands with arguments
	for cmd, handler := range commandMap {
		if strings.HasPrefix(msgContent, cmd+" ") {
			handler(mc)
			return
		}
	}
}
