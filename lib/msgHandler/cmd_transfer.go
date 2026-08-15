package msgHandler

func init() {
	RegisterCommand(Command{
		Name:        "!شروط التحويل",
		Description: "للحصول على شروط وآلية التحويل للكلية",
		Handler:     (*MessageContext).handleTransferConditions,
	})
}

func (mc *MessageContext) handleTransferConditions() {
	mc.HelperLib.ReplyDocument("./files/TRANSFERRING_CONDITIONS.pdf")
}
