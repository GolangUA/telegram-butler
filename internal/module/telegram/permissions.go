package telegram

import "github.com/mymmrac/telego"

// DenyAllPermissions returns ChatPermissions with all permissions set to false.
func DenyAllPermissions() telego.ChatPermissions {
	return telego.ChatPermissions{
		CanSendMessages:       new(false),
		CanSendAudios:         new(false),
		CanSendDocuments:      new(false),
		CanSendPhotos:         new(false),
		CanSendVideos:         new(false),
		CanSendVideoNotes:     new(false),
		CanSendVoiceNotes:     new(false),
		CanSendPolls:          new(false),
		CanSendOtherMessages:  new(false),
		CanAddWebPagePreviews: new(false),
		CanEditTag:            new(false),
		CanChangeInfo:         new(false),
		CanInviteUsers:        new(false),
		CanPinMessages:        new(false),
		CanManageTopics:       new(false),
	}
}
