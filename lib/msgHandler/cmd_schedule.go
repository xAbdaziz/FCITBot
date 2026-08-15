package msgHandler

func init() {
	RegisterCommand(Command{
		Name:        "!الجدول",
		Description: "للحصول على رابط موقع KauIndex",
		Handler:     (*MessageContext).handleSchedule,
	})
}

func (mc *MessageContext) handleSchedule() {
	mc.HelperLib.ReplyText("https://kauindex.com")
}
