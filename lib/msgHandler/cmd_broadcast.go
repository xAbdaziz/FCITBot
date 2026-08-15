package msgHandler

import (
	"strconv"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func init() {
	RegisterCommand(Command{
		Name:        "!broadcast",
		Description: "إرسال رسالة لجميع القروبات (خاص بالمطور)",
		Handler:     (*MessageContext).handleBroadcast,
	})
}

func (mc *MessageContext) handleBroadcast() {
	if mc.Author == mc.Owner.String() {
		groups, _ := mc.Client.GetJoinedGroups(mc.Ctx)
		for i, group := range groups {
			_, _ = mc.Client.SendMessage(mc.Ctx, group.JID.ToNonAD(), &waE2E.Message{Conversation: proto.String(mc.QuotedMsgText + strconv.Itoa(i))})
		}
	}
}
