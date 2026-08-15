package msgHandler

import (
	"strings"
)

func init() {
	RegisterCommand(Command{
		Name:        "!منشن الكل",
		Description: "لمنشنة جميع اعضاء المجموعة",
		Handler:     (*MessageContext).handleMentionAll,
	})
}

func (mc *MessageContext) handleMentionAll() {
	if !mc.HelperLib.IsUserAdmin(mc.Chat, mc.Author) {
		mc.HelperLib.ReplyText("عذراً، هذا الأمر متاح للمشرفين فقط")
		return
	}

	if mc.QuotedMsg == nil {
		mc.HelperLib.ReplyText("الرجاء استخدام الأمر على الرسالة المراد منشنة اعضاء المجموعة عليها")
		return
	}
	text := "⚠️ إعلان مهم ⚠️\n\n"
	var usersJID []string
	users := mc.HelperLib.GetGroupMembers(mc.Chat)
	for _, user := range users {
		if user.PhoneNumber.String() != mc.BotNum {
			text += "@" + strings.ReplaceAll(user.JID.ToNonAD().String(), "@lid", "") + "\n"
			usersJID = append(usersJID, user.JID.ToNonAD().String())
		}
	}
	text += "\n⚠️ يرجى الاطلاع على الرسالة أعلاه ⚠️"
	mc.HelperLib.ReplyAndMention(text, usersJID)
}
