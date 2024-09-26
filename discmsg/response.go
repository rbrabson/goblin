package discmsg

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

func EditResponse(s *discordgo.Session, i *discordgo.InteractionCreate, embeds []*discordgo.MessageEmbed, components []discordgo.MessageComponent) {
	log.Trace("--> EditResponse")
	defer log.Trace("<-- EditResponse")

	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &embeds,
		Components: &components,
	})
	if err != nil {
		log.Error("Unable to edit a response, error:", err)
	}

}

// SendResponse sends a response to a user interaction. The message can ephemeral or non-ephemeral,
// depending on whether the ephemeral boolean is set to `true`.
func SendResponse(s *discordgo.Session, i *discordgo.InteractionCreate, msg string, ephemeral ...bool) {
	log.Trace("--> SendResponse")
	defer log.Trace("<-- SendResponse")

	var err error
	if len(ephemeral) == 0 || !ephemeral[0] {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: msg,
			},
		})
	} else {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: msg,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
	if err != nil {
		log.Error("Unable to send a response, error:", err)
	}
}

// SendEphemeralResponse is a utility routine used to send an ephemeral response to a user's message or button press.
// It is shorthand for SendMessage(s, i, msg, true).
func SendEphemeralResponse(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	log.Trace("--> SendEphemeralResponse")
	defer log.Trace("<-- SendEphemeralResponse")

	SendResponse(s, i, msg, true)
}
