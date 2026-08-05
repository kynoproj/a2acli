package cli

import "github.com/a2aproject/a2a-go/v2/a2a"

// buildUserMessage constructs a user-role text message, optionally attaching a
// task ID and/or context ID so the message continues an existing conversation.
// Empty taskID/contextID leave TaskInfo zero-valued, letting the server start a
// fresh task.
func buildUserMessage(text, taskID, contextID string) *a2a.Message {
	info := a2a.TaskInfo{TaskID: a2a.TaskID(taskID), ContextID: contextID}
	return a2a.NewMessageForTask(a2a.MessageRoleUser, info, a2a.NewTextPart(text))
}
