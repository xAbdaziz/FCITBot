package msgHandler

func init() {
	RegisterCommand(Command{
		Name:        "!المسارات",
		Description: "للحصول على قائمة المسارات الموجودة بالكلية",
		Handler:     (*MessageContext).handleTracks,
	})
}

func (mc *MessageContext) handleTracks() {
	mc.HelperLib.ReplyDocument("./files/FCIT_TRACKS.pdf")
}
