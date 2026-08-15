package msgHandler

func init() {
	RegisterCommand(Command{
		Name:        "!المواد الاختيارية",
		Description: "للحصول على قائمة المواد الاختيارية لكل تخصص",
		Handler:     (*MessageContext).handleElectiveCourses,
	})
}

func (mc *MessageContext) handleElectiveCourses() {
	mc.HelperLib.ReplyDocument("./files/ELECTIVE_COURSES.pdf")
}
