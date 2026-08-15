package msgHandler

import (
	"FCITBot/models"
)

func init() {
	RegisterCommand(Command{
		Name:        "!احفظ",
		Description: "لحفظ ملاحظة في الملاحظات",
		Handler:     (*MessageContext).handleSaveNote,
	})
	RegisterCommand(Command{
		Name:        "!هات",
		Description: "لإسترجاع ملاحظة من الملاحظات",
		Handler:     (*MessageContext).handleGetNote,
	})
	RegisterCommand(Command{
		Name:        "!احذف",
		Description: "لحذف ملاحظة من الملاحظات",
		Handler:     (*MessageContext).handleDeleteNote,
	})
	RegisterCommand(Command{
		Name:        "!الملاحظات",
		Description: "يستخدم هذا الأمر لإظهار الملاحظات المحفوظة",
		Handler:     (*MessageContext).handleListNotes,
	})
}

func (mc *MessageContext) handleSaveNote() {
	if !mc.HelperLib.IsUserAdmin(mc.Chat, mc.Author) {
		mc.HelperLib.ReplyText("عذراً، هذا الأمر متاح للمشرفين فقط")
		return
	}
	if len(mc.MsgContentSplit) != 2 || mc.QuotedMsg == nil {
		mc.HelperLib.ReplyText("استخدام خاطئ\nقم بالرد على الرسالة المراد حفظها ثم كتابة احفظ مع اسم الملاحظة بدون مسافة\n\nمثال: !احفظ اسم_الملاحظة")
		return
	}
	noteName := mc.MsgContentSplit[1]

	var finalMsgText string
	if mc.QuotedMsg != nil {
		if mc.QuotedMsg.GetExtendedTextMessage().GetText() != "" {
			finalMsgText = mc.QuotedMsg.GetExtendedTextMessage().GetText()
		} else if mc.QuotedMsg.Conversation != nil {
			finalMsgText = *mc.QuotedMsg.Conversation
		}
	}
	if finalMsgText != "" {
		note := models.GroupsNotes{GroupID: mc.Chat.String(), NoteName: noteName}
		var existing models.GroupsNotes
		err := mc.GormDB.Where("group_id = ? AND note_name = ?", mc.Chat.String(), noteName).First(&existing).Error
		if err == nil {
			existing.NoteContent = finalMsgText
			mc.GormDB.Save(&existing)
		} else {
			note.NoteContent = finalMsgText
			mc.GormDB.Create(&note)
		}
		mc.HelperLib.ReplyText("تم حفظ الملاحظة \"" + noteName + "\"")
	} else {
		mc.HelperLib.ReplyText("مقدر احفظ غير النصوص حالياً")
	}
}

func (mc *MessageContext) handleGetNote() {
	if len(mc.MsgContentSplit) != 2 {
		mc.HelperLib.ReplyText("استخدام خاطئ\nاكتب هات مع اسم الملاحظة بدون مسافة\n\nمثال: !هات اسم_الملاحظة ")
		return
	}
	noteName := mc.MsgContentSplit[1]
	var note models.GroupsNotes
	err := mc.GormDB.Where("group_id = ? AND note_name = ?", mc.Chat.String(), noteName).First(&note).Error
	if err != nil {
		mc.HelperLib.ReplyText("لا توجد ملاحظة بأسم " + "\"" + noteName + "\"")
		return
	}
	mc.HelperLib.ReplyText(note.NoteContent)
}

func (mc *MessageContext) handleDeleteNote() {
	if !mc.HelperLib.IsUserAdmin(mc.Chat, mc.Author) {
		mc.HelperLib.ReplyText("عذراً، هذا الأمر متاح للمشرفين فقط")
		return
	}
	if len(mc.MsgContentSplit) != 2 {
		mc.HelperLib.ReplyText("استخدام خاطئ\nاكتب احذف مع اسم الملاحظة بدون مسافة\n\nمثال: !احذف اسم_الملاحظة ")
		return
	}
	noteName := mc.MsgContentSplit[1]
	var note models.GroupsNotes
	err := mc.GormDB.Where("group_id = ? AND note_name = ?", mc.Chat.String(), noteName).First(&note).Error
	if err != nil {
		mc.HelperLib.ReplyText("لا توجد ملاحظة بأسم " + "\"" + noteName + "\"")
		return
	}
	mc.GormDB.Delete(&note)
	mc.HelperLib.ReplyText("تم حذف الملاحظة " + "\"" + noteName + "\"")
}

func (mc *MessageContext) handleListNotes() {
	var notes []models.GroupsNotes
	mc.GormDB.Where("group_id = ?", mc.Chat.String()).Find(&notes)
	if len(notes) == 0 {
		mc.HelperLib.ReplyText("لا توجد ملاحظات محفوظة.")
		return
	}
	notesList := "الملاحظات المحفوظة:"
	for _, n := range notes {
		notesList += "\n- " + n.NoteName
	}
	mc.HelperLib.ReplyText(notesList)
}
