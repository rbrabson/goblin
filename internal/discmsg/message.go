package discmsg

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

// SendMessage is a utility routine used to send a message to a channel.
func SendMessage(s *discordgo.Session, channelID string, msg string, components []discordgo.MessageComponent, embeds []*discordgo.MessageEmbed) *discordgo.Message {
	log.Trace("--> SendMessage")
	defer log.Trace("<-- SendMessage")

	message, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content:    msg,
		Components: components,
		Embeds:     embeds,
	})
	if err != nil {
		log.WithFields(log.Fields{"channelID": channelID}).Error("Unable to send a message")
		return nil
	}

	log.WithFields(log.Fields{"channel": channelID, "messageID": message.ID}).Trace("sent message")
	return message
}

// EditMessage edits the current message in a channel.
func EditMessage(s *discordgo.Session, channelID string, messageID string, msg string, components []discordgo.MessageComponent, embeds []*discordgo.MessageEmbed) error {
	log.Trace("--> EditMessage")
	defer log.Trace("<-- EditMessage")

	_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         messageID,
		Channel:    channelID,
		Content:    &msg,
		Components: &components,
		Embeds:     &embeds,
	})
	if err != nil {
		log.WithFields(log.Fields{"channelID": channelID, "messageID": messageID, "error": err}).Error("unable to edit a message")
		return err
	}

	log.WithFields(log.Fields{"channel": channelID, "messageID": messageID}).Trace("edited message")
	return nil
}
