package msgHandler

func init() {
	RegisterCommand(Command{
		Name:        "!القروبات",
		Description: "للحصول على رابط قروبات الكلية",
		Handler:     (*MessageContext).handleGroups,
	})
}

func (mc *MessageContext) handleGroups() {
	mc.HelperLib.ReplyText("https://groups.kstacks.org")
}
