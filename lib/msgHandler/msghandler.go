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

type MessageContext struct {
	message         *events.Message
	client          *whatsmeow.Client
	gormDB          *gorm.DB
	helperLib       *helper.Bot
	chat            types.JID
	author          string
	botNum          string
	quotedMsg       *waE2E.Message
	quotedMsgText   string
	quotedMsgAuthor string
	owner           types.JID
	msgContentSplit []string
	ctx             context.Context
}

type cmdHandler func(*MessageContext)

var commandMap = map[string]cmdHandler{
	"!الأوامر":            (*MessageContext).handleCommands,
	"!اطرد":               (*MessageContext).handleKick,
	"!احفظ":               (*MessageContext).handleSaveNote,
	"!هات":                (*MessageContext).handleGetNote,
	"!احذف":               (*MessageContext).handleDeleteNote,
	"!الملاحظات":          (*MessageContext).handleListNotes,
	"!تبليغ":              (*MessageContext).handleReport,
	"!منشن الكل":          (*MessageContext).handleMentionAll,
	"!خطة":                (*MessageContext).handlePlan,
	"!التقويم الأكاديمي":  (*MessageContext).handleCalendar,
	"!شروط التحويل":       (*MessageContext).handleTransferConditions,
	"!الفرق بين التخصصات": (*MessageContext).handleMajorDifferences,
	"!المسارات":           (*MessageContext).handleTracks,
	"!المكافأة":           (*MessageContext).handleAllowance,
	"!المواد الاختيارية":  (*MessageContext).handleElectiveCourses,
	"!broadcast":          (*MessageContext).handleBroadcast,
	"!الجدول":             (*MessageContext).handleSchedule,
	"!القروبات":           (*MessageContext).handleGroups,
}

func Handle(message *events.Message, client *whatsmeow.Client, gormDB *gorm.DB) {
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
		message:         message,
		client:          client,
		gormDB:          gormDB,
		helperLib:       helperLib,
		chat:            chat,
		author:          author,
		botNum:          botNum,
		quotedMsg:       quotedMsg,
		quotedMsgText:   quotedMsgText,
		quotedMsgAuthor: quotedMsgAuthor,
		owner:           owner.ToNonAD(),
		msgContentSplit: msgContentSplit,
		ctx:             ctx,
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

func (mc *MessageContext) handleCommands() {
	mc.helperLib.ReplyText(cmds)
}

func (mc *MessageContext) handleKick() {
	if !mc.helperLib.IsUserAdmin(mc.chat, mc.author) {
		mc.helperLib.ReplyText("عذراً، هذا الأمر متاح للمشرفين فقط")
		return
	}
	if !mc.helperLib.IsUserAdmin(mc.chat, mc.botNum) {
		mc.helperLib.ReplyText("عذراً، لا أملك صلاحيات المشرف في هذه المجموعة")
		return
	}
	if mc.quotedMsg == nil {
		mc.helperLib.ReplyText("الرجاء تحديد العضو المراد طرده بالرد على رسالته")
		return
	}
	if mc.quotedMsgAuthor == mc.botNum {
		mc.helperLib.ReplyText("لا يمكنني طرد نفسي من المجموعة")
		return
	}
	if mc.quotedMsgAuthor == mc.owner.String() {
		mc.helperLib.ReplyText("عذراً، لا يمكن طرد مطور البوت")
		return
	}
	if !mc.helperLib.MemberIsInGroup(mc.chat, mc.quotedMsgAuthor) {
		mc.helperLib.ReplyText("العضو غير موجود بالمجموعة")
		return
	}
	if mc.helperLib.IsUserAdmin(mc.chat, mc.quotedMsgAuthor) {
		mc.helperLib.ReplyText("عذراً، لا يمكن طرد المشرفين")
		return
	}
	usertoKick, _ := types.ParseJID(mc.quotedMsgAuthor)
	_, _ = mc.client.UpdateGroupParticipants(mc.ctx, mc.chat, []types.JID{usertoKick}, whatsmeow.ParticipantChangeRemove)
	revokeMessage := mc.client.BuildRevoke(mc.chat, usertoKick, mc.message.Message.ExtendedTextMessage.GetContextInfo().GetStanzaID())
	_, _ = mc.client.SendMessage(mc.ctx, mc.chat, revokeMessage)
	mc.helperLib.ReplyText("تم طرد العضو من المجموعة")
}

func (mc *MessageContext) handleSaveNote() {
	if !mc.helperLib.IsUserAdmin(mc.chat, mc.author) {
		mc.helperLib.ReplyText("عذراً، هذا الأمر متاح للمشرفين فقط")
		return
	}
	if len(mc.msgContentSplit) != 2 || mc.quotedMsg == nil {
		mc.helperLib.ReplyText("استخدام خاطئ\nقم بالرد على الرسالة المراد حفظها ثم كتابة احفظ مع اسم الملاحظة بدون مسافة\n\nمثال: !احفظ اسم_الملاحظة")
		return
	}
	noteName := mc.msgContentSplit[1]

	var finalMsgText string
	if mc.quotedMsg != nil {
		if mc.quotedMsg.GetExtendedTextMessage().GetText() != "" {
			finalMsgText = mc.quotedMsg.GetExtendedTextMessage().GetText()
		} else if mc.quotedMsg.Conversation != nil {
			finalMsgText = *mc.quotedMsg.Conversation
		}
	}
	if finalMsgText != "" {
		note := models.GroupsNotes{GroupID: mc.chat.String(), NoteName: noteName}
		var existing models.GroupsNotes
		err := mc.gormDB.Where("group_id = ? AND note_name = ?", mc.chat.String(), noteName).First(&existing).Error
		if err == nil {
			existing.NoteContent = finalMsgText
			mc.gormDB.Save(&existing)
		} else {
			note.NoteContent = finalMsgText
			mc.gormDB.Create(&note)
		}
		mc.helperLib.ReplyText("تم حفظ الملاحظة \"" + noteName + "\"")
	} else {
		mc.helperLib.ReplyText("مقدر احفظ غير النصوص حالياً")
	}
}

func (mc *MessageContext) handleGetNote() {
	if len(mc.msgContentSplit) != 2 {
		mc.helperLib.ReplyText("استخدام خاطئ\nاكتب هات مع اسم الملاحظة بدون مسافة\n\nمثال: !هات اسم_الملاحظة ")
		return
	}
	noteName := mc.msgContentSplit[1]
	var note models.GroupsNotes
	err := mc.gormDB.Where("group_id = ? AND note_name = ?", mc.chat.String(), noteName).First(&note).Error
	if err != nil {
		mc.helperLib.ReplyText("لا توجد ملاحظة بأسم " + "\"" + noteName + "\"")
		return
	}
	mc.helperLib.ReplyText(note.NoteContent)
}

func (mc *MessageContext) handleDeleteNote() {
	if !mc.helperLib.IsUserAdmin(mc.chat, mc.author) {
		mc.helperLib.ReplyText("عذراً، هذا الأمر متاح للمشرفين فقط")
		return
	}
	if len(mc.msgContentSplit) != 2 {
		mc.helperLib.ReplyText("استخدام خاطئ\nاكتب احذف مع اسم الملاحظة بدون مسافة\n\nمثال: !احذف اسم_الملاحظة ")
		return
	}
	noteName := mc.msgContentSplit[1]
	var note models.GroupsNotes
	err := mc.gormDB.Where("group_id = ? AND note_name = ?", mc.chat.String(), noteName).First(&note).Error
	if err != nil {
		mc.helperLib.ReplyText("لا توجد ملاحظة بأسم " + "\"" + noteName + "\"")
		return
	}
	mc.gormDB.Delete(&note)
	mc.helperLib.ReplyText("تم حذف الملاحظة " + "\"" + noteName + "\"")
}

func (mc *MessageContext) handleListNotes() {
	var notes []models.GroupsNotes
	mc.gormDB.Where("group_id = ?", mc.chat.String()).Find(&notes)
	if len(notes) == 0 {
		mc.helperLib.ReplyText("لا توجد ملاحظات محفوظة.")
		return
	}
	notesList := "الملاحظات المحفوظة:"
	for _, n := range notes {
		notesList += "\n- " + n.NoteName
	}
	mc.helperLib.ReplyText(notesList)
}

func (mc *MessageContext) handleReport() {
	if mc.quotedMsg == nil {
		mc.helperLib.ReplyText("الرجاء استخدام الأمر على الرسالة المراد التبليغ عنها")
		return
	}
	adminsNum := ""
	var adminsJID []string
	admins := mc.helperLib.GetGroupAdmins(mc.chat)
	for _, admin := range admins {
		if admin.PhoneNumber.String() != mc.botNum {
			adminsNum += "@" + strings.ReplaceAll(admin.JID.ToNonAD().String(), "@lid", "") + "\n"
			adminsJID = append(adminsJID, admin.JID.ToNonAD().String())
		}
	}
	mc.helperLib.ReplyAndMention(adminsNum, adminsJID)
	mc.helperLib.ReplyText("تم الإبلاغ عن الرسالة")
}

func (mc *MessageContext) handleMentionAll() {
	if !mc.helperLib.IsUserAdmin(mc.chat, mc.author) {
		mc.helperLib.ReplyText("عذراً، هذا الأمر متاح للمشرفين فقط")
		return
	}

	if mc.quotedMsg == nil {
		mc.helperLib.ReplyText("الرجاء استخدام الأمر على الرسالة المراد منشنة اعضاء المجموعة عليها")
		return
	}
	text := "⚠️ إعلان مهم ⚠️\n\n"
	var usersJID []string
	users := mc.helperLib.GetGroupMembers(mc.chat)
	for _, user := range users {
		if user.PhoneNumber.String() != mc.botNum {
			text += "@" + strings.ReplaceAll(user.JID.ToNonAD().String(), "@lid", "") + "\n"
			usersJID = append(usersJID, user.JID.ToNonAD().String())
		}
	}
	text += "\n⚠️ يرجى الاطلاع على الرسالة أعلاه ⚠️"
	mc.helperLib.ReplyAndMention(text, usersJID)
}

func (mc *MessageContext) handlePlan() {
	if len(mc.msgContentSplit) != 2 {
		mc.helperLib.ReplyText("استخدام خاطئ\nاكتب خطة مع اسم التخصص\n\nمثال: !خطة IS ")
		return
	}
	path := ""
	major := strings.ToUpper(mc.msgContentSplit[1])
	switch major {
	case "CS":
		path = "./files/CS_PLAN.pdf"
	case "IT":
		path = "./files/IT_PLAN.pdf"
	case "IS":
		path = "./files/IS_PLAN.pdf"
	}
	if path == "" {
		mc.helperLib.ReplyText("تخصص غير معروف")
		return
	}
	mc.helperLib.ReplyDocument(path)
}

func (mc *MessageContext) handleCalendar() {
	mc.helperLib.ReplyDocument("./files/CALENDAR.pdf")
}

func (mc *MessageContext) handleTransferConditions() {
	mc.helperLib.ReplyDocument("./files/TRANSFERRING_CONDITIONS.pdf")
}

func (mc *MessageContext) handleMajorDifferences() {
	mc.helperLib.ReplyDocument("./files/DIFFERENCE_BETWEEN_MAJORS.pdf")
}

func (mc *MessageContext) handleTracks() {
	mc.helperLib.ReplyDocument("./files/FCIT_TRACKS.pdf")
}

func (mc *MessageContext) handleAllowance() {
	mc.helperLib.Allowance()
}

func (mc *MessageContext) handleElectiveCourses() {
	mc.helperLib.ReplyDocument("./files/ELECTIVE_COURSES.pdf")
}

func (mc *MessageContext) handleBroadcast() {
	if mc.author == mc.owner.String() {
		groups, _ := mc.client.GetJoinedGroups(mc.ctx)
		for i, group := range groups {
			_, _ = mc.client.SendMessage(mc.ctx, group.JID.ToNonAD(), &waE2E.Message{Conversation: proto.String(mc.quotedMsgText + strconv.Itoa(i))})
		}
	}
}

func (mc *MessageContext) handleSchedule() {
	mc.helperLib.ReplyText("https://betterkau.com")
}

func (mc *MessageContext) handleGroups() {
	mc.helperLib.ReplyText("https://fcit-groups.abdaziz.dev")
}
