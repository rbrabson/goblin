package heist

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/olekukonko/tablewriter"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/bank"
	"github.com/rbrabson/goblin/guild"
	"github.com/rbrabson/goblin/internal/channel"
	"github.com/rbrabson/goblin/internal/discmsg"
	"github.com/rbrabson/goblin/internal/format"
	"github.com/rbrabson/goblin/internal/unicode"
	log "github.com/sirupsen/logrus"
)

// componentHandlers are the buttons that appear on messages sent by this bot.
var (
	componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"join_heist": joinHeist,
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"heist":       heist,
		"heist-admin": heistAdmin,
	}

	adminCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "heist-admin",
			Description: "Heist admin commands.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "clear",
					Description: "Clears the criminal settings for the user.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "id",
							Description: "ID of the player to clear",
							Required:    true,
						},
					},
				},
				{
					Name:        "config",
					Description: "Configures the Heist bot.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "info",
							Description: "Returns the configuration information for the server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
						{
							Name:        "bail",
							Description: "Sets the base cost of bail.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "amount",
									Description: "The base cost of bail.",
									Required:    true,
								},
							},
						},
						{
							Name:        "cost",
							Description: "Sets the cost to plan or join a heist.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "amount",
									Description: "The cost to plan or join a heist.",
									Required:    true,
								},
							},
						},
						{
							Name:        "death",
							Description: "Sets how long players remain dead.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "time",
									Description: "The time the player remains dead, in seconds.",
									Required:    true,
								},
							},
						},
						{
							Name:        "patrol",
							Description: "Sets the time the authorities will prevent a new heist.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "time",
									Description: "The time the authorities will patrol, in seconds.",
									Required:    true,
								},
							},
						},
						{
							Name:        "sentence",
							Description: "Sets the base apprehension time when caught.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "time",
									Description: "The base time, in seconds.",
									Required:    true,
								},
							},
						},
						{
							Name:        "wait",
							Description: "Sets how long players can gather others for a heist.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "time",
									Description: "The time to wait for players to join the heist, in seconds.",
									Required:    true,
								},
							},
						},
					},
				},
				{
					Name:        "theme",
					Description: "Commands that interact with the heist themes.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "list",
							Description: "Gets the list of available heist themes.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
						{
							Name:        "set",
							Description: "Sets the current heist theme.",
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "name",
									Description: "Name of the theme to set.",
									Required:    true,
								},
							},
							Type: discordgo.ApplicationCommandOptionSubCommand,
						},
					},
				},
				{
					Name:        "reset",
					Description: "Resets a new heist that is hung.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "vault-reset",
					Description: "Resets the vaults to their maximum value.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}

	memberCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "heist",
			Description: "Heist game commands.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "bail",
					Description: "Bail a player out of jail.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "id",
							Description: "ID of the player to bail. Defaults to you.",
							Required:    false,
						},
					},
				},
				{
					Name:        "stats",
					Description: "Shows a user's stats.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "start",
					Description: "Plans a new heist.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "targets",
					Description: "Gets the list of available heist targets.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}
)

// config routes the configuration commands to the proper handlers.
func config(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> heist.config")
	defer log.Trace("<-- heist.config")

	options := i.ApplicationCommandData().Options[0].Options
	switch options[0].Name {
	case "cost":
		configCost(s, i)
	case "sentence":
		configSentence(s, i)
	case "patrol":
		configPatrol(s, i)
	case "bail":
		configBail(s, i)
	case "death":
		configDeath(s, i)
	case "wait":
		configWait(s, i)
	case "info":
		configInfo(s, i)
	}
}

// heistAdmin routes the commands to the subcommand and subcommandgroup handlers
func heistAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> heist.heistAmin")
	defer log.Trace("<-- heist.heistAdmin")

	p := discmsg.GetPrinter(language.AmericanEnglish)

	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		resp := p.Sprintf("You do not have permission to use this command.")
		discmsg.SendEphemeralResponse(s, i, resp)
		return
	}

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "clear":
		clearMember(s, i)
	case "config":
		config(s, i)
	case "reset":
		resetHeist(s, i)
	case "vault-reset":
		resetVaults(s, i)
	case "theme":
		theme(s, i)
	}
}

// heist routes the commands to the subcommand and subcommandgroup handlers
func heist(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> heist")
	defer log.Trace("<-- heist")

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "bail":
		bailoutPlayer(s, i)
	case "start":
		planHeist(s, i)
	case "stats":
		playerStats(s, i)
	case "targets":
		listTargets(s, i)
	}
}

// theme routes the theme commands to the proper handlers.
func theme(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> heist.theme")
	defer log.Trace("<-- heist.theme")

	options := i.ApplicationCommandData().Options[0].Options
	switch options[0].Name {
	case "list":
		listThemes(s, i)
	case "set":
		setTheme(s, i)
	}
}

// planHeist plans a new heist
func planHeist(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> heist.planHeist")
	defer log.Trace("<-- heist.planHeist")

	// Create a new heist
	heist, err := NewHeist(i.GuildID, i.Member.User.ID)
	if err != nil {
		log.WithField("error", err).Error("unable to create the heist")
		discmsg.SendEphemeralResponse(s, i, unicode.FirstToUpper(err.Error()))
		return
	}

	theme := GetTheme(i.GuildID)
	discmsg.SendResponse(s, i, "Starting a "+theme.Heist+"...")
	heist.Organizer.guildMember.SetName(i.Member.User.Username, i.Member.DisplayName())
	heist.interaction = i

	// The organizer has to pay a fee to plan the heist.
	account := bank.GetAccount(i.GuildID, i.Member.User.ID)
	account.Withdraw(heist.config.HeistCost)

	heistMessage(s, i, heist, heist.Organizer, "plan")

	waitForHeistToStart(s, i, heist)

	if len(heist.Crew) < 2 {
		heistMessage(s, i, heist, heist.Organizer, "cancel")
		p := discmsg.GetPrinter(language.AmericanEnglish)
		msg := p.Sprintf("The %s was cancelled due to lack of interest.", heist.theme.Heist)
		s.ChannelMessageSend(i.ChannelID, msg)
		log.WithFields(log.Fields{"guild": heist.GuildID, "heist": heist.theme.Heist}).Info("Heist cancelled due to lack of interest")
		heistLock.Lock()
		defer heistLock.Unlock()
		delete(currentHeists, heist.GuildID)
		return
	}

	defer heist.End()
	mute := channel.NewChannelMute(s, i)
	mute.MuteChannel()
	defer mute.UnmuteChannel()

	err = heistMessage(s, i, heist, heist.Organizer, "start")
	if err != nil {
		log.WithField("error", err).Error("Unable to mark the heist message as started")
	}

	res, err := heist.Start()
	if err != nil {
		log.WithField("error", err).Error("unable to start the heist")
		discmsg.SendEphemeralResponse(s, i, unicode.FirstToUpper(err.Error()))
		return
	}

	log.Debug("heist is starting")
	p := discmsg.GetPrinter(language.AmericanEnglish)
	msg := p.Sprintf("Get ready! The %s is starting with %d members.", heist.theme.Heist, len(heist.Crew))
	s.ChannelMessageSend(i.ChannelID, msg)

	time.Sleep(3 * time.Second)
	heistMessage(s, i, heist, heist.Organizer, "start")

	sendHeistResults(s, i, res)

	res.Target.StealFromValut(res.TotalStolen)
}

// waitForHeistToStart waits until the planning stage for the heist expires.
func waitForHeistToStart(s *discordgo.Session, i *discordgo.InteractionCreate, heist *Heist) {
	log.Trace("--> heist.waitHeist")
	defer log.Trace("<-- hesit.waitHeist")

	// Wait for the heist to be ready to start
	waitTime := heist.StartTime.Add(heist.config.WaitTime)
	log.WithFields(log.Fields{"guild": heist.GuildID, "waitTime": waitTime, "configWaitTime": heist.config.WaitTime, "currentTime": time.Now()}).Debug("wait for heist to start")
	for !time.Now().After(waitTime) {
		maximumWait := time.Until(waitTime)
		timeToWait := min(maximumWait, time.Duration(5*time.Second))
		if timeToWait < 0 {
			log.WithFields(log.Fields{"guild": heist.GuildID, "maximumWait": maximumWait, "timeToWait": timeToWait}).Debug("wait for the heist to start is over")
			break
		}
		time.Sleep(timeToWait)
		log.WithFields(log.Fields{"guild": heist.GuildID, "startTiime": heist.StartTime, "until": time.Until(heist.StartTime.Add(heist.config.WaitTime))}).Trace("waiting for the heist to start")
		heistMessage(s, i, heist, heist.Organizer, "update")
	}
}

// sendHeistResults sends the results of the heist to the channel
func sendHeistResults(s *discordgo.Session, i *discordgo.InteractionCreate, res *HeistResult) {
	log.Trace("--> heist.sendHeistResults")
	defer log.Trace("<-- heist.sendHeistResults")

	p := discmsg.GetPrinter(language.AmericanEnglish)
	theme := GetTheme(i.GuildID)

	log.Debug("Hitting " + res.Target.Name)
	msg := p.Sprintf("The %s has decided to hit **%s**.", theme.Crew, res.Target.Name)
	s.ChannelMessageSend(i.ChannelID, msg)
	time.Sleep(3 * time.Second)

	// Process the results
	for _, result := range res.AllResults {
		guildMember := result.Player.guildMember
		msg = p.Sprintf(result.Message+"\n", "**"+guildMember.Name+"**")
		if result.Status == APPREHENDED {
			msg += p.Sprintf("`%s dropped out of the game.`", guildMember.Name)
		}
		s.ChannelMessageSend(i.ChannelID, msg)
		time.Sleep(3 * time.Second)
	}

	if len(res.Escaped) == 0 {
		msg = "\nNo one made it out safe."
		s.ChannelMessageSend(i.ChannelID, msg)
	} else {
		msg = "\nThe raid is now over. Distributing player spoils."
		s.ChannelMessageSend(i.ChannelID, msg)
		// Render the results into a table and returnt he results.
		var tableBuffer strings.Builder
		table := tablewriter.NewWriter(&tableBuffer)
		table.SetBorder(false)
		table.SetAutoWrapText(false)
		table.SetAutoFormatHeaders(true)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetHeaderLine(false)
		table.SetBorder(false)
		table.SetTablePadding("\t")
		table.SetNoWhiteSpace(true)
		table.SetHeader([]string{"Player", "Loot", "Bonus", "Total"})
		for _, result := range res.AllResults {
			guildMember := result.Player.guildMember
			if result.Status == FREE || result.Status == APPREHENDED {
				data := []string{guildMember.Name, p.Sprintf("%d", result.StolenCredits), p.Sprintf("%d", result.BonusCredits), p.Sprintf("%d", result.StolenCredits+result.BonusCredits)}
				table.Append(data)
			}

		}
		table.Render()
		s.ChannelMessageSend(i.ChannelID, "```\n"+tableBuffer.String()+"```")
	}

	// Update the status for each player and then save the information
	for _, result := range res.AllResults {
		result.Player.heist = result.heist
		switch result.Status {
		case APPREHENDED:
			result.Player.Apprehended()
		case DEAD:
			result.Player.Died()
		default:
			result.Player.Escaped()
		}

		if len(res.Escaped) > 0 && result.StolenCredits != 0 {
			account := bank.GetAccount(i.GuildID, result.Player.MemberID)
			account.Deposit(result.StolenCredits + result.BonusCredits)
			log.WithFields(log.Fields{"Member": account.MemberID, "Stolen": result.StolenCredits, "Bonus": result.BonusCredits}).Debug("heist Loot")
		}
	}

	heistLock.Lock()
	h := currentHeists[i.GuildID]
	alertTimes[i.GuildID] = time.Now().Add(h.config.PoliceAlert)
	heistLock.Unlock()
	heistMessage(s, i, h, h.Organizer, "ended")
}

// joinHeist attempts to join a heist that is being planned
func joinHeist(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> joinHeist")
	defer log.Trace("<-- joinHeist")

	heist := currentHeists[i.GuildID]
	if heist == nil {
		theme := GetTheme(i.GuildID)
		discmsg.SendEphemeralResponse(s, i, fmt.Sprintf("No %s is planned", theme.Heist))
		return
	}

	heistMember := getHeistMember(i.GuildID, i.Member.User.ID)
	heistMember.guildMember.SetName(i.Member.User.Username, i.Member.DisplayName())
	err := heist.AddCrewMember(heistMember)
	if err != nil {
		discmsg.SendEphemeralResponse(s, i, unicode.FirstToUpper(err.Error()))
		return
	}

	// Withdraw the cost of the heist from the player's account. We know the player already
	// has the required number of credits as this is verified when adding them to the heist.
	account := bank.GetAccount(i.GuildID, heistMember.MemberID)
	account.Withdraw(heist.config.HeistCost)

	p := discmsg.GetPrinter(language.AmericanEnglish)
	resp := p.Sprintf("You have joined the %s at a cost of %d credits.", heist.theme.Heist, heist.config.HeistCost)
	discmsg.SendEphemeralResponse(s, i, resp)

	heistMessage(s, heist.interaction, heist, heistMember, "join")
}

// playerStats shows a player's heist stats
func playerStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> playerStats")
	defer log.Trace("<-- playerStats")

	theme := GetTheme(i.GuildID)
	player := getHeistMember(i.GuildID, i.Member.User.ID)
	player.guildMember.SetName(i.Member.User.Username, i.Member.DisplayName())
	caser := cases.Caser(cases.Title(language.Und, cases.NoLower))

	account := bank.GetAccount(i.GuildID, i.Member.User.ID)

	var sentence string
	if player.Status == APPREHENDED {
		if player.JailTimer.Before(time.Now()) {
			sentence = "Served"
		} else {
			timeRemaining := time.Until(player.JailTimer)
			sentence = format.Duration(timeRemaining)
		}
	} else {
		sentence = "None"
	}

	embeds := []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Title:       player.guildMember.Name,
			Description: player.CriminalLevel.String(),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Status",
					Value:  player.Status.String(),
					Inline: true,
				},
				{
					Name:   "Spree",
					Value:  fmt.Sprintf("%d", player.Spree),
					Inline: true,
				},
				{
					Name:   caser.String(theme.Bail),
					Value:  fmt.Sprintf("%d", player.BailCost),
					Inline: true,
				},
				{
					Name:   caser.String(theme.Sentence),
					Value:  sentence,
					Inline: true,
				},
				{
					Name:   "Apprehended",
					Value:  fmt.Sprintf("%d", player.JailCounter),
					Inline: true,
				},
				{
					Name:   "Total Deaths",
					Value:  fmt.Sprintf("%d", player.Deaths),
					Inline: true,
				},
				{
					Name:   "Lifetime Apprehensions",
					Value:  fmt.Sprintf("%d", player.TotalJail),
					Inline: true,
				},
				{
					Name:   "Credits",
					Value:  fmt.Sprintf("%d", account.CurrentBalance),
					Inline: true,
				},
			},
		},
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: embeds,
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Error("Unable to send the player stats to Discord, error:", err)
	}
}

// bailoutPlayer bails a player player out from jail. This defaults to the player initiating the command, but can
// be another player as well.
func bailoutPlayer(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> bailoutPlayer")
	log.Trace("<-- bailoutPlayer")

	var playerID string
	options := i.ApplicationCommandData().Options[0].Options
	for _, option := range options {
		if option.Name == "id" {
			playerID = strings.TrimSpace(option.StringValue())
		}
	}

	initiatingHeistMember := getHeistMember(i.GuildID, i.Member.User.ID)
	initiatingHeistMember.guildMember.SetName(i.Member.User.Username, i.Member.DisplayName())
	account := bank.GetAccount(i.GuildID, i.Member.User.ID)

	discmsg.SendEphemeralResponse(s, i, "Bailing "+playerID+"...")
	var heistMember *HeistMember
	if playerID != "" {
		heistMember = getHeistMember(i.GuildID, i.Member.User.ID)
	} else {
		heistMember = initiatingHeistMember
	}

	if heistMember.Status != APPREHENDED && heistMember.Status != OOB {
		var msg string
		if heistMember.MemberID == i.Member.User.ID {
			msg = "You are not in jail"
		} else {
			msg = fmt.Sprintf("%s is not in jail", initiatingHeistMember.guildMember.Name)
		}
		discmsg.EditResponse(s, i, msg)
		return
	}
	if heistMember.Status == APPREHENDED && heistMember.RemainingJailTime() <= 0 {
		discmsg.EditResponse(s, i, "You have already served your sentence.")
		heistMember.ClearJailAndDeathStatus()
		return
	}
	if account.CurrentBalance < heistMember.BailCost {
		msg := fmt.Sprintf("You do not have enough credits to play the bail of %d", heistMember.BailCost)
		discmsg.EditResponse(s, i, msg)
		return
	}

	account.Withdraw(heistMember.BailCost)
	heistMember.Status = OOB

	var msg string
	if heistMember.MemberID == initiatingHeistMember.MemberID {
		msg = fmt.Sprintf("Congratulations, you are now free! You spent %d credits on your bail. Enjoy your freedom while it lasts.", heistMember.BailCost)
		discmsg.EditResponse(s, i, msg)
	} else {
		member := guild.GetMember(heistMember.GuildID, heistMember.MemberID)
		initiatingMember := initiatingHeistMember.guildMember
		msg = fmt.Sprintf("Congratulations, %s, %s bailed you out by spending %d credits and now you are free!. Enjoy your freedom while it lasts.",
			member.Name,
			initiatingMember.Name,
			heistMember.BailCost,
		)
		discmsg.EditResponse(s, i, msg)
	}
}

// heistMessage sends the main command used to plan, join and leave a heist. It also handles the case where
// the heist starts, disabling the buttons to join/leave/cancel the heist.
func heistMessage(s *discordgo.Session, i *discordgo.InteractionCreate, heist *Heist, member *HeistMember, action string) error {
	log.Trace("--> heistMessage")
	defer log.Trace("<-- heistMessage")

	var status string
	var buttonDisabled bool
	switch action {
	case "plan", "join", "leave":
		until := time.Until(heist.StartTime.Add(heist.config.WaitTime))
		status = "Starts in " + format.Duration(until)
		buttonDisabled = false
	case "update":
		until := time.Until(heist.StartTime.Add(heist.config.WaitTime))
		status = "Starts in " + format.Duration(until)
		buttonDisabled = false
	case "start":
		status = "Started"
		buttonDisabled = true
	case "cancel":
		status = "Canceled"
		buttonDisabled = true
	default:
		status = "Ended"
		buttonDisabled = true
	}

	heist.mutex.Lock()
	crew := make([]string, 0, len(heist.Crew))
	for _, crewMember := range heist.Crew {
		crew = append(crew, crewMember.guildMember.Name)
	}
	heist.mutex.Unlock()

	caser := cases.Caser(cases.Title(language.Und, cases.NoLower))
	p := discmsg.GetPrinter(language.AmericanEnglish)
	msg := p.Sprintf("A new %s is being planned by %s. You can join the %s for a cost of %d credits at any time prior to the %s starting.",
		heist.theme.Heist,
		heist.Organizer.guildMember.Name,
		heist.theme.Heist,
		heist.config.HeistCost,
		heist.theme.Heist,
	)

	embeds := []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Title:       "Heist",
			Description: msg,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Status",
					Value:  status,
					Inline: true,
				},
				{
					Name:   fmt.Sprintf("%s (%d members)", caser.String(heist.theme.Crew), len(crew)),
					Value:  strings.Join(crew, ", "),
					Inline: true,
				},
			},
		},
	}
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Join",
				Style:    discordgo.SuccessButton,
				Disabled: buttonDisabled,
				CustomID: "join_heist",
				Emoji:    nil,
			},
		}},
	}
	emptymsg := ""
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &embeds,
		Components: &components,
		Content:    &emptymsg,
	})
	if err != nil {
		log.WithFields(log.Fields{"error": err, "guild": member.GuildID}).Error("unable to send the heist message")
		return err
	}

	return nil
}

/******** ADMIN COMMANDS ********/

// Reset resets the heist in case it hangs
func resetHeist(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> resetHeist")
	defer log.Trace("<-- resetHeist")

	mute := channel.NewChannelMute(s, i)
	defer mute.UnmuteChannel()

	heistLock.Lock()
	heist := currentHeists[i.GuildID]
	delete(currentHeists, i.GuildID)
	heistLock.Unlock()
	if heist == nil {
		theme := GetTheme(i.GuildID)
		msg := fmt.Sprintf("No %s is being planned; the channel was un-muted", theme.Heist)
		discmsg.SendEphemeralResponse(s, i, msg)
		return
	}

	heistMessage(s, i, heist, heist.Organizer, "cancel")

	discmsg.SendResponse(s, i, fmt.Sprintf("The %s has been reset", heist.theme.Heist))
}

// resetVaults sets the vaults within the guild to their maximum value.
func resetVaults(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> resetVaults")
	defer log.Trace("<-- resetVaults")

	ResetVaultsToMaximumValue(i.GuildID)
	discmsg.SendResponse(s, i, "Vaults have been reset to their maximum value")
}

// listTargets displays a list of available heist targets.
func listTargets(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> listTargets")
	defer log.Trace("<-- listTargets")

	theme := GetTheme(i.GuildID)
	targets := GetTargets(i.GuildID, theme.Name)

	if len(targets) == 0 {
		msg := "There aren't any targets!"
		discmsg.SendEphemeralResponse(s, i, msg)
		return
	}

	// Lets return the data in an Ascii table. Ideally, it would be using a Discord embed, but unfortunately
	// Discord only puts three columns per row, which isn't enough for our purposes.
	var tableBuffer strings.Builder
	table := tablewriter.NewWriter(&tableBuffer)
	table.SetBorder(false)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.SetHeader([]string{"ID", "Max Crew", theme.Vault, "Max " + theme.Vault, "Success Rate"})
	for _, target := range targets {
		data := []string{target.Name, fmt.Sprintf("%d", target.CrewSize), fmt.Sprintf("%d", target.Vault), fmt.Sprintf("%d", target.VaultMax), fmt.Sprintf("%.2f", target.Success)}
		table.Append(data)
	}
	table.Render()

	discmsg.SendEphemeralResponse(s, i, "```\n"+tableBuffer.String()+"\n```")
}

// clearMember clears the criminal state of the player.
func clearMember(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> clearMember")
	log.Trace("<-- clearMember")

	memberID := i.ApplicationCommandData().Options[0].Options[0].StringValue()
	member := guild.GetMember(i.GuildID, memberID)
	heistMember := getHeistMember(i.GuildID, memberID)
	heistMember.ClearJailAndDeathStatus()
	discmsg.SendResponse(s, i, fmt.Sprintf("Heist membber \"%s\"'s settings cleared", member.Name))
}

// listThemes returns the list of available themes that may be used for heists
func listThemes(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> listThemes")
	defer log.Trace("<-- listThemes")

	themes, err := GetThemeNames(i.GuildID)
	if err != nil {
		log.Warning("Unable to get the themes, error:", err)
	}

	embeds := []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Title:       "Available Themes",
			Description: "Available Themes for the Heist bot",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Themes",
					Value:  strings.Join(themes[:], ", "),
					Inline: true,
				},
			},
		},
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: embeds,
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Error("Unable to send list of themes to the user, error:", err)
	}
}

// setTheme sets the heist theme to the one specified in the command
func setTheme(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> setTheme")
	defer log.Trace("<-- setTheme")

	var themeName string
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	for _, option := range options {
		if option.Name == "name" {
			themeName = strings.TrimSpace(option.StringValue())
		}
	}

	config := GetConfig(i.GuildID)
	if themeName == config.Theme {
		discmsg.SendEphemeralResponse(s, i, "Theme `"+themeName+"` is already being used.")
		return
	}
	theme := GetTheme(i.GuildID)
	config.Theme = theme.Name
	log.Debug("Now using theme ", config.Theme)

	discmsg.SendResponse(s, i, "Theme "+themeName+" is now being used.")
}

// configCost sets the cost to plan or join a heist
func configCost(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> configCost")
	defer log.Trace("<-- configCost")

	config := GetConfig(i.GuildID)
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	cost := options[0].IntValue()
	config.HeistCost = int(cost)

	discmsg.SendResponse(s, i, fmt.Sprintf("Cost set to %d", cost))
	writeConfig(config)
}

// configSentence sets the base aprehension time when a player is apprehended.
func configSentence(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> configSentence")
	defer log.Trace("<-- configSentence")

	options := i.ApplicationCommandData().Options[0]
	if options == nil {
		discmsg.SendEphemeralResponse(s, i, "No sentence time provided (option missing)")
		return
	}
	options = options.Options[0]
	if options == nil {
		discmsg.SendEphemeralResponse(s, i, "No sentence time provided (2nd level option missing)")
		return
	}

	config := GetConfig(i.GuildID)
	if i.ApplicationCommandData().Options[0].Options[0].IntValue() == 0 {
		config.SentenceBase = 0
		discmsg.SendResponse(s, i, "Sentence disabled")
		writeConfig(config)
		return
	}
	sentence := i.ApplicationCommandData().Options[0].Options[0].IntValue()
	config.SentenceBase = time.Duration(sentence * int64(time.Second))

	discmsg.SendResponse(s, i, fmt.Sprintf("Sentence set to %d", sentence))

	writeConfig(config)
}

// configPatrol sets the time authorities will prevent a new heist following one being completed.
func configPatrol(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> configPatrol")
	defer log.Trace("<-- configPatrol")

	config := GetConfig(i.GuildID)
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	patrol := options[0].IntValue()
	config.PoliceAlert = time.Duration(patrol * int64(time.Second))

	discmsg.SendResponse(s, i, fmt.Sprintf("Patrol set to %d", patrol))

	writeConfig(config)
}

// configBail sets the base cost of bail.
func configBail(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> configBail")
	defer log.Trace("<-- configBail")

	config := GetConfig(i.GuildID)
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	bail := options[0].IntValue()
	config.BailBase = int(bail)

	discmsg.SendResponse(s, i, fmt.Sprintf("Bail set to %d", bail))

	writeConfig(config)
}

// configDeath sets how long players remain dead.
func configDeath(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> configDeath")
	defer log.Trace("<-- configDeath")

	config := GetConfig(i.GuildID)
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	death := options[0].IntValue()
	config.PoliceAlert = time.Duration(death * int64(time.Second))

	discmsg.SendResponse(s, i, fmt.Sprintf("Death set to %d", death))

	writeConfig(config)
}

// configWait sets how long players wait for others to join the heist.
func configWait(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> configWait")
	defer log.Trace("<-- configWait")

	config := GetConfig(i.GuildID)
	options := i.ApplicationCommandData().Options[0].Options[0].Options
	wait := options[0].IntValue()
	config.WaitTime = time.Duration(wait * int64(time.Second))

	discmsg.SendResponse(s, i, fmt.Sprintf("Wait set to %d", wait))

	writeConfig(config)
}

// configInfo returns the configuration for the Heist bot on this server.
func configInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> configInfo")
	defer log.Trace("<-- configInfo")

	config := GetConfig(i.GuildID)

	embed := &discordgo.MessageEmbed{
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "bail",
				Value:  fmt.Sprintf("%d", config.BailBase),
				Inline: true,
			},
			{
				Name:   "cost",
				Value:  fmt.Sprintf("%d", config.HeistCost),
				Inline: true,
			},
			{
				Name:   "death",
				Value:  fmt.Sprintf("%.f", config.DeathTimer.Seconds()),
				Inline: true,
			},
			{
				Name:   "patrol",
				Value:  fmt.Sprintf("%.f", config.PoliceAlert.Seconds()),
				Inline: true,
			},
			{
				Name:   "sentence",
				Value:  fmt.Sprintf("%.f", config.SentenceBase.Seconds()),
				Inline: true,
			},
			{
				Name:   "wait",
				Value:  fmt.Sprintf("%.f", config.WaitTime.Seconds()),
				Inline: true,
			},
		},
	}

	embeds := []*discordgo.MessageEmbed{
		embed,
	}
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Heist Configuration",
			Embeds:  embeds,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Error("Unable to send a response, error:", err)
	}
}
