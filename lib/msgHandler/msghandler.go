package msgHandler

import (
	"FCITBot/lib/helper"
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
	"os"
	"strings"
)

const cmdOpe = "!"

var cmdsFile, _ = os.ReadFile("cmds.txt")
var cmds = string(cmdsFile)

func Handle(message *events.Message, client *whatsmeow.Client, groupNotes *sql.DB, misc *sql.DB) {
	helperLib := helper.BotContext(client, message, misc)
	if message.Info.IsFromMe {
		return
	}
	if !message.Info.IsGroup {
		helperLib.ReplyText("يا هلا، معاك بوت الحاسبات\nرجاءً تحدث معي في القروب فقط")
		return
	}

	myNum := os.Getenv("OWNER_NUMBER")
	botNum := os.Getenv("BOT_NUMBER")

	msgContent := helperLib.GetCMD()
	msgContentSplit := strings.Split(msgContent, " ")
	quotedMsgContext := message.Message.ExtendedTextMessage.GetContextInfo()
	quotedMsg := quotedMsgContext.GetQuotedMessage()
	quotedMsgText := quotedMsg.GetConversation()
	quotedMsgAuthor := quotedMsgContext.GetParticipant()
	chat := message.Info.Chat.ToNonAD()
	author := message.Info.Sender.ToNonAD().String()

	if msgContent == cmdOpe+"الأوامر" {
		helperLib.ReplyText(cmds)
		return

	} else if msgContentSplit[0] == cmdOpe+"اطرد" {
		if !helperLib.IsUserAdmin(chat, author) {
			helperLib.ReplyText("حرك حرك تراك مو ادمن")
			return
		}
		if !helperLib.IsUserAdmin(chat, botNum) {
			helperLib.ReplyText("انا مو ادمن")
			return
		}
		if quotedMsg == nil {
			helperLib.ReplyText("الرجاء تحديد العضو المراد طرده بالرد على رسالته")
			return
		}
		if quotedMsgAuthor == botNum {
			helperLib.ReplyText("يواد قم بس، ما رح اطرد نفسي")
			return
		}
		if quotedMsgAuthor == myNum {
			helperLib.ReplyText("يواد قم بس، ما رح اطرد مطوري")
			return
		}
		if !helperLib.MemberIsInGroup(chat, quotedMsgAuthor) {
			helperLib.ReplyText("العضو غير موجود بالمجموعة")
			return
		}
		if helperLib.IsUserAdmin(chat, quotedMsgAuthor) {
			helperLib.ReplyText("مقدر اطرد ادمن")
			return
		}
		usertoKick, _ := types.ParseJID(quotedMsgAuthor)
		_, _ = client.UpdateGroupParticipants(chat, []types.JID{usertoKick}, whatsmeow.ParticipantChangeRemove)
		revokeMessage := client.BuildRevoke(chat, usertoKick, quotedMsgContext.GetStanzaId())
		_, _ = client.SendMessage(context.Background(), chat, revokeMessage)
		helperLib.ReplyText("تم طرد " + strings.ReplaceAll(quotedMsgAuthor, "@s.whatsapp.net", ""))
		return

	} else if msgContentSplit[0] == cmdOpe+"احفظ" {
		if !helperLib.IsUserAdmin(chat, author) {
			helperLib.ReplyText("حرك حرك تراك مو ادمن")
			return
		}
		if len(msgContentSplit) != 2 || quotedMsg == nil {
			helperLib.ReplyText("استخدام خاطئ\nقم بالرد على الرسالة المراد حفظها ثم كتابة احفظ مع اسم الملاحظة بدون مسافة\n\nمثال: !احفظ اسم_الملاحظة")
			return
		}
		noteName := msgContentSplit[1]

		// Check if quotedMsg is not nil and either Conversation or ExtendedTextMessage.Text is not nil
		if quotedMsg != nil && (quotedMsg.Conversation != nil || (quotedMsg.GetExtendedTextMessage() != nil && quotedMsg.GetExtendedTextMessage().Text != nil)) {
			// Extract the text from the quoted message
			if extendedTextMsg := quotedMsg.GetExtendedTextMessage(); extendedTextMsg != nil && extendedTextMsg.Text != nil {
				quotedMsgText = extendedTextMsg.GetText()
			} else {
				quotedMsgText = *quotedMsg.Conversation
			}
			_, err := groupNotes.Exec(fmt.Sprintf("INSERT INTO %s (\"noteName\", \"noteContent\") VALUES ($1, $2) ON CONFLICT (\"noteName\") DO UPDATE SET \"noteContent\" = excluded.\"noteContent\";", pgx.Identifier{chat.String()}.Sanitize()), noteName, quotedMsgText)
			if err != nil {
				println(err)
				return
			}
			helperLib.ReplyText("تم حفظ الملاحظة " + "\"" + noteName + "\"")
			return
		} else {
			helperLib.ReplyText("مقدر احفظ غير النصوص حالياً")
			return
		}
	} else if msgContentSplit[0] == cmdOpe+"هات" {
		if len(msgContentSplit) != 2 {
			helperLib.ReplyText("استخدام خاطئ\nاكتب هات مع اسم الملاحظة بدون مسافة\n\nمثال: !هات اسم_الملاحظة ")
			return
		}
		noteName := msgContentSplit[1]
		var noteExists bool
		_ = groupNotes.QueryRow(fmt.Sprintf("SELECT EXISTS(SELECT * FROM %s WHERE \"noteName\"=$1);", pgx.Identifier{chat.String()}.Sanitize()), noteName).Scan(&noteExists)
		if !noteExists {
			helperLib.ReplyText("لا توجد ملاحظة بأسم " + "\"" + noteName + "\"")
			return
		}
		var noteContent string
		_ = groupNotes.QueryRow(fmt.Sprintf("SELECT \"noteContent\" FROM %s WHERE \"noteName\"=$1;", pgx.Identifier{chat.String()}.Sanitize()), noteName).Scan(&noteContent)
		helperLib.ReplyText(noteContent)
		return

	} else if msgContentSplit[0] == cmdOpe+"احذف" {
		if !helperLib.IsUserAdmin(chat, author) {
			helperLib.ReplyText("حرك حرك تراك مو ادمن")
			return
		}
		if len(msgContentSplit) != 2 {
			helperLib.ReplyText("استخدام خاطئ\nاكتب احذف مع اسم الملاحظة بدون مسافة\n\nمثال: !احذف اسم_الملاحظة ")
			return
		}
		noteName := msgContentSplit[1]
		var noteExists bool
		_ = groupNotes.QueryRow(fmt.Sprintf("SELECT EXISTS(SELECT * FROM %s where \"noteName\"=$1);", pgx.Identifier{chat.String()}.Sanitize()), noteName).Scan(&noteExists)
		if !noteExists {
			helperLib.ReplyText("لا توجد ملاحظة بأسم " + "\"" + noteName + "\"")
			return
		}
		_, _ = groupNotes.Exec(fmt.Sprintf("DELETE FROM %s where \"noteName\"=$1", pgx.Identifier{chat.String()}.Sanitize()), noteName)
		helperLib.ReplyText("تم حذف الملاحظة " + "\"" + noteName + "\"")
		return
	} else if msgContent == cmdOpe+"الملاحظات" {
		var notesExists bool
		_ = groupNotes.QueryRow(fmt.Sprintf("SELECT EXISTS(SELECT * FROM %s);", pgx.Identifier{chat.String()}.Sanitize())).Scan(&notesExists)
		if notesExists != true {
			helperLib.ReplyText("لا توجد ملاحظات محفوظة.")
			return
		}
		notes := "الملاحظات المحفوظة:"
		listOfNotes, _ := groupNotes.Query(fmt.Sprintf("SELECT \"noteName\" from %s;", pgx.Identifier{chat.String()}.Sanitize()))
		for listOfNotes.Next() {
			var NoteName string
			_ = listOfNotes.Scan(&NoteName)
			notes += "\n- " + NoteName
		}
		helperLib.ReplyText(notes)
		return

	} else if msgContent == cmdOpe+"تبليغ" {
		if quotedMsg == nil {
			helperLib.ReplyText("الرجاء استخدام الأمر على الرسالة المراد التبليغ عنها")
			return
		}
		adminsNum := ""
		var adminsJID []string
		admins := helperLib.GetGroupAdmins(chat)
		for _, admin := range admins {
			if admin.JID.String() != botNum {
				adminsNum += "@" + strings.ReplaceAll(admin.JID.ToNonAD().String(), "@s.whatsapp.net", "") + "\n"
				adminsJID = append(adminsJID, admin.JID.ToNonAD().String())
			}
		}
		helperLib.ReplyAndMention(adminsNum, adminsJID)
		helperLib.ReplyText("تم الإبلاغ عن الرسالة")
		return

	} else if msgContent == cmdOpe+"منشن الكل" {
		if !helperLib.IsUserAdmin(chat, author) {
			helperLib.ReplyText("حرك حرك تراك مو ادمن")
			return
		}

		if quotedMsg == nil {
			helperLib.ReplyText("الرجاء استخدام الأمر على الرسالة المراد منشنة اعضاء المجموعة عليها")
			return
		}
		text := "⚠️⚠️⚠️⚠️⚠️ مهم ⚠️⚠️⚠️⚠️⚠️"
		var usersJID []string
		users := helperLib.GetGroupMembers(chat)
		for _, user := range users {
			if user.JID.String() != botNum {
				text += "@" + strings.ReplaceAll(user.JID.ToNonAD().String(), "@s.whatsapp.net", "") + "\n"
				usersJID = append(usersJID, user.JID.ToNonAD().String())
			}
		}
		text += "⚠️⚠️⚠️⚠️⚠️ مهم ⚠️⚠️⚠️⚠️⚠️"
		helperLib.ReplyAndMention(text, usersJID)
		return

	} else if msgContentSplit[0] == cmdOpe+"خطة" {
		if len(msgContentSplit) != 2 {
			helperLib.ReplyText("استخدام خاطئ\nاكتب خطة مع اسم التخصص\n\nمثال: !خطة IS ")
			return
		}
		path := ""
		major := strings.ToUpper(msgContentSplit[1])
		switch major {
		case "CS":
			path = "./files/CS_PLAN.pdf"
			break
		case "IT":
			path = "./files/IT_PLAN.pdf"
			break
		case "IS":
			path = "./files/IS_PLAN.pdf"
			break
		}
		if path == "" {
			helperLib.ReplyText("تخصص غير معروف")
			return
		}
		helperLib.ReplyDocument(path)
		return
	} else if msgContent == cmdOpe+"درايف" {
		helperLib.ReplyText("Gone forever")
		return
	} else if msgContent == cmdOpe+"التقويم الأكاديمي" {
		helperLib.ReplyDocument("./files/CALENDAR.pdf")
		return
	} else if msgContent == cmdOpe+"شروط التحويل" {
		helperLib.ReplyDocument("./files/TRANSFERRING_CONDITIONS.pdf")
		return
	} else if msgContent == cmdOpe+"الفرق بين التخصصات" {
		helperLib.ReplyDocument("./files/DIFFERENCE_BETWEEN_MAJORS.pdf")
		return
	} else if msgContent == cmdOpe+"المسارات" {
		helperLib.ReplyDocument("./files/FCIT_TRACKS.pdf")
		return
	} else if msgContent == cmdOpe+"اقتراحات" {
		helperLib.ReplyText("يا هلا، اذا عندك اقتراحات تواصل مع مطوري على التيليجرام\n@ِxAbdaziz")
		return
	} else if msgContent == cmdOpe+"القاعات" {
		helperLib.ReplyText("رابط قاعات مواد الترم الأول 2024:\nhttps://cutt.us/Fcit202401")
		return
	} else if msgContent == cmdOpe+"الإجازة" {
		helperLib.Vacation()
	} else if msgContent == cmdOpe+"المكافأة" {
		helperLib.Allowance()
	} else if msgContent == cmdOpe+"المواد الاختيارية" {
		helperLib.ReplyDocument("./files/ELECTIVE_COURSES.pdf")
	} else if msgContent == cmdOpe+"broadcast" {
		if author == myNum {
			groups, _ := client.GetJoinedGroups()
			for i, group := range groups {
				_, _ = client.SendMessage(context.Background(), group.JID.ToNonAD(), &waProto.Message{Conversation: proto.String(quotedMsgText + string(i))})
			}
		}
	} else if msgContent == cmdOpe+"الجدول" {
		helperLib.ReplyText("https://betterkau.com")
		return
	} else if msgContent == cmdOpe+"القروبات" {
		helperLib.ReplyText("https://fcit-groups.abdaziz.dev")
		return
	}
}
