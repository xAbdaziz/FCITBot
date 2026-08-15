package msgHandler

import (
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

func init() {
	RegisterCommand(Command{
		Name:        "!اطرد",
		Description: "يستخدم هذا الأمر لطرد شخص من المجموعة",
		Handler:     (*MessageContext).handleKick,
	})
}

func (mc *MessageContext) handleKick() {
	if !mc.HelperLib.IsUserAdmin(mc.Chat, mc.Author) {
		mc.HelperLib.ReplyText("عذراً، هذا الأمر متاح للمشرفين فقط")
		return
	}
	if !mc.HelperLib.IsUserAdmin(mc.Chat, mc.BotNum) {
		mc.HelperLib.ReplyText("عذراً، لا أملك صلاحيات المشرف في هذه المجموعة")
		return
	}
	if mc.QuotedMsg == nil {
		mc.HelperLib.ReplyText("الرجاء تحديد العضو المراد طرده بالرد على رسالته")
		return
	}
	if mc.QuotedMsgAuthor == mc.BotNum {
		mc.HelperLib.ReplyText("لا يمكنني طرد نفسي من المجموعة")
		return
	}
	if mc.QuotedMsgAuthor == mc.Owner.String() {
		mc.HelperLib.ReplyText("عذراً، لا يمكن طرد مطور البوت")
		return
	}
	if !mc.HelperLib.MemberIsInGroup(mc.Chat, mc.QuotedMsgAuthor) {
		mc.HelperLib.ReplyText("العضو غير موجود بالمجموعة")
		return
	}
	if mc.HelperLib.IsUserAdmin(mc.Chat, mc.QuotedMsgAuthor) {
		mc.HelperLib.ReplyText("عذراً، لا يمكن طرد المشرفين")
		return
	}
	usertoKick, _ := types.ParseJID(mc.QuotedMsgAuthor)
	_, _ = mc.Client.UpdateGroupParticipants(mc.Ctx, mc.Chat, []types.JID{usertoKick}, whatsmeow.ParticipantChangeRemove)
	revokeMessage := mc.Client.BuildRevoke(mc.Chat, usertoKick, mc.Message.Message.ExtendedTextMessage.GetContextInfo().GetStanzaID())
	_, _ = mc.Client.SendMessage(mc.Ctx, mc.Chat, revokeMessage)
	mc.HelperLib.ReplyText("تم طرد العضو من المجموعة")
}
