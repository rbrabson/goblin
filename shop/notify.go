package shop

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

// NotifyUser sends a direct message to a member on Discord.
func SendMessageToUser(s *discordgo.Session, memberID string, msg string) {
	c, err := s.UserChannelCreate(memberID)
	if err != nil {
		log.WithFields(log.Fields{"member": memberID, "error": err}).Warn("error creating private channel")
		return
	}

	_, err = s.ChannelMessageSend(c.ID, msg)
	if err != nil {
		log.WithFields(log.Fields{"channel": c.ID, "member": memberID, "message": msg, "error": err.Error()}).Error("failed to send mesage to the member")
	}
}
