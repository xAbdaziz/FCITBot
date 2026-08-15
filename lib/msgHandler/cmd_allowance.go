package msgHandler

func init() {
	RegisterCommand(Command{
		Name:        "!المكافأة",
		Description: "لأظهار الوقت المتبقي حتى المكافأة القادمة",
		Handler:     (*MessageContext).handleAllowance,
	})
}

func (mc *MessageContext) handleAllowance() {
	mc.HelperLib.Allowance()
}
