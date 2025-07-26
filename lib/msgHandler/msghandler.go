package msgHandler

import (
	"FCITBot/lib/helper"
	"FCITBot/models"

	"context"
	"os"
	"strings"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"

	"gorm.io/gorm"
)

var cmdsFile, _ = os.ReadFile("cmds.txt")
var cmds = string(cmdsFile)

func Handle(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB) {
	helperLib := helper.BotContext(client, message, gormDB)
	if message.Info.IsFromMe {
		return
	}
	if !message.Info.IsGroup {
		helperLib.ReplyText("يا هلا، معاك بوت الحاسبات\nرجاءً تحدث معي في القروب فقط")
		return
	}

	ownerJID, _ := types.ParseJID(os.Getenv("OWNER_NUMBER"))
	owner, err := client.Store.LIDs.GetLIDForPN(context.Background(), ownerJID)
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

	const cmdOpe = "!"
	command := msgContentSplit[0]

	switch command {
	case cmdOpe + "الأوامر":
		helperLib.ReplyText(cmds)
		return

	case cmdOpe + "اطرد":
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
		if quotedMsgAuthor == owner.ToNonAD().String() {
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
		revokeMessage := client.BuildRevoke(chat, usertoKick, quotedMsgContext.GetStanzaID())
		_, _ = client.SendMessage(context.Background(), chat, revokeMessage)
		helperLib.ReplyText("تم طرد العضو من المجموعة")
		return

	case cmdOpe + "احفظ":
		if !helperLib.IsUserAdmin(chat, author) {
			helperLib.ReplyText("حرك حرك تراك مو ادمن")
			return
		}
		if len(msgContentSplit) != 2 || quotedMsg == nil {
			helperLib.ReplyText("استخدام خاطئ\nقم بالرد على الرسالة المراد حفظها ثم كتابة احفظ مع اسم الملاحظة بدون مسافة\n\nمثال: !احفظ اسم_الملاحظة")
			return
		}
		noteName := msgContentSplit[1]

		if quotedMsg != nil && (quotedMsg.Conversation != nil || (quotedMsg.GetExtendedTextMessage() != nil && quotedMsg.GetExtendedTextMessage().Text != nil)) {
			if extendedTextMsg := quotedMsg.GetExtendedTextMessage(); extendedTextMsg != nil && extendedTextMsg.Text != nil {
				quotedMsgText = extendedTextMsg.GetText()
			} else {
				quotedMsgText = *quotedMsg.Conversation
			}
			// GORM: Upsert note
			note := models.GroupsNotes{GroupID: chat.String(), NoteName: noteName}
			var existing models.GroupsNotes
			err := gormDB.Where("group_id = ? AND note_name = ?", chat.String(), noteName).First(&existing).Error
			if err == nil {
				existing.NoteContent = quotedMsgText
				gormDB.Save(&existing)
			} else {
				note.NoteContent = quotedMsgText
				gormDB.Create(&note)
			}
			helperLib.ReplyText("تم حفظ الملاحظة \"" + noteName + "\"")
			return
		} else {
			helperLib.ReplyText("مقدر احفظ غير النصوص حالياً")
			return
		}

	case cmdOpe + "هات":
		if len(msgContentSplit) != 2 {
			helperLib.ReplyText("استخدام خاطئ\nاكتب هات مع اسم الملاحظة بدون مسافة\n\nمثال: !هات اسم_الملاحظة ")
			return
		}
		noteName := msgContentSplit[1]
		var note models.GroupsNotes
		err := gormDB.Where("group_id = ? AND note_name = ?", chat.String(), noteName).First(&note).Error
		if err != nil {
			helperLib.ReplyText("لا توجد ملاحظة بأسم " + "\"" + noteName + "\"")
			return
		}
		helperLib.ReplyText(note.NoteContent)
		return

	case cmdOpe + "احذف":
		if !helperLib.IsUserAdmin(chat, author) {
			helperLib.ReplyText("حرك حرك تراك مو ادمن")
			return
		}
		if len(msgContentSplit) != 2 {
			helperLib.ReplyText("استخدام خاطئ\nاكتب احذف مع اسم الملاحظة بدون مسافة\n\nمثال: !احذف اسم_الملاحظة ")
			return
		}
		noteName := msgContentSplit[1]
		var note models.GroupsNotes
		err := gormDB.Where("group_id = ? AND note_name = ?", chat.String(), noteName).First(&note).Error
		if err != nil {
			helperLib.ReplyText("لا توجد ملاحظة بأسم " + "\"" + noteName + "\"")
			return
		}
		gormDB.Delete(&note)
		helperLib.ReplyText("تم حذف الملاحظة " + "\"" + noteName + "\"")
		return

	case cmdOpe + "الملاحظات":
		var notes []models.GroupsNotes
		gormDB.Where("group_id = ?", chat.String()).Find(&notes)
		if len(notes) == 0 {
			helperLib.ReplyText("لا توجد ملاحظات محفوظة.")
			return
		}
		notesList := "الملاحظات المحفوظة:"
		for _, n := range notes {
			notesList += "\n- " + n.NoteName
		}
		helperLib.ReplyText(notesList)
		return

	case cmdOpe + "تبليغ":
		if quotedMsg == nil {
			helperLib.ReplyText("الرجاء استخدام الأمر على الرسالة المراد التبليغ عنها")
			return
		}
		adminsNum := ""
		var adminsJID []string
		admins := helperLib.GetGroupAdmins(chat)
		for _, admin := range admins {
			if admin.PhoneNumber.String() != botNum {
				adminsNum += "@" + strings.ReplaceAll(admin.JID.ToNonAD().String(), "@lid", "") + "\n"
				adminsJID = append(adminsJID, admin.JID.ToNonAD().String())
			}
		}
		helperLib.ReplyAndMention(adminsNum, adminsJID)
		helperLib.ReplyText("تم الإبلاغ عن الرسالة")
		return

	case cmdOpe + "منشن الكل":
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
			if user.PhoneNumber.String() != botNum {
				text += "@" + strings.ReplaceAll(user.JID.ToNonAD().String(), "@lid", "") + "\n"
				usersJID = append(usersJID, user.JID.ToNonAD().String())
			}
		}
		text += "⚠️⚠️⚠️⚠️⚠️ مهم ⚠️⚠️⚠️⚠️⚠️"
		helperLib.ReplyAndMention(text, usersJID)
		return

	case cmdOpe + "خطة":
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

	case cmdOpe + "درايف":
		helperLib.ReplyText("Gone forever")
		return

	case cmdOpe + "التقويم الأكاديمي":
		helperLib.ReplyDocument("./files/CALENDAR.pdf")
		return

	case cmdOpe + "شروط التحويل":
		helperLib.ReplyDocument("./files/TRANSFERRING_CONDITIONS.pdf")
		return

	case cmdOpe + "الفرق بين التخصصات":
		helperLib.ReplyDocument("./files/DIFFERENCE_BETWEEN_MAJORS.pdf")
		return

	case cmdOpe + "المسارات":
		helperLib.ReplyDocument("./files/FCIT_TRACKS.pdf")
		return

	case cmdOpe + "اقتراحات":
		helperLib.ReplyText("يا هلا، اذا عندك اقتراحات تواصل مع مطوري على التيليجرام\n@ِxAbdaziz")
		return

	case cmdOpe + "القاعات":
		helperLib.ReplyText("رابط قاعات مواد الترم الأول 2024:\nhttps://cutt.us/Fcit202401")
		return

	case cmdOpe + "الإجازة":
		//helperLib.Vacation()

	case cmdOpe + "المكافأة":
		helperLib.Allowance()

	case cmdOpe + "المواد الاختيارية":
		helperLib.ReplyDocument("./files/ELECTIVE_COURSES.pdf")

	case cmdOpe + "broadcast":
		if author == owner.ToNonAD().String() {
			groups, _ := client.GetJoinedGroups()
			for i, group := range groups {
				_, _ = client.SendMessage(context.Background(), group.JID.ToNonAD(), &waE2E.Message{Conversation: proto.String(quotedMsgText + string(i))})
			}
		}

	case cmdOpe + "الجدول":
		helperLib.ReplyText("https://betterkau.com")
		return

	case cmdOpe + "القروبات":
		helperLib.ReplyText("https://fcit-groups.abdaziz.dev")
		return
	}
}
