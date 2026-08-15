package msgHandler

func init() {
	RegisterCommand(Command{
		Name:        "!التقويم الأكاديمي",
		Description: "للحصول على التقويم الأكاديمي",
		Handler:     (*MessageContext).handleCalendar,
	})
}

func (mc *MessageContext) handleCalendar() {
	mc.HelperLib.ReplyDocument("./files/CALENDAR.pdf")
}
