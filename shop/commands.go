package shop

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/disgomsg"
	page "github.com/rbrabson/disgopage"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/guild"
	"github.com/rbrabson/goblin/internal/disctime"
	"github.com/rbrabson/goblin/internal/unicode"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	MaxShopItemsDisplayed = 25
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
						// {
						// 	Name:        "command",
						// 	Description: "Adds a purchasable custom command to the shop.",
						// 	Type:        discordgo.ApplicationCommandOptionSubCommand,
						// 	Options: []*discordgo.ApplicationCommandOption{
						// 		{
						// 			Type:        discordgo.ApplicationCommandOptionString,
						// 			Name:        "name",
						// 			Description: "The name of the custom command.",
						// 			Required:    true,
						// 		},
						// 		{
						// 			Type:        discordgo.ApplicationCommandOptionInteger,
						// 			Name:        "cost",
						// 			Description: "The cost of the custom command.",
						// 			Required:    true,
						// 		},
						// 		{
						// 			Type:        discordgo.ApplicationCommandOptionString,
						// 			Name:        "description",
						// 			Description: "The description of the custom command.",
						// 			Required:    false,
						// 		},
						// 		{
						// 			Type:        discordgo.ApplicationCommandOptionInteger,
						// 			Name:        "max-purchases",
						// 			Description: "The maximum number of purchases of the custom command. Defaults to 1.",
						// 			Required:    false,
						// 		},
						// 	},
						// },
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
						// {
						// 	Name:        "command",
						// 	Description: "Adds a purchasable custom command to the shop.",
						// 	Type:        discordgo.ApplicationCommandOptionSubCommand,
						// 	Options: []*discordgo.ApplicationCommandOption{
						// 		{
						// 			Type:        discordgo.ApplicationCommandOptionString,
						// 			Name:        "name",
						// 			Description: "The name of the custom command.",
						// 			Required:    true,
						// 		},
						// 	},
						// },
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
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}

	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You do not have permission to use this command."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "add":
		addShopItem(s, i)
	case "delete":
		removeShopItem(s, i)
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
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
	}
}

// addShopItem routes the add shop item commands to the proper handers.
func addShopItem(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	switch options[0].Options[0].Name {
	case "role":
		addRoleToShop(s, i)
	case "command":
		addCustomCommandToShop(s, i)
	default:
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Command `%s\\%s` is not recognized.", options[0].Name, options[0].Options[0].Name)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
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
				slog.Error("Failed to parse role duration",
					slog.String("guildID", i.GuildID),
					slog.String("roleName", roleName),
					slog.String("roleDuration", roleDuration),
					slog.Any("error", err),
				)
				resp := disgomsg.NewResponse(
					disgomsg.WithContent(fmt.Sprintf("Invalid duration: %s", err.Error())),
				)
				if err := resp.SendEphemeral(s, i.Interaction); err != nil {
					slog.Error("unable to send ephemeral response", slog.Any("error", err))
				}
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
		slog.Error("failed to perform role create checks",
			slog.String("guildID", i.GuildID),
			slog.String("roleName", roleName),
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
	}

	// Add the role to the shop.
	shop := GetShop(i.GuildID)
	role := NewRole(i.GuildID, roleName, roleDesc, roleCost, roleDuration, roleRenewable)
	err = role.AddToShop(shop)
	if err != nil {
		slog.Error("failed to add role to shop",
			slog.String("guildID", i.GuildID),
			slog.String("roleName", roleName),
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Failed to add role `%s` to the shop: %s", roleName, err)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}

	// Register the component handlers for the item
	shopItem := (*ShopItem)(role)
	registerShopItemComponentHandlers(shopItem)

	slog.Info("role added to shop",
		slog.String("guildID", i.GuildID),
		slog.String("roleName", roleName),
	)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Role `%s` has been added to the shop.", roleName)),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("unable to send ephemeral response", slog.Any("error", err))
	}
}

// addCustomCommandToShop adds a custom command to the shop.
func addCustomCommandToShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Get the options for the custom command to be added
	var commandName string
	var commandCost int
	var commandDescription string
	options := i.ApplicationCommandData().Options
	for _, option := range options[0].Options[0].Options {
		switch option.Name {
		case "name":
			commandName = option.StringValue()
		case "cost":
			commandCost = int(option.IntValue())
		case "description":
			commandDescription = option.StringValue()
		}
	}

	// Verify the custom command can be added to the shop
	err := customCommandCreateChecks(i.GuildID, commandName)
	if err != nil {
		slog.Error("failed to perform custom command create checks",
			slog.String("guildID", i.GuildID),
			slog.String("commandName", commandName),
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
	}

	// Add the custom command to the shop.
	shop := GetShop(i.GuildID)
	command := NewCustomCommand(i.GuildID, commandName, commandDescription, commandCost)
	err = command.AddToShop(shop)
	if err != nil {
		slog.Error("failed to add custom command to shop",
			slog.String("guildID", i.GuildID),
			slog.String("commandName", commandName),
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Failed to add custom command `%s` to the shop: %s", commandName, err)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}

	// Register the component handlers for the item
	shopItem := (*ShopItem)(command)
	registerShopItemComponentHandlers(shopItem)

	slog.Info("custom command added to shop",
		slog.String("guildID", i.GuildID),
		slog.String("commandName", commandName),
	)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(fmt.Sprintf("Custom command `%s` has been added to the shop.", commandName)),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("unable to send ephemeral response", slog.Any("error", err))
	}
}

// removeShopItem routes the remove shop item commands to the proper handers.
func removeShopItem(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options
	switch options[0].Options[0].Name {
	case "role":
		removeRoleFromShop(s, i)
	case "command":
		removeCustomCommandFromShop(s, i)
	default:
		msg := p.Sprint("Command `%s\\%s` is not recognized.", options[0].Name, options[0].Options[0].Name)
		slog.Warn(msg)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(msg),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
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
		slog.Error("role not found in shop",
			slog.String("guildID", i.GuildID),
			slog.String("roleName", roleName),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(p.Sprintf("Role `%s` not found in the shop.", roleName)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}

	shop := GetShop(i.GuildID)
	err := role.RemoveFromShop(shop)
	if err != nil {
		slog.Error("failed to remove role from shop",
			slog.String("guildID", role.GuildID),
			slog.String("roleName", role.Name),
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(p.Sprintf("Failed to remove role `%s` from the shop: %s", roleName, err)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}
	config := GetConfig(i.GuildID)
	if _, err := publishShop(s, i.GuildID, config.ChannelID, config.MessageID); err != nil {
		slog.Error("failed to publish the shop", slog.Any("error", err))
	}

	slog.Info("role removed from shop",
		slog.String("guildID", i.GuildID),
		slog.String("roleName", roleName),
	)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Role `%s` has been removed from the shop.", roleName)),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("unable to send ephemeral response", slog.Any("error", err))
	}
}

// removeCustomCommandFromShop removes a custom command from the shop.
func removeCustomCommandFromShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options

	// Get the custom command details
	commandName := options[0].Options[0].Options[0].StringValue()

	command := GetCustomCommand(i.GuildID, commandName)
	if command == nil {
		slog.Error("custom command not found in shop",
			slog.String("guildID", i.GuildID),
			slog.String("commandName", commandName),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(p.Sprintf("Custom command `%s` not found in the shop.", commandName)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}

	shop := GetShop(i.GuildID)
	err := command.RemoveFromShop(shop)
	if err != nil {
		slog.Error("failed to remove custom command from shop",
			slog.String("guildID", command.GuildID),
			slog.String("commandName", command.Name),
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(p.Sprintf("Failed to remove custom command `%s` from the shop: %s", commandName, err)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}
	config := GetConfig(i.GuildID)
	if _, err := publishShop(s, i.GuildID, config.ChannelID, config.MessageID); err != nil {
		slog.Error("failed to publish the shop", slog.Any("error", err))
	}

	slog.Info("custom command removed from shop",
		slog.String("guildID", i.GuildID),
		slog.String("commandName", commandName),
	)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Custom command `%s` has been removed from the shop.", commandName)),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("unable to send ephemeral response", slog.Any("error", err))
	}
}

// banMember bans a member from the shop.
func banMember(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options
	memberID := options[0].Options[0].UserValue(s).ID

	member := GetMember(i.GuildID, memberID)
	err := member.AddRestriction(ShopBan)
	if err != nil {
		slog.Error("failed to ban member from shop",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", memberID),
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Failed to ban member <@%s> from shop: %s", memberID, err)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
	} else {
		slog.Info("member banned from shop",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", memberID),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(p.Sprintf("Member <@%s> has been banned from the shop.", memberID)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
	}
}

// unbanMember unbans a member from the shop.
func unbanMember(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options
	memberID := options[0].Options[0].UserValue(s).ID

	member, err := getMember(i.GuildID, memberID)
	if err != nil {
		slog.Warn("shop member not found in the database",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", memberID),
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(p.Sprintf("Member <@%s> is not banned from the shop.", memberID)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}

	err = member.RemoveRestriction(ShopBan)
	if err != nil {
		slog.Error("failed to ban member from shop",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", memberID),
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(p.Sprintf("Failed to ban member <@%s> from shop: %s", memberID, err)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}

	slog.Info("member banned from shop",
		slog.String("guildID", i.GuildID),
		slog.String("memberID", memberID),
	)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("The ban from the shop for member <@%s> has been removed.", memberID)),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("unable to send ephemeral response", slog.Any("error", err))
	}
}

// setShopChannel sets the channel to which to publish the shop items.
func setShopChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)
	options := i.ApplicationCommandData().Options
	channelID := options[0].Options[0].ChannelValue(s).ID
	_, err := s.State.Channel(channelID)
	if err != nil {
		slog.Error("failed to get channel from state",
			slog.String("guildID", i.GuildID),
			slog.String("channelID", channelID),
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Failed to get channel %s: %s", channelID, err)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}
	config := GetConfig(i.GuildID)
	messageID, _ := publishShop(s, i.GuildID, channelID, config.MessageID)
	config.SetChannel(channelID)
	config.SetMessageID(messageID)

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Shop channel set to <#%s>", channelID)),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("unable to send ephemeral response", slog.Any("error", err))
	}
}

// setShopModChannel sets the channel to which to publish the shop items.
func setShopModChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	channelID := options[0].Options[0].ChannelValue(s).ID
	_, err := s.State.Channel(channelID)
	if err != nil {
		slog.Error("failed to get mod channel from state",
			slog.String("guildID", i.GuildID),
			slog.String("channelID", channelID),
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Failed to get mod channel %s: %s", channelID, err)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}
	config := GetConfig(i.GuildID)
	config.SetModChannel(channelID)

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(fmt.Sprintf("Shop mod channel set to <#%s>", channelID)),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("unable to send ephemeral response", slog.Any("error", err))
	}
}

// setNotificationID sets the ID of the member to nofify when a purchase requires additional actions.
func setNotificationID(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	memberID := options[0].Options[0].ChannelValue(s).ID
	_, err := s.State.Channel(memberID)
	if err != nil {
		slog.Error("failed to get member from state",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", memberID),
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Failed to get member %s: %s", memberID, err)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}
	config := GetConfig(i.GuildID)
	config.SetNotificationID(memberID)

	resp := disgomsg.NewResponse(
		disgomsg.WithContent(fmt.Sprintf("Shop notification ID set to <@%s>", memberID)),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("unable to send ephemeral response", slog.Any("error", err))
	}
}

// refreshShop refreshes the shop items in the shop channel.
func refreshShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config := GetConfig(i.GuildID)
	if config.ChannelID == "" {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("No shop channel set. Use `/shop-admin channel` to set the shop channel."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}
	messageID, err := publishShop(s, i.GuildID, config.ChannelID, config.MessageID)
	if err != nil {
		slog.Error("failed to publish shop",
			slog.String("guildID", i.GuildID),
			slog.String("channelID", config.ChannelID),
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Failed to publish shop: %s", err)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}
	config.SetMessageID(messageID)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(fmt.Sprintf("Shop items refreshed and published to <#%s>", config.ChannelID)),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("unable to send ephemeral response", slog.Any("error", err))
	}
	slog.Info("shop refreshed",
		slog.String("guildID", i.GuildID),
	)
}

// shop routes the shop commands to the proper handers.
func shop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if status == discord.STOPPING || status == discord.STOPPED {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("The system is shutting down."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
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
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
	}
}

// listPurchasesFromShop lists the purchases made by the member in the shop.
func listPurchasesFromShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p := message.NewPrinter(language.AmericanEnglish)

	purchases := GetAllPurchases(i.GuildID, i.Member.User.ID)

	if len(purchases) == 0 {
		slog.Debug("no purchases found",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("You haven't made any purchases from the shop!"),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
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
		if (i+1)%PurchasesPerPage != 0 && (i+1) < len(purchases) {
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
		slog.Error("unable to send shop purchases",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.Any("error", err),
		)
		return
	}

	slog.Debug("shop purchases listed",
		slog.String("guildID", i.GuildID),
		slog.String("memberID", i.Member.User.ID),
	)
}

// initiatePurchase is used to buy an item from the shop using a button in the shop channel.
func initiatePurchase(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if status == discord.STOPPING || status == discord.STOPPED {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("The system is shutting down."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}

	member, err := readMember(i.GuildID, i.Member.User.ID)
	if err == nil {
		if member.HasRestriction(ShopBan) {
			resp := disgomsg.NewResponse(
				disgomsg.WithContent("You aren't able to purchase items from the shop."),
			)
			if err := resp.SendEphemeral(s, i.Interaction); err != nil {
				slog.Error("unable to send ephemeral response", slog.Any("error", err))
			}
			return
		}
	}

	strs := strings.Split(i.Interaction.MessageComponentData().CustomID, ":")
	itemType := strs[1]
	itemName := strs[2]

	switch itemType {
	case ROLE:
		initiatePurchaseOfRoleFromShop(s, i, itemName)
	case CustomCommandCollection:
		initiatePurchaseOfCustomCommandFromShop(s, i, itemName)
	default:
		slog.Error("unknown item type",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.String("itemType", itemType),
			slog.String("itemName", itemName),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Unknown item type `%s`", itemType)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
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
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}

	role := GetRole(i.GuildID, roleName)
	shopItem := (*ShopItem)(role)
	sendConfirmationMessage(s, i, shopItem)
	slog.Info("purchase of role initiated",
		slog.String("guildID", i.GuildID),
		slog.String("memberID", i.Member.User.ID),
		slog.String("roleName", roleName),
	)
}

// initiatePurchaseOfCustomCommandFromShop initiates the purchases of a custom command from the shop.
// The member will be prompted to confirm the purchase.
func initiatePurchaseOfCustomCommandFromShop(s *discordgo.Session, i *discordgo.InteractionCreate, commandName string) {
	// Make sure the member can purchase the role
	err := customCommandPurchaseChecks(i.GuildID, i.Member.User.ID, commandName)
	if err != nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}

	command := GetCustomCommand(i.GuildID, commandName)
	shopItem := (*ShopItem)(command)
	sendConfirmationMessage(s, i, shopItem)
	slog.Info("purchase of custom command initiated",
		slog.String("guildID", i.GuildID),
		slog.String("memberID", i.Member.User.ID),
		slog.String("commandName", commandName),
	)
}

// completePurchase is used to finalize the purchase of an item from the shop.
// It is called when the member confirms the purchase using a "Buy" button.
func completePurchase(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if status == discord.STOPPING || status == discord.STOPPED {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent("The system is shutting down."),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}

	strs := strings.Split(i.Interaction.MessageComponentData().CustomID, ":")
	itemType := strs[2]
	itemName := strs[3]

	switch itemType {
	case ROLE:
		completePurchaseOfRoleFromShop(s, i, itemName)
	case CustomCommandCollection:
		completePurchaseOfCustomCommandFromShop(s, i, itemName)
	default:
		slog.Error("unknown item type",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.String("itemType", itemType),
			slog.String("itemName", itemName),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(fmt.Sprintf("Unknown item type `%s`", itemType)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
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
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}

	// Purchase the role
	role := GetRole(i.GuildID, roleName)
	purchase, err := role.Purchase(i.Member.User.ID, false)
	if err != nil {
		slog.Error("failed to purchase role",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.String("roleName", roleName),
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}

	// Assign the role to the user. If the role can't be assigned, then undo the purchase of the role.
	err = guild.AssignRole(s, i.GuildID, i.Member.User.ID, roleName)
	if err != nil {
		if err := purchase.Return(); err != nil {
			slog.Error("failed to purchase role")
		}
		slog.Error("failed to assign role",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.String("roleName", roleName),
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(p.Sprintf("Failed to assign role `%s` to you: %s", roleName, err)),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}

	slog.Info("role purchased",
		slog.String("guildID", i.GuildID),
		slog.String("memberID", i.Member.User.ID),
		slog.String("roleName", roleName),
	)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Role `%s` has been purchased and assigned to you.", roleName)),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("unable to send ephemeral response", slog.Any("error", err))
	}
}

// Complete the purchase of a custom command from the shop. This is called after the purchase has been confirmed by
// the member.
func completePurchaseOfCustomCommandFromShop(s *discordgo.Session, i *discordgo.InteractionCreate, commandName string) {
	p := message.NewPrinter(language.AmericanEnglish)

	// Make sure the member can purchase the role
	err := customCommandPurchaseChecks(i.GuildID, i.Member.User.ID, commandName)
	if err != nil {
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}

	// Purchase the custom command
	command := GetCustomCommand(i.GuildID, commandName)
	_, err = command.Purchase(s, i.Member.User.ID)
	if err != nil {
		slog.Error("failed to purchase custom command",
			slog.String("guildID", i.GuildID),
			slog.String("memberID", i.Member.User.ID),
			slog.String("commandName", commandName),
			slog.Any("error", err),
		)
		resp := disgomsg.NewResponse(
			disgomsg.WithContent(unicode.FirstToUpper(err.Error())),
		)
		if err := resp.SendEphemeral(s, i.Interaction); err != nil {
			slog.Error("unable to send ephemeral response", slog.Any("error", err))
		}
		return
	}

	slog.Info("role purchased",
		slog.String("guildID", i.GuildID),
		slog.String("memberID", i.Member.User.ID),
		slog.String("commandName", commandName),
	)
	resp := disgomsg.NewResponse(
		disgomsg.WithContent(p.Sprintf("Custom command `%s` has been purchased.", commandName)),
	)
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("unable to send ephemeral response", slog.Any("error", err))
	}
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
	if err := resp.SendEphemeral(s, i.Interaction); err != nil {
		slog.Error("unable to send ephemeral response", slog.Any("error", err))
	}
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
				slog.Error("failed to parse item duration",
					slog.String("guildID", guildID),
					slog.String("itemName", item.Name),
					slog.String("itemDuration", item.Duration),
					slog.Any("error", err),
				)
			}
			sb.WriteString(p.Sprintf("\nDuration: %s", disctime.FormatDuration(duration)))
			slog.Info("item duration",
				slog.String("guildID", guildID),
				slog.String("itemName", item.Name),
				slog.String("itemDuration", item.Duration),
				slog.Any("duration", duration),
				slog.String("formattedDuration", disctime.FormatDuration(duration)),
			)
		}
		if len(shopItems)+1 < len(items) && len(shopItems)+1 < MaxShopItemsDisplayed {
			sb.WriteString("\n\u200B")
		}
		embed := &discordgo.MessageEmbedField{
			Name:   p.Sprintf("%s %s", unicode.FirstToUpper(item.Type), item.Name),
			Value:  sb.String(),
			Inline: false,
		}
		shopItems = append(shopItems, embed)

		if len(shopItems) == MaxShopItemsDisplayed {
			slog.Warn("maximum number of shop items reached",
				slog.String("guildID", guildID),
				slog.Int("numItems", len(shopItems)),
			)
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
			slog.Error("failed to edit shop items",
				slog.String("guildID", guildID),
				slog.String("channelID", channelID),
				slog.String("messageID", messageID),
				slog.Any("error", err),
			)
			messageID = ""
			config := GetConfig(guildID)
			config.MessageID = messageID
			if err := writeConfig(config); err != nil {
				slog.Error("failed to write config file",
					slog.Any("error", err),
				)
			}
		}
		slog.Info("shop items updated",
			slog.String("guildID", guildID),
			slog.String("channelID", channelID),
			slog.String("messageID", messageID),
			slog.Int("numItems", len(items)),
		)
	}
	if messageID == "" {
		msg := disgomsg.NewMessage(
			disgomsg.WithComponents(components),
			disgomsg.WithEmbeds(embeds),
		)
		msgID, err := msg.Send(s, channelID)
		if err != nil {
			slog.Error("failed to publish shop items",
				slog.String("guildID", guildID),
				slog.String("channelID", channelID),
				slog.Any("error", err),
			)
			return "", err
		}
		config := GetConfig(guildID)
		config.MessageID = msgID
		if err := writeConfig(config); err != nil {
			slog.Error("failed to write config file",
				slog.Any("error", err),
			)
		}
	}

	slog.Info("shop items published",
		slog.String("guildID", guildID),
		slog.String("channelID", channelID),
		slog.String("messageID", messageID),
		slog.Int("numItems", len(items)),
	)
	return messageID, nil
}

// getShopButtons returns the buttons for the shop items, which may be used to purchase items in the shop.
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
		if itemsIncludedInButtons >= MaxShopItemsDisplayed {
			break
		}
	}

	return rows
}
