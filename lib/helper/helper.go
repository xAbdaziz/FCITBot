package helper

import (
	"context"
	"database/sql"
	"fmt"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Bot struct {
	Client *whatsmeow.Client
	Msg    *events.Message
	Misc   *sql.DB
}

func BotContext(client *whatsmeow.Client, message *events.Message, misc *sql.DB) *Bot {
	return &Bot{
		Client: client,
		Msg:    message,
		Misc:   misc,
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
	myNum := os.Getenv("OWNER_NUMBER")
	for _, admin := range admins {
		if admin.JID.String() == user || myNum == user {
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

func diffBetweenDates(date string) (counter string) {
	dateNowUnix := time.Now().UnixMilli()
	dateThen, _ := time.Parse(time.RFC3339, date)
	dateThenUnix := dateThen.UnixMilli()
	if (dateThenUnix - dateNowUnix) >= 0 {
		delta := math.Abs(float64(dateThenUnix-dateNowUnix)) / 1000

		days := math.Floor(delta / 86400)
		delta -= days * 86400

		hours := math.Mod(math.Floor(delta/3600), 24)
		delta -= hours * 3600

		minutes := math.Mod(math.Floor(delta/60), 60)
		delta -= minutes * 60

		seconds := math.Floor(math.Mod(delta, 60))

		daysName := "يوم"
		hoursName := "ساعة"
		minutesName := "دقيقة"
		secondsName := "ثانية"

		if days < 1 {
			return "no allowance"
		}

		if days <= 10 {
			daysName = "أيام"
		}
		if hours <= 10 {
			hoursName = "ساعات"
		}
		if minutes <= 10 {
			minutesName = "دقائق"
		}
		if seconds <= 10 {
			secondsName = "ثواني"
		}

		counter := fmt.Sprintf("%.0f %s %.0f %s %.0f %s %.0f %s", days, daysName, hours, hoursName, minutes, minutesName, seconds, secondsName)

		return counter
	} else {
		return ""
	}
}

func (botContext *Bot) Vacation() {
	var name string
	var date string
	var duration string
	var createdAt string
	err := botContext.Misc.QueryRow("SELECT * from vacations ORDER BY created_at ASC").Scan(&name, &date, &duration, &createdAt)
	if err != nil {
		panic(err)
		return
	}
	diff := diffBetweenDates(date)
	if diff != "" {
		botContext.ReplyText("اقرب إجازة هي: إجازة " + name + "\n" + "تبدأ بعد: " + diff + "\n" + "مدتها: " + duration)
	} else {
		_, err := botContext.Misc.Exec("DELETE FROM vacations WHERE date=$1", date)
		if err != nil {
			panic(err)
			return
		}
		botContext.Vacation()
	}
}

func (botContext *Bot) Allowance() {
	day := 27
	var month int
	var year int
	_ = botContext.Misc.QueryRow("SELECT * from allowance").Scan(&month, &year)
	date := fmt.Sprintf("%d-%02d-%dT02:00:00.000+03:00", year, month, day)
	dateDay, _ := time.Parse(time.RFC3339, date)
	if dateDay.Weekday() == 6 {
		day++
	} else if dateDay.Weekday() == 5 {
		day--
	}
	date = fmt.Sprintf("%d-%02d-%dT02:00:00.000+03:00", year, month, day)
	diff := diffBetweenDates(date)
	if diff == "no allowance" {
		replies := [17]string{"حرك يا فقير", "قطعناها عنك، روح دور لك على شغلة", "معدلك تعبان ما فيه فلوس", "شفلك حياة", "broke guy", "القم يا فقير", "Go work at McDonald's, broke guy", "If poverty gave out degrees, you’d have a PhD", "عطنا رقم بابا عشان نعطيك مصروف", "McDonald’s is hiring bro", "حمل إحسان وشوف قسم التبرعات", "غداً تُرزقون يا معشر المحتاجين", "اليوم تشحت، بكرة تصرف", "You’re 24 hours away from being slightly less pathetic", "Put the fries in the bag lil bro", "شكله خلص حبر الطابعة وما عندك حقه", "Bro so broke he can't afford calculator batteries 💀💀"}
		rand.Seed(time.Now().UnixNano())
		reply := replies[rand.Intn(len(replies))]
		botContext.ReplyText(reply)
		return
	}
	if diff != "" {
		botContext.ReplyText("يتبقى على إيداع المكافأة:\n" + diff)
	} else {
		if month == 12 {
			month = 1
			year++
		} else {
			month++
		}
		botContext.Misc.QueryRow("UPDATE allowance SET year=$1, month=$2;", year, month)
		botContext.Allowance()
	}
}
