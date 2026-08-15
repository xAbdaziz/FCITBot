package msgHandler

import (
	"strings"
)

func init() {
	RegisterCommand(Command{
		Name:        "!خطة",
		Description: "للحصول على خطة التخصصات",
		Handler:     (*MessageContext).handlePlan,
	})
}

func (mc *MessageContext) handlePlan() {
	if len(mc.MsgContentSplit) != 2 {
		mc.HelperLib.ReplyText("استخدام خاطئ\nاكتب خطة مع اسم التخصص\n\nمثال: !خطة IS ")
		return
	}
	path := ""
	major := strings.ToUpper(mc.MsgContentSplit[1])
	switch major {
	case "CS":
		path = "./files/CS_PLAN.pdf"
	case "IT":
		path = "./files/IT_PLAN.pdf"
	case "IS":
		path = "./files/IS_PLAN.pdf"
	}
	if path == "" {
		mc.HelperLib.ReplyText("تخصص غير معروف")
		return
	}
	mc.HelperLib.ReplyDocument(path)
}
