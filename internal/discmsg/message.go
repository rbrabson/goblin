package discmsg

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

// SendMessage is a utility routine used to send a message to a channel.
func SendMessage(s *discordgo.Session, channelID string, msg string, embeds []*discordgo.MessageEmbed) *discordgo.Message {
	log.Trace("--> SendMessage")
	defer log.Trace("<-- SendMessage")

	message, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: msg,
		Embeds:  embeds,
	})
	if err != nil {
		log.WithFields(log.Fields{"channelID": channelID}).Error("Unable to send a message")
		return nil
	}

	log.WithFields(log.Fields{"channel": channelID, "messageID": message.ID}).Error("sent message")
	return message
}

// EditMessage edits the current message in a channel.
func EditMessage(s *discordgo.Session, channelID string, messageID string, msg string, embeds []*discordgo.MessageEmbed) {
	log.Trace("--> EditMessage")
	defer log.Trace("<-- EditMessage")

	_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:      messageID,
		Channel: channelID,
		Content: &msg,
		Embeds:  &embeds,
	})
	if err != nil {
		log.WithFields(log.Fields{"channelID": channelID, "messageID": messageID}).Error("unable to edit a message")
		return
	}

	log.WithFields(log.Fields{"channel": channelID, "messageID": messageID}).Error("edited message")
}
