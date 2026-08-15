package msgHandler

func init() {
	RegisterCommand(Command{
		Name:        "!الفرق بين التخصصات",
		Description: "للحصول على الفروقات بين التخصصات",
		Handler:     (*MessageContext).handleMajorDifferences,
	})
}

func (mc *MessageContext) handleMajorDifferences() {
	mc.HelperLib.ReplyDocument("./files/DIFFERENCE_BETWEEN_MAJORS.pdf")
}
