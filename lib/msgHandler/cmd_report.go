package msgHandler

import (
	"strings"
)

func init() {
	RegisterCommand(Command{
		Name:        "!تبليغ",
		Description: "للتبليغ عن رسالة مخالفة لمدراء المجموعة",
		Handler:     (*MessageContext).handleReport,
	})
}

func (mc *MessageContext) handleReport() {
	if mc.QuotedMsg == nil {
		mc.HelperLib.ReplyText("الرجاء استخدام الأمر على الرسالة المراد التبليغ عنها")
		return
	}
	adminsNum := ""
	var adminsJID []string
	admins := mc.HelperLib.GetGroupAdmins(mc.Chat)
	for _, admin := range admins {
		if admin.PhoneNumber.String() != mc.BotNum {
			adminsNum += "@" + strings.ReplaceAll(admin.JID.ToNonAD().String(), "@lid", "") + "\n"
			adminsJID = append(adminsJID, admin.JID.ToNonAD().String())
		}
	}
	mc.HelperLib.ReplyAndMention(adminsNum, adminsJID)
	mc.HelperLib.ReplyText("تم الإبلاغ عن الرسالة")
}
