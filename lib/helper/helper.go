package helper

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"

	"gorm.io/gorm"
)

type Bot struct {
	Client *whatsmeow.Client
	Msg    *events.Message
	gormDB *gorm.DB
}

func BotContext(client *whatsmeow.Client, message *events.Message, gormDB *gorm.DB) *Bot {
	return &Bot{
		Client: client,
		Msg:    message,
		gormDB: gormDB,
	}
}

func (botContext *Bot) GetCMD() string {
	extended := botContext.Msg.Message.GetExtendedTextMessage().GetText()
	text := botContext.Msg.Message.GetConversation()
	imageMatch := botContext.Msg.Message.GetImageMessage().GetCaption()
	videoMatch := botContext.Msg.Message.GetVideoMessage().GetCaption()
	var command string
	if text != "" {
		command = text
	} else if imageMatch != "" {
		command = imageMatch
	} else if videoMatch != "" {
		command = videoMatch
	} else if extended != "" {
		command = extended
	}
	return command
}

func (botContext *Bot) Reply(msg *waE2E.Message) {
	chatId := botContext.Msg.Info.Chat.ToNonAD()
	msgId := botContext.Msg.Info.ID
	author := botContext.Msg.Info.Sender.ToNonAD()
	_, err := botContext.Client.SendMessage(context.Background(), chatId, msg)
	if err != nil {
		println(err)
		return
	}
	_ = botContext.Client.MarkRead([]types.MessageID{msgId}, time.Now(), chatId, author)
}

func (botContext *Bot) ReplyText(reply string) {
	msgId := botContext.Msg.Info.ID
	author := botContext.Msg.Info.Sender.ToNonAD()
	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(reply),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:      &msgId,
				Participant:   proto.String(author.String()),
				QuotedMessage: botContext.Msg.Message,
			},
		},
	}
	botContext.Reply(msg)
}

func (botContext *Bot) ReplyAndMention(reply string, JIDs []string) {
	quotedMsg := botContext.Msg.Message.ExtendedTextMessage.ContextInfo
	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(reply),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:      quotedMsg.StanzaID,
				Participant:   proto.String(*quotedMsg.Participant),
				QuotedMessage: quotedMsg.QuotedMessage,
				MentionedJID:  JIDs,
			},
		},
	}
	botContext.Reply(msg)
}

func (botContext *Bot) ReplyDocument(file string) {
	msgId := botContext.Msg.Info.ID
	author := botContext.Msg.Info.Sender.ToNonAD()
	content, err := os.ReadFile(file)
	if err != nil {
		println(err)
		return
	}
	resp, err := botContext.Client.Upload(context.Background(), content, whatsmeow.MediaDocument)
	if err != nil {
		println(err)
		return
	}
	msg := &waE2E.Message{
		DocumentMessage: &waE2E.DocumentMessage{
			FileName:      proto.String(filepath.Base(file)),
			Mimetype:      proto.String(http.DetectContentType(content)), // replace this with the actual mime type
			URL:           &resp.URL,
			Title:         proto.String(filepath.Base(file)),
			FileSHA256:    resp.FileSHA256,
			FileLength:    &resp.FileLength,
			MediaKey:      resp.MediaKey,
			FileEncSHA256: resp.FileEncSHA256,
			DirectPath:    &resp.DirectPath,
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:      &msgId,
				Participant:   proto.String(author.String()),
				QuotedMessage: botContext.Msg.Message,
			},
			Caption: proto.String(strings.ReplaceAll(strings.ReplaceAll(filepath.Base(file), "_", " "), ".pdf", "")),
		},
	}
	botContext.Reply(msg)
}

func (botContext *Bot) GetGroupMembers(chat types.JID) []types.GroupParticipant {
	groupInfo, _ := botContext.Client.GetGroupInfo(chat)
	return groupInfo.Participants
}

func (botContext *Bot) GetGroupAdmins(chat types.JID) []types.GroupParticipant {
	members := botContext.GetGroupMembers(chat)
	var admins []types.GroupParticipant
	for _, member := range members {
		if member.IsAdmin || member.IsSuperAdmin {
			admins = append(admins, member)
		}
	}
	return admins
}

func (botContext *Bot) IsUserAdmin(chat types.JID, user string) bool {
	admins := botContext.GetGroupAdmins(chat)
	ownerJID, _ := types.ParseJID(os.Getenv("OWNER_NUMBER"))
	owner, _ := botContext.Client.Store.LIDs.GetLIDForPN(context.Background(), ownerJID)
	for _, admin := range admins {
		if admin.JID.String() == user || owner.String() == user {
			return true
		}
	}
	return false
}

func (botContext *Bot) MemberIsInGroup(chat types.JID, user string) bool {
	groupMembers := botContext.GetGroupMembers(chat)
	for _, member := range groupMembers {
		if member.JID.String() == user {
			return true
		}
	}
	return false
}

func (botContext *Bot) Allowance() {
	saudiTZ := time.FixedZone("UTC+3", 3*60*60)
	now := time.Now().In(saudiTZ)

	allowanceDate := time.Date(now.Year(), now.Month(), 27, 2, 0, 0, 0, saudiTZ)
	allowanceDate = adjustForWeekend(allowanceDate)

	// If we're past this month's allowance date, calculate next month's
	if now.After(allowanceDate) {
		allowanceDate = time.Date(now.Year(), now.Month()+1, 27, 2, 0, 0, 0, saudiTZ)
		allowanceDate = adjustForWeekend(allowanceDate)
	}

	diff := allowanceDate.Sub(now)

	if diff <= 0 {
		botContext.ReplyText("موعد إيداع المكافأة اليوم!")
		return
	}

	days := int(diff.Hours() / 24)
	hours := int(diff.Hours()) % 24
	minutes := int(diff.Minutes()) % 60
	seconds := int(diff.Seconds()) % 60

	daysName := "يوم"
	hoursName := "ساعة"
	minutesName := "دقيقة"
	secondsName := "ثانية"

	if days >= 3 && days <= 10 {
		daysName = "أيام"
	}
	if hours >= 3 && hours <= 10 {
		hoursName = "ساعات"
	}
	if minutes >= 3 && minutes <= 10 {
		minutesName = "دقائق"
	}
	if seconds >= 3 && seconds <= 10 {
		secondsName = "ثواني"
	}

	countdown := fmt.Sprintf("%d %s %d %s %d %s %d %s", days, daysName, hours, hoursName, minutes, minutesName, seconds, secondsName)
	botContext.ReplyText("يتبقى على إيداع المكافأة:\n" + countdown)
}

// adjustForWeekend moves the date to avoid weekends (Friday/Saturday in Saudi Arabia)
func adjustForWeekend(date time.Time) time.Time {
	switch date.Weekday() {
	case time.Friday:
		// Move to Thursday (day before weekend)
		return date.AddDate(0, 0, -1)
	case time.Saturday:
		// Move to Sunday (day after weekend)
		return date.AddDate(0, 0, 1)
	default:
		return date
	}
}
