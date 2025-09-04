package msgHandler

import (
	"FCITBot/lib/helper"
	"FCITBot/models"

	"context"
	"os"
	"strconv"
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

type cmdHandler func(*events.Message, *whatsmeow.Client, *gorm.DB, *helper.Bot, types.JID, string, string, *waE2E.Message, string, string, types.JID, []string)

var commandMap = map[string]cmdHandler{
	"!الأوامر":            handleCommands,
	"!اطرد":               handleKick,
	"!احفظ":               handleSaveNote,
	"!هات":                handleGetNote,
	"!احذف":               handleDeleteNote,
	"!الملاحظات":          handleListNotes,
	"!تبليغ":              handleReport,
	"!منشن الكل":          handleMentionAll,
	"!خطة":                handlePlan,
	"!التقويم الأكاديمي":  handleCalendar,
	"!شروط التحويل":       handleTransferConditions,
	"!الفرق بين التخصصات": handleMajorDifferences,
	"!المسارات":           handleTracks,
	"!المكافأة":           handleAllowance,
	"!المواد الاختيارية":  handleElectiveCourses,
	"!broadcast":          handleBroadcast,
	"!الجدول":             handleSchedule,
	"!القروبات":          handleGroups,
}

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

	// Try exact command match first
	if handler, exists := commandMap[msgContent]; exists {
		handler(message, client, gormDB, helperLib, chat, author, botNum, quotedMsg, quotedMsgText, quotedMsgAuthor, owner.ToNonAD(), msgContentSplit)
		return
	}

	// Try prefix match for commands with arguments
	for cmd, handler := range commandMap {
		if strings.HasPrefix(msgContent, cmd+" ") {
			handler(message, client, gormDB, helperLib, chat, author, botNum, quotedMsg, quotedMsgText, quotedMsgAuthor, owner.ToNonAD(), msgContentSplit)
			return
		}
	}
}

func handleCommands(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB, helperLib *helper.Bot, chat types.JID, author, botNum string, quotedMsg *waE2E.Message, quotedMsgText, quotedMsgAuthor string, owner types.JID, msgContentSplit []string) {
	helperLib.ReplyText(cmds)
}

func handleKick(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB, helperLib *helper.Bot, chat types.JID, author, botNum string, quotedMsg *waE2E.Message, quotedMsgText, quotedMsgAuthor string, owner types.JID, msgContentSplit []string) {
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
	if quotedMsgAuthor == owner.String() {
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
	revokeMessage := client.BuildRevoke(chat, usertoKick, message.Message.ExtendedTextMessage.GetContextInfo().GetStanzaID())
	_, _ = client.SendMessage(context.Background(), chat, revokeMessage)
	helperLib.ReplyText("تم طرد العضو من المجموعة")
}

func handleSaveNote(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB, helperLib *helper.Bot, chat types.JID, author, botNum string, quotedMsg *waE2E.Message, quotedMsgText, quotedMsgAuthor string, owner types.JID, msgContentSplit []string) {
	if !helperLib.IsUserAdmin(chat, author) {
		helperLib.ReplyText("حرك حرك تراك مو ادمن")
		return
	}
	if len(msgContentSplit) != 2 || quotedMsg == nil {
		helperLib.ReplyText("استخدام خاطئ\nقم بالرد على الرسالة المراد حفظها ثم كتابة احفظ مع اسم الملاحظة بدون مسافة\n\nمثال: !احفظ اسم_الملاحظة")
		return
	}
	noteName := msgContentSplit[1]

	var finalMsgText string
	if quotedMsg != nil {
		if quotedMsg.GetExtendedTextMessage().GetText() != "" {
			finalMsgText = quotedMsg.GetExtendedTextMessage().GetText()
		} else if quotedMsg.Conversation != nil {
			finalMsgText = *quotedMsg.Conversation
		}
	}
	if finalMsgText != "" {
		note := models.GroupsNotes{GroupID: chat.String(), NoteName: noteName}
		var existing models.GroupsNotes
		err := gormDB.Where("group_id = ? AND note_name = ?", chat.String(), noteName).First(&existing).Error
		if err == nil {
			existing.NoteContent = finalMsgText
			gormDB.Save(&existing)
		} else {
			note.NoteContent = finalMsgText
			gormDB.Create(&note)
		}
		helperLib.ReplyText("تم حفظ الملاحظة \"" + noteName + "\"")
	} else {
		helperLib.ReplyText("مقدر احفظ غير النصوص حالياً")
	}
}

func handleGetNote(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB, helperLib *helper.Bot, chat types.JID, author, botNum string, quotedMsg *waE2E.Message, quotedMsgText, quotedMsgAuthor string, owner types.JID, msgContentSplit []string) {
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
}

func handleDeleteNote(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB, helperLib *helper.Bot, chat types.JID, author, botNum string, quotedMsg *waE2E.Message, quotedMsgText, quotedMsgAuthor string, owner types.JID, msgContentSplit []string) {
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
}

func handleListNotes(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB, helperLib *helper.Bot, chat types.JID, author, botNum string, quotedMsg *waE2E.Message, quotedMsgText, quotedMsgAuthor string, owner types.JID, msgContentSplit []string) {
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
}

func handleReport(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB, helperLib *helper.Bot, chat types.JID, author, botNum string, quotedMsg *waE2E.Message, quotedMsgText, quotedMsgAuthor string, owner types.JID, msgContentSplit []string) {
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
}

func handleMentionAll(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB, helperLib *helper.Bot, chat types.JID, author, botNum string, quotedMsg *waE2E.Message, quotedMsgText, quotedMsgAuthor string, owner types.JID, msgContentSplit []string) {
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
}

func handlePlan(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB, helperLib *helper.Bot, chat types.JID, author, botNum string, quotedMsg *waE2E.Message, quotedMsgText, quotedMsgAuthor string, owner types.JID, msgContentSplit []string) {
	if len(msgContentSplit) != 2 {
		helperLib.ReplyText("استخدام خاطئ\nاكتب خطة مع اسم التخصص\n\nمثال: !خطة IS ")
		return
	}
	path := ""
	major := strings.ToUpper(msgContentSplit[1])
	switch major {
	case "CS":
		path = "./files/CS_PLAN.pdf"
	case "IT":
		path = "./files/IT_PLAN.pdf"
	case "IS":
		path = "./files/IS_PLAN.pdf"
	}
	if path == "" {
		helperLib.ReplyText("تخصص غير معروف")
		return
	}
	helperLib.ReplyDocument(path)
}

func handleCalendar(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB, helperLib *helper.Bot, chat types.JID, author, botNum string, quotedMsg *waE2E.Message, quotedMsgText, quotedMsgAuthor string, owner types.JID, msgContentSplit []string) {
	helperLib.ReplyDocument("./files/CALENDAR.pdf")
}

func handleTransferConditions(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB, helperLib *helper.Bot, chat types.JID, author, botNum string, quotedMsg *waE2E.Message, quotedMsgText, quotedMsgAuthor string, owner types.JID, msgContentSplit []string) {
	helperLib.ReplyDocument("./files/TRANSFERRING_CONDITIONS.pdf")
}

func handleMajorDifferences(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB, helperLib *helper.Bot, chat types.JID, author, botNum string, quotedMsg *waE2E.Message, quotedMsgText, quotedMsgAuthor string, owner types.JID, msgContentSplit []string) {
	helperLib.ReplyDocument("./files/DIFFERENCE_BETWEEN_MAJORS.pdf")
}

func handleTracks(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB, helperLib *helper.Bot, chat types.JID, author, botNum string, quotedMsg *waE2E.Message, quotedMsgText, quotedMsgAuthor string, owner types.JID, msgContentSplit []string) {
	helperLib.ReplyDocument("./files/FCIT_TRACKS.pdf")
}

func handleAllowance(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB, helperLib *helper.Bot, chat types.JID, author, botNum string, quotedMsg *waE2E.Message, quotedMsgText, quotedMsgAuthor string, owner types.JID, msgContentSplit []string) {
	helperLib.Allowance()
}

func handleElectiveCourses(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB, helperLib *helper.Bot, chat types.JID, author, botNum string, quotedMsg *waE2E.Message, quotedMsgText, quotedMsgAuthor string, owner types.JID, msgContentSplit []string) {
	helperLib.ReplyDocument("./files/ELECTIVE_COURSES.pdf")
}

func handleBroadcast(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB, helperLib *helper.Bot, chat types.JID, author, botNum string, quotedMsg *waE2E.Message, quotedMsgText, quotedMsgAuthor string, owner types.JID, msgContentSplit []string) {
	if author == owner.String() {
		groups, _ := client.GetJoinedGroups()
		for i, group := range groups {
			_, _ = client.SendMessage(context.Background(), group.JID.ToNonAD(), &waE2E.Message{Conversation: proto.String(quotedMsgText + strconv.Itoa(i))})
		}
	}
}

func handleSchedule(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB, helperLib *helper.Bot, chat types.JID, author, botNum string, quotedMsg *waE2E.Message, quotedMsgText, quotedMsgAuthor string, owner types.JID, msgContentSplit []string) {
	helperLib.ReplyText("https://betterkau.com")
}

func handleGroups(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB, helperLib *helper.Bot, chat types.JID, author, botNum string, quotedMsg *waE2E.Message, quotedMsgText, quotedMsgAuthor string, owner types.JID, msgContentSplit []string) {
	helperLib.ReplyText("https://fcit-groups.abdaziz.dev")
}