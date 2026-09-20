package msgHandler

import (
	"fmt"
	"strings"
)

func init() {
	RegisterCommand(Command{
		Name:        "!الأوامر",
		Description: "للحصول على جميع الأوامر",
		Handler:     (*MessageContext).handleCommands,
	})
}

func (mc *MessageContext) handleCommands() {
	var sb strings.Builder
	sb.WriteString("📋 *قائمة الأوامر المتاحة:*\n\n")
	for _, cmd := range commands {
		sb.WriteString(fmt.Sprintf("• *%s*\n  %s\n\n", cmd.Name, cmd.Description))
	}
	mc.HelperLib.ReplyText(strings.TrimSpace(sb.String()))
}
