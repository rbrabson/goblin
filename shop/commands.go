package shop

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
	page "github.com/rbrabson/disgopage"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/guild"
	"github.com/rbrabson/goblin/internal/disctime"
	"github.com/rbrabson/goblin/internal/unicode"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	MAX_SHOP_ITEMS_DISPLAYED = 25
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"shop-admin": shopAdmin,
		"shop":       shop,
		"purchase":   initiatePurchase,
		"confirm":    completePurchase,
	}
)

var (
	adminCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "shop-admin",
			Description: "Commands used to interact with the shop for this server.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "add",
					Description: "Adds an item to the shop that may be purchased by a member.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "role",
							Description: "Adds a purchasable role to the shop.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "name",
									Description: "The name of the role that may be purchased.",
									Required:    true,
								},
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "cost",
									Description: "The cost of the role.",
									Required:    true,
								},
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "description",
									Description: "The description of the role that may be purchased.",
									Required:    false,
								},
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "duration",
									Description: "The duration of the role.",
									Required:    false,
								},
								{
									Type:        discordgo.ApplicationCommandOptionBoolean,
									Name:        "renewable",
									Description: "Whether the role is renewable.",
									Required:    false,
								},
							},
						},
						{
							Name:        "command",
							Description: "Adds a purchasable custom command to the shop.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "name",
									Description: "The name of the role that may be purchased.",
									Required:    true,
								},
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "cost",
									Description: "The cost of the role.",
									Required:    true,
								},
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "description",
									Description: "The description of the role that may be purchased.",
									Required:    false,
								},
							},
						},
					},
				},
				{
					Name:        "delete",
					Description: "Removes a purchasable item from the shop.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "role",
							Description: "Removes a purchasable role to the shop.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "name",
									Description: "The name of the role to be removed.",
									Required:    true,
								},
							},
						},
						{
							Name:        "command",
							Description: "Adds a purchasable custom command to the shop.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "name",
									Description: "The name of the role that may be purchased.",
									Required:    true,
								},
							},
						},
					},
				},
				{
					Name:        "ban",
					Description: "Bans a member from the shop.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "id",
							Description: "The ID of the member to ban from the shop.",
							Required:    true,
						},
					},
				},
				{
					Name:        "unban",
					Description: "Removes the ban of a member from the shop.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "id",
							Description: "The ID of the member to have the ban from the shop removed.",
							Required:    true,
						},
					},
				},
				{
					Name:        "channel",
					Description: "Sets the channel to which to pulish the shop items.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionChannel,
							Name:        "id",
							Description: "The ID of the channel to which to publish the shop items.",
							Required:    true,
						},
					},
				},
				{
					Name:        "mod-channel",
					Description: "Sets the channel to which to publish notices when an item is purchased or the purchase removed.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionChannel,
							Name:        "id",
							Description: "The ID of the channel to which to publish the shop items.",
							Required:    true,
						},
					},
				},
				{
					Name:        "notification-id",
					Description: "Sets the member to which to notify when a purchase requires additional action.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionChannel,
							Name:        "id",
							Description: "The ID of the member to notify when a purchase requires additional action.",
							Required:    true,
						},
					},
				},
				{
					Name:        "publish",
					Description: "Publishes the shop items in the shop channel.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}

	memberCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "shop",
			Description: "Commands used by a member to purchase items in the shop.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "purchases",
					Description: "Lists the items in the shop that may be purchased.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}
)

var (
	paginator *page.Paginator
)

// shopAdmin routes the shop admin commands to the proper handers.
func shopAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if status == discord.STOPPING || status == discord.STOPPED {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("The system is shutting down."),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You do not have permission to use this command."),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "add":
		addShopItem(s, i)
	case "delete":
		removeShopItem(s, i)
	case "update":
		updateShopItem(s, i)
	case "ban":
		banMember(s, i)
	case "unban":
		unbanMember(s, i)
	case "channel":
		setShopChannel(s, i)
	case "mod-channel":
		setShopModChannel(s, i)
	case "member-id":
		setNotificationID(s, i)
	case "publish":
		refreshShop(s, i)
	default:
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Command `%s` is not recognized.", options[0].Name)),
		)
		resp.SendEphemeral(s, i.Interaction)
	}
}

// addShopItem routes the add shop item commands to the proper handers.
func addShopItem(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	switch options[0].Options[0].Name {
	case "role":
		addRoleToShop(s, i)
	case "command":
		addCommandToShop(s, i)
	default:
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Command `%s\\%s` is not recognized.", options[0].Name, options[0].Options[0].Name)),
		)
		resp.SendEphemeral(s, i.Interaction)
	}

	config := GetConfig(i.GuildID)
	messageID, _ := publishShop(s, i.GuildID, config.ChannelID, config.MessageID)
	config.SetMessageID(messageID)
}

// addRoleToShop adds a role to the shop.
func addRoleToShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	// Get the options for the role to be added
	var roleName string
	var roleCost int
	var roleDesc string
	var roleDuration string
	var roleRenewable bool
	options := i.ApplicationCommandData().Options
	for _, option := range options[0].Options[0].Options {
		log.WithFields(log.Fields{"guildID": i.GuildID, "option": option}).Trace("processing option")
		switch option.Name {
		case "name":
			roleName = option.StringValue()
		case "cost":
			roleCost = int(option.IntValue())
		case "description":
			roleDesc = option.StringValue()
		case "duration":
			roleDuration = strings.ToUpper(option.StringValue())
			_, err := disctime.ParseDuration(roleDuration)
			if err != nil {
				log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "roleDuration": option.StringValue()}).Errorf("Failed to parse role duration: %s", err)
				resp := disgomsg.NewResponse(
					disgomsg.WithContent(fmt.Sprintf("Invalid duration: %s", err.Error())),
				)
				resp.SendEphemeral(s, i.Interaction)
				return
			}
		case "renewable":
			roleRenewable = option.BoolValue()
		}
	}
	if roleDesc == "" {
		roleDesc = roleName + " role"
	}

	// Verify the role can be added to the shop
	err := roleCreateChecks(s, i, roleName)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Errorf("failed to perform role create checks: %s", err)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
		)
		resp.SendEphemeral(s, i.Interaction)
	}

	// Add the role to the shop.
	shop := GetShop(i.GuildID)
	role := NewRole(i.GuildID, roleName, roleDesc, roleCost, roleDuration, roleRenewable)
	err = role.AddToShop(shop)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "roleDesc": roleDesc, "roleCost": roleCost, "roleDuration": roleDuration, "roleRenewable": roleRenewable}).Errorf("failed to add role to shop: %s", err)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Failed to add role `%s` to the shop: %s", roleName, err)),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	// Register the component handlers for the item
	shopItem := (*ShopItem)(role)
	registerShopItemComponentHandlers(shopItem)

	log.WithFields(log.Fields{"role": role}).Info("role added to shop")
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Role `%s` has been added to the shop.", roleName)),
	)
	resp.Send(s, i.Interaction)
}

// addCommandToShop adds a custom command to the shop.
func addCommandToShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Get the options for the custom command to be added
	var commandCost int
	var commandName string
	var commandDescription string
	options := i.ApplicationCommandData().Options
	for _, option := range options[0].Options[0].Options {
		log.WithFields(log.Fields{"guildID": i.GuildID, "option": option}).Trace("processing option")
		switch option.Name {
		case "cost":
			commandCost = int(option.IntValue())
		case "name":
			commandName = option.StringValue()
		case "description":
			commandDescription = option.StringValue()
		}
	}

	// Verify the custom command can be added to the shop
	err := customCommandCreateChecks(s, i, commandName)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID}).Errorf("failed to perform custom command create checks: %s", err)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
		)
		resp.SendEphemeral(s, i.Interaction)
	}

	// Add the custom command to the shop.
	shop := GetShop(i.GuildID)
	command := NewCustomCommand(i.GuildID, commandName, commandDescription, commandCost)
	err = command.AddToShop(shop)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "commandCost": commandCost}).Errorf("failed to add custom command to shop: %s", err)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Failed to add custom command to the shop: %s", err)),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	// Register the component handlers for the item
	shopItem := (*ShopItem)(command)
	registerShopItemComponentHandlers(shopItem)

	log.WithFields(log.Fields{"command": command}).Info("custom command added to shop")
	resp := disgomsg.NewResponse(
		disgomsg.WithContent("Custom command has been added to the shop."),
	)
	resp.Send(s, i.Interaction)
}

// removeShopItem routes the remove shop item commands to the proper handers.
func removeShopItem(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options
	switch options[0].Options[0].Name {
	case "role":
		removeRoleFromShop(s, i)
	case "command":
		removeCommandFromShop(s, i)
	default:
		msg := p.Sprint("Command `%s\\%s` is not recognized.", options[0].Name, options[0].Options[0].Name)
		log.Warn(msg)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(msg),
		)
		resp.SendEphemeral(s, i.Interaction)
	}
}

// removeRoleFromShop removes a role from the shop.
func removeRoleFromShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options

	// Get the role details
	roleName := options[0].Options[0].Options[0].StringValue()

	role := GetRole(i.GuildID, roleName)
	if role == nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Error("role not found in shop")
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(p.Sprintf("Role `%s` not found in the shop.", roleName)),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	shop := GetShop(i.GuildID)
	err := role.RemoveFromShop(shop)
	if err != nil {
		log.WithFields(log.Fields{"guildID": role.GuildID, "roleName": role.Name}).Errorf("failed to remove role from shop: %s", err)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(p.Sprintf("Failed to remove role `%s` from the shop: %s", roleName, err)),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}
	config := GetConfig(i.GuildID)
	publishShop(s, i.GuildID, config.ChannelID, config.MessageID)

	log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Info("role removed from shop")
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Role `%s` has been removed from the shop.", roleName)),
	)
	resp.Send(s, i.Interaction)
}

// removeCommandFromShop removes a custom command from the shop.
func removeCommandFromShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options

	// Get the role details
	commandName := options[0].Options[0].Options[0].StringValue()

	command := GetCustomCommand(i.GuildID, commandName)
	if command == nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "commandName": commandName}).Error("custom command not found in shop")
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(p.Sprintf("Custom command `%s` not found in the shop.", commandName)),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	shop := GetShop(i.GuildID)
	err := command.RemoveFromShop(shop)
	if err != nil {
		log.WithFields(log.Fields{"guildID": command.GuildID, "commandName": command.Name}).Errorf("failed to remove custom command from shop: %s", err)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(p.Sprintf("Failed to remove custom command `%s` from the shop: %s", commandName, err)),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}
	config := GetConfig(i.GuildID)
	publishShop(s, i.GuildID, config.ChannelID, config.MessageID)

	log.WithFields(log.Fields{"guildID": i.GuildID, "commandName": commandName}).Info("custom command removed from shop")
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Custom command `%s` has been removed from the shop.", commandName)),
	)
	resp.Send(s, i.Interaction)
}

// updateShopItem routes the update shop item commands to the proper handers.
func updateShopItem(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	switch options[0].Options[0].Name {
	case "role":
		updateRoleInShop(s, i)
	default:
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Command `%s\\%s` is not recognized.", options[0].Name, options[0].Options[0].Name)),
		)
		resp.SendEphemeral(s, i.Interaction)
	}
}

// updateRoleInShop updates a role in the shop.
func updateRoleInShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	// Get the roleOptions details
	roleOptions := options[0].Options[0]
	roleName := roleOptions.Options[0].StringValue()
	roleDesc := roleOptions.Options[1].StringValue()
	roleCost := int(roleOptions.Options[2].IntValue())
	var roleDuration string
	if len(roleOptions.Options) > 3 {
		roleDuration = roleOptions.Options[3].StringValue()
		_, err := disctime.ParseDuration(roleDuration)
		if err != nil {
			log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "roleDuration": roleDuration}).Errorf("Failed to parse role duration: %s", err)
			resp := disgomsg.NewResponse(
				disgomsg.WithContent(fmt.Sprintf("Invalid duration: %s", err.Error())),
			)
			resp.SendEphemeral(s, i.Interaction)
			return
		}
	}
	var roleRenewable bool
	if len(roleOptions.Options) > 4 {
		roleRenewable = roleOptions.Options[4].BoolValue()
	}

	role := GetRole(i.GuildID, roleName)
	if role == nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Errorf("failed to read role from shop")
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Role `%s` not found in the shop.", roleName)),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	err := role.Update(roleName, roleDesc, roleCost, roleDuration, roleRenewable)
	if err != nil {
		log.WithFields(log.Fields{"role": role}).Errorf("Failed to update role in shop: %s", err)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Failed to update role `%s` in the shop: %s", roleName, err)),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	log.WithFields(log.Fields{"role": role}).Info("Role updated in shop")
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(fmt.Sprintf("Role `%s` has been updated in the shop.", roleName)),
	)
	resp.Send(s, i.Interaction)
}

// banMember bans a member from the shop.
func banMember(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options
	memberID := options[0].Options[0].UserValue(s).ID

	member := GetMember(i.GuildID, memberID)
	err := member.AddRestriction(SHOP_BAN)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": memberID}).Errorf("failed to ban member from shop: %s", err)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Failed to ban member <@%s> from shop: %s", memberID, err)),
		)
		resp.SendEphemeral(s, i.Interaction)
	} else {
		log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": memberID}).Info("member banned from shop")
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(p.Sprintf("Member <@%s> has been banned from the shop.", memberID)),
		)
		resp.Send(s, i.Interaction)
	}
}

// unbanMember unbans a member from the shop.
func unbanMember(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options
	memberID := options[0].Options[0].UserValue(s).ID

	member, err := getMember(i.GuildID, memberID)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": memberID, "error": err}).Warn("shop member not found in the database")
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(p.Sprintf("Member <@%s> is not banned from the shop.", memberID)),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	err = member.RemoveRestriction(SHOP_BAN)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": memberID}).Errorf("failed to ban member from shop: %s", err)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(p.Sprintf("Failed to ban member <@%s> from shop: %s", memberID, err)),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": memberID}).Info("member banned from shop")
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("The ban from the shop for member <@%s> has been removed.", memberID)),
	)
	resp.Send(s, i.Interaction)
}

// setShopChannel sets the channel to which to publish the shop items.
func setShopChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)
	options := i.ApplicationCommandData().Options
	channelID := options[0].Options[0].ChannelValue(s).ID
	_, err := s.State.Channel(channelID)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "channelID": channelID}).Error("failed to get channel from state")
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Failed to get channel %s: %s", channelID, err)),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}
	config := GetConfig(i.GuildID)
	messageID, _ := publishShop(s, i.GuildID, channelID, config.MessageID)
	config.SetChannel(channelID)
	config.SetMessageID(messageID)

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Shop channel set to <#%s>", channelID)),
	)
	resp.Send(s, i.Interaction)
}

// setShopModChannel sets the channel to which to publish the shop items.
func setShopModChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	channelID := options[0].Options[0].ChannelValue(s).ID
	_, err := s.State.Channel(channelID)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "channelID": channelID}).Error("failed to get mod channel from state")
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Failed to get mod channel %s: %s", channelID, err)),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}
	config := GetConfig(i.GuildID)
	config.SetModChannel(channelID)

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(fmt.Sprintf("Shop mod channel set to <#%s>", channelID)),
	)
	resp.Send(s, i.Interaction)
}

// setNotificationID sets the ID of the member to nofify when a purchase requires additional actions.
func setNotificationID(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	memberID := options[0].Options[0].ChannelValue(s).ID
	_, err := s.State.Channel(memberID)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": memberID}).Error("failed to get member from state")
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Failed to get member %s: %s", memberID, err)),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}
	config := GetConfig(i.GuildID)
	config.SetNotificationID(memberID)

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(fmt.Sprintf("Shop notification ID set to <@%s>", memberID)),
	)
	resp.Send(s, i.Interaction)
}

// refreshShop refreshes the shop items in the shop channel.
func refreshShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config := GetConfig(i.GuildID)
	if config.ChannelID == "" {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("No shop channel set. Use `/shop-admin channel` to set the shop channel."),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}
	messageID, err := publishShop(s, i.GuildID, config.ChannelID, config.MessageID)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "channelID": config.ChannelID}).Error("failed to publish shop")
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Failed to publish shop: %s", err)),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}
	config.SetMessageID(messageID)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(fmt.Sprintf("Shop items refreshed and published to <#%s>", config.ChannelID)),
	)
	resp.SendEphemeral(s, i.Interaction)
	log.WithFields(log.Fields{"guildID": i.GuildID}).Info("shop refreshed")
}

// shop routes the shop commands to the proper handers.
func shop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if status == discord.STOPPING || status == discord.STOPPED {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("The system is shutting down."),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "purchases":
		listPurchasesFromShop(s, i)
	default:
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Command `%s\\%s` is not recognized.", options[0].Name, options[0].Options[0].Name)),
		)
		resp.SendEphemeral(s, i.Interaction)
	}
}

// listPurchasesFromShop lists the purchases made by the member in the shop.
func listPurchasesFromShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	purchases := GetAllPurchases(i.GuildID, i.Member.User.ID)

	if len(purchases) == 0 {
		log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": i.Member.User.ID}).Debug("no purchases found")
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You haven't made any purchases from the shop!"),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	embedFields := make([]*discordgo.MessageEmbedField, 0, len(purchases))

	for i, purchase := range purchases {
		sb := strings.Builder{}
		sb.WriteString(p.Sprintf("Description: %s\n", purchase.Item.Description))
		sb.WriteString(p.Sprintf("Price: %d\n", purchase.Item.Price))
		sb.WriteString(p.Sprintf("Purchased on: %s", purchase.PurchasedOn.Format("02 Jan 2006")))
		switch {
		case purchase.ExpiresOn.IsZero():
			// NO-OP
		case !purchase.HasExpired():
			sb.WriteString(p.Sprintf("\nExpires On: %s", purchase.ExpiresOn.Format("02 Jan 2006")))
			// sb.WriteString(p.Sprintf("Auto-Renew: %t\n", purchase.AutoRenew))
		default:
			sb.WriteString(p.Sprintf("\nExpired On: %s", purchase.ExpiresOn.Format("02 Jan 2006")))
		}
		if (i+1)%PURCHASES_PER_PAGE != 0 && (i+1) < len(purchases) {
			sb.WriteString("\n\u200B")
		}
		embedFields = append(embedFields, &discordgo.MessageEmbedField{
			Name:   p.Sprintf("%s %s", unicode.FirstToUpper(purchase.Item.Type), purchase.Item.Name),
			Value:  sb.String(),
			Inline: false,
		})
	}

	err := paginator.CreateInteractionResponse(s, i, "Purchases", embedFields, true)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": i.Member.User.ID, "error": err}).Error("unable to send shop purchases")
		return
	}

	log.WithFields(log.Fields{"guildID": i.GuildID, "numItems": len(purchases)}).Debug("shop purchases listed")
}

// initiatePurchase is used to buy an item from the shop using a button in the shop channel.
func initiatePurchase(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if status == discord.STOPPING || status == discord.STOPPED {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("The system is shutting down."),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	member, err := readMember(i.GuildID, i.Member.User.ID)
	if err == nil {
		if member.HasRestriction(SHOP_BAN) {
			resp := disgomsg.NewResponse(
				disgomsg.WithContent("You have been banned from the shop."),
			)
			resp.SendEphemeral(s, i.Interaction)
			return
		}
	}

	strs := strings.Split(i.Interaction.MessageComponentData().CustomID, ":")
	itemType := strs[1]
	itemName := strs[2]

	switch itemType {
	case ROLE:
		initiatePurchaseOfRoleFromShop(s, i, itemName)
	default:
		log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": i.Member.User.ID, "itemType": itemType, "itemName": itemName}).Error("unknown item type")
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Unknown item type `%s`", itemType)),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}
}

// initiatePurchaseOfRoleFromShop initiates the purchases of a role from the shop.
// The member will be prompted to confirm the purchase.
func initiatePurchaseOfRoleFromShop(s *discordgo.Session, i *discordgo.InteractionCreate, roleName string) {
	// Make sure the member can purchase the role
	err := rolePurchaseChecks(s, i, roleName)
	if err != nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	role := GetRole(i.GuildID, roleName)
	shopItem := (*ShopItem)(role)
	sendConfirmationMessage(s, i, shopItem)
	log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": i.Member.User.ID, "roleName": roleName}).Info("purchase of role initiated")
}

// completePurchase is used to finalize the purchase of an item from the shop.
// It is called when the member confirms the purchase using a "Buy" button.
func completePurchase(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if status == discord.STOPPING || status == discord.STOPPED {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("The system is shutting down."),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	strs := strings.Split(i.Interaction.MessageComponentData().CustomID, ":")
	itemType := strs[2]
	itemName := strs[3]

	switch itemType {
	case ROLE:
		completePurchaseOfRoleFromShop(s, i, itemName)
	default:
		log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": i.Member.User.ID, "itemType": itemType, "itemName": itemName}).Error("unknown item type")
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Unknown item type `%s`", itemType)),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}
}

// Complete the purchase of a role from the shop. This is called after the purchase has been confirmed by
// the member.
func completePurchaseOfRoleFromShop(s *discordgo.Session, i *discordgo.InteractionCreate, roleName string) {
	p := message.NewPrinter(language.AmericanEnglish)

	// Make sure the member can purchase the role
	err := rolePurchaseChecks(s, i, roleName)
	if err != nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	// Purchase the role
	role := GetRole(i.GuildID, roleName)
	purchase, err := role.Purchase(i.Member.User.ID, false)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "memberID": i.Member.User.ID, "error": err}).Errorf("failed to purchase role")
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	// Assign the role to the user. If the role can't be assigned, then undo the purchase of the role.
	err = guild.AssignRole(s, i.GuildID, i.Member.User.ID, roleName)
	if err != nil {
		purchase.Return()
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "memberID": i.Member.User.ID, "error": err}).Error("failed to assign role")
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(p.Sprintf("Failed to assign role `%s` to you: %s", roleName, err)),
		)
		resp.SendEphemeral(s, i.Interaction)
		return
	}

	log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": i.Member.User.ID, "roleName": roleName}).Info("role purchased")
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Role `%s` has been purchased and assigned to you.", roleName)),
	)
	resp.SendEphemeral(s, i.Interaction)

}

// sendConfirmationMessage sends a message to the member to confirm the purchase of an item from the shop.
func sendConfirmationMessage(s *discordgo.Session, i *discordgo.InteractionCreate, item *ShopItem) {
	p := message.NewPrinter(language.AmericanEnglish)

	shopItems := make([]*discordgo.MessageEmbedField, 0, 1)
	sb := strings.Builder{}
	sb.WriteString(p.Sprintf("Description: %s\n", item.Description))
	sb.WriteString(p.Sprintf("Cost: %d", item.Price))
	if item.Duration != "" {
		duration, _ := disctime.ParseDuration(item.Duration)
		sb.WriteString(p.Sprintf("\nDuration: %s\n", disctime.FormatDuration(duration)))
	}
	embed := &discordgo.MessageEmbedField{
		Name:   p.Sprintf("%s: %s", unicode.FirstToUpper(item.Type), item.Name),
		Value:  sb.String(),
		Inline: false,
	}
	shopItems = append(shopItems, embed)

	embeds := []*discordgo.MessageEmbed{
		{
			Title:  "Confirm the purchase of this item",
			Fields: shopItems,
		},
	}

	actionRow := []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "Buy",
			Style:    discordgo.SuccessButton,
			CustomID: "shop:buy:" + item.Type + ":" + item.Name,
		},
	}
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: actionRow,
		},
	}

	resp := disgomsg.NewResponse(
		disgomsg.WithComponents(components),
		disgomsg.WithEmbeds(embeds),
	)
	resp.SendEphemeral(s, i.Interaction)
}

// publishShop publishes the shop items to the channel.
func publishShop(s *discordgo.Session, guildID string, channelID string, messageID string) (string, error) {
	p := message.NewPrinter(language.AmericanEnglish)

	shop := GetShop(guildID)
	items := shop.Items

	// Create the message for the shop items
	shopItems := make([]*discordgo.MessageEmbedField, 0, len(items))
	sb := strings.Builder{}
	for _, item := range items {
		sb.Reset()
		sb.WriteString(p.Sprintf("Description: %s\n", item.Description))
		sb.WriteString(p.Sprintf("Cost: %d", item.Price))
		if item.Duration != "" {
			duration, err := disctime.ParseDuration(item.Duration)
			if err != nil {
				log.WithFields(log.Fields{"guildID": guildID, "itemName": item.Name, "itemDuration": item.Duration}).Errorf("failed to parse item duration: %s", err)
			}
			sb.WriteString(p.Sprintf("\nDuration: %s", disctime.FormatDuration(duration)))
			log.WithFields(log.Fields{"guildID": guildID, "itemName": item.Name, "itemDuration": item.Duration, "duration": duration, "formattedDuration": disctime.FormatDuration(duration)}).Info("item duration")
			// sb.WriteString(p.Sprintf("\nAuto-Rewable: %t", item.AutoRenewable))
		}
		if len(shopItems)+1 < len(items) && len(shopItems)+1 < MAX_SHOP_ITEMS_DISPLAYED {
			sb.WriteString("\n\u200B")
		}
		embed := &discordgo.MessageEmbedField{
			Name:   p.Sprintf("%s %s", unicode.FirstToUpper(item.Type), item.Name),
			Value:  sb.String(),
			Inline: false,
		}
		shopItems = append(shopItems, embed)

		if len(shopItems) == MAX_SHOP_ITEMS_DISPLAYED {
			log.WithFields(log.Fields{"guildID": guildID, "numItems": len(shopItems)}).Warn("maximum number of shop items reached")
			break
		}
	}
	embeds := []*discordgo.MessageEmbed{
		{
			Title:  "Shop Items",
			Fields: shopItems,
		},
	}

	// Build the buttons to use for purchasing shop items
	actionRows := getShopButtons(shop)
	components := make([]discordgo.MessageComponent, 0, len(actionRows))
	for _, row := range actionRows {
		components = append(components, row)
	}

	if messageID != "" {
		msg := disgomsg.NewMessage(
			disgomsg.WithComponents(components),
			disgomsg.WithEmbeds(embeds),
			disgomsg.WithChannelID(channelID),
		)
		err := msg.WithChannelID(channelID).WithMessageID(messageID).Edit(s)
		if err != nil {
			log.WithFields(log.Fields{"guildID": guildID, "channelID": channelID, "messageID": messageID}).Error("failed to edit shop items")
			messageID = ""
			config := GetConfig(guildID)
			config.MessageID = messageID
			writeConfig(config)
		}
		log.WithFields(log.Fields{"guildID": guildID, "channelID": channelID, "messageID": messageID}).Info("shop items updated")
	}
	if messageID == "" {
		msg := disgomsg.NewMessage(
			disgomsg.WithComponents(components),
			disgomsg.WithEmbeds(embeds),
		)
		msgID, err := msg.Send(s, channelID)
		if err != nil {
			log.WithFields(log.Fields{"guildID": guildID, "channelID": channelID}).Error("failed to publish shop items")
			return "", err
		}
		config := GetConfig(guildID)
		config.MessageID = msgID
		writeConfig(config)
	}

	log.WithFields(log.Fields{"guildID": guildID, "numItems": len(items)}).Info("shop items published")
	return messageID, nil
}

// getShopButtons returns the buttons for the shop items, which may be used to
// purchase items in the shop.
func getShopButtons(shop *Shop) []discordgo.ActionsRow {
	buttonsPerRow := 5
	rows := make([]discordgo.ActionsRow, 0, len(shop.Items)/buttonsPerRow)

	itemsIncludedInButtons := 0
	for len(shop.Items) > itemsIncludedInButtons {
		racersNotInButtons := len(shop.Items) - itemsIncludedInButtons
		buttonsForNextRow := min(buttonsPerRow, racersNotInButtons)
		buttons := make([]discordgo.MessageComponent, 0, buttonsForNextRow)
		for j := 0; j < buttonsForNextRow; j++ {
			index := j + itemsIncludedInButtons
			item := shop.Items[index]
			customID := "shop:" + item.Type + ":" + item.Name
			button := discordgo.Button{
				Label:    unicode.FirstToUpper(item.Type) + ": " + item.Name,
				Style:    discordgo.PrimaryButton,
				CustomID: customID,
				Emoji:    nil,
			}
			buttons = append(buttons, button)
		}
		itemsIncludedInButtons += buttonsForNextRow

		row := discordgo.ActionsRow{Components: buttons}
		rows = append(rows, row)
		if itemsIncludedInButtons >= MAX_SHOP_ITEMS_DISPLAYED {
			break
		}
	}

	return rows
}
