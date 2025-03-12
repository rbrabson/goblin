package shop

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/guild"
	"github.com/rbrabson/goblin/internal/discmsg"
	"github.com/rbrabson/goblin/internal/disctime"
	"github.com/rbrabson/goblin/internal/unicode"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"shop-admin": shopAdmin,
		"shop":       shop,
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
			},
		},
	}

	memberCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "shop",
			Description: "Commands used by a member to purchase items in the shop.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "buy",
					Description: "Adds an item to the shop that may be purchased by a member.",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "role",
							Description: "Buys a custom role for the server.",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "name",
									Description: "The name of the role to be purchased.",
									Required:    true,
								},
							},
						},
					},
				},
				{
					Name:        "purchases",
					Description: "Lists the items in the shop that may be purchased.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "list",
					Description: "Lists the items in the shop.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}
)

// shopAdmin routes the shop admin commands to the proper handers.
func shopAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> shop.shopAdmin")
	defer log.Trace("<-- shop.shopAdmin")

	if status == discord.STOPPING || status == discord.STOPPED {
		discmsg.SendEphemeralResponse(s, i, "The system is currently shutting down.")
		return
	}

	p := discmsg.GetPrinter(language.AmericanEnglish)

	if !guild.IsAdmin(s, i.GuildID, i.Member.User.ID) {
		resp := p.Sprintf("You do not have permission to use this command.")
		discmsg.SendEphemeralResponse(s, i, resp)
		return
	}

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "add":
		addShopItem(s, i)
	case "remove":
		removeShopItem(s, i)
	case "update":
		updateShopItem(s, i)
	case "list":
		listShopItems(s, i)
	default:
		msg := p.Sprint("Command `%s` is not recognized.", options[0].Name)
		discmsg.SendEphemeralResponse(s, i, msg)
	}
}

// addShopItem routes the add shop item commands to the proper handers.
func addShopItem(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> shop.addShopItem")
	defer log.Trace("<-- shop.addShopItem")

	p := discmsg.GetPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options
	switch options[0].Options[0].Name {
	case "role":
		addRoleToShop(s, i)
	default:
		msg := p.Sprint("Command `%s\\%s` is not recognized.", options[0].Name, options[0].Options[0].Name)
		discmsg.SendEphemeralResponse(s, i, msg)
	}
}

// addRoleToShop adds a role to the shop.
func addRoleToShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> shop.addRoleToShop")
	defer log.Trace("<-- shop.addRoleToShop")

	p := discmsg.GetPrinter(language.AmericanEnglish)

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
				discmsg.SendEphemeralResponse(s, i, p.Sprintf("Invalid duration: %s", err.Error()))
				return
			}
		case "renewable":
			roleRenewable = option.BoolValue()
		}
	}
	if roleDesc == "" {
		roleDesc = roleName + " role"
	}

	// Verify the role exists on the server
	if role := guild.GetGuildRole(s, i.GuildID, roleName); role == nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Error("role not found on server")
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Role `%s` not found on the server.", roleName))
		return
	}

	// Add the role to the shop. If it already exists, this will return an error.
	shop := GetShop(i.GuildID)
	_, err := shop.AddShopItem(roleName, roleDesc, ROLE, roleCost, roleDuration, roleRenewable)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "roleDesc": roleDesc, "roleCost": roleCost, "roleDuration": roleDuration, "roleRenewable": roleRenewable}).Errorf("failed to add role to shop: %s", err)
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Failed to add role `%s` to the shop: %s", roleName, err))
		return
	}

	log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "roleDesc": roleDesc, "roleCost": roleCost, "roleDuration": roleDuration, "roleRenewable": roleRenewable}).Info("role added to shop")
	discmsg.SendNonEphemeralResponse(s, i, p.Sprintf("Role `%s` has been added to the shop.", roleName))
}

// removeShopItem routes the remove shop item commands to the proper handers.
func removeShopItem(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> shop.removeShopItem")
	defer log.Trace("<-- shop.removeShopItem")

	p := discmsg.GetPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options
	switch options[0].Options[0].Name {
	case "role":
		removeRoleFromShop(s, i)
	default:
		msg := p.Sprint("Command `%s\\%s` is not recognized.", options[0].Name, options[0].Options[0].Name)
		log.Warn(msg)
		discmsg.SendEphemeralResponse(s, i, msg)
	}
}

// removeRoleFromShop removes a role from the shop.
func removeRoleFromShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> shop.removeRoleFromShop")
	defer log.Trace("<-- shop.removeRoleFromShop")

	p := discmsg.GetPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options

	// Get the role details
	role := options[0].Options[0]
	roleName := role.Options[0].StringValue()

	shop := GetShop(i.GuildID)
	err := shop.RemoveShopItem(roleName, ROLE)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Errorf("failed to remove role from shop: %s", err)
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Failed to remove role `%s` from the shop: %s", roleName, err))
		return
	}

	log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Info("role removed from shop")
	discmsg.SendNonEphemeralResponse(s, i, p.Sprintf("Role `%s` has been removed from the shop.", roleName))
}

// updateShopItem routes the update shop item commands to the proper handers.
func updateShopItem(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> shop.updateShopItem")
	defer log.Trace("<-- shop.updateShopItem")

	p := discmsg.GetPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options
	switch options[0].Options[0].Name {
	case "role":
		updateRoleInShop(s, i)
	default:
		msg := p.Sprint("Command `%s\\%s` is not recognized.", options[0].Name, options[0].Options[0].Name)
		discmsg.SendEphemeralResponse(s, i, msg)
	}
}

// updateRoleInShop updates a role in the shop.
func updateRoleInShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> shop.updateRoleInShop")
	defer log.Trace("<-- shop.updateRoleInShop")

	p := discmsg.GetPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options

	// Get the role details
	role := options[0].Options[0]
	roleName := role.Options[0].StringValue()
	roleDesc := role.Options[1].StringValue()
	roleCost := int(role.Options[2].IntValue())
	var roleDuration string
	if len(role.Options) > 3 {
		roleDuration = role.Options[3].StringValue()
		_, err := disctime.ParseDuration(roleDuration)
		if err != nil {
			log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "roleDuration": roleDuration}).Errorf("Failed to parse role duration: %s", err)
			discmsg.SendEphemeralResponse(s, i, p.Sprintf("invalid duration: %s", err.Error()))
			return
		}
	}
	var roleRenewable bool
	if len(role.Options) > 4 {
		roleRenewable = role.Options[4].BoolValue()
	}

	item, err := readShopItem(i.GuildID, roleName, ROLE)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Errorf("failed to read role from shop: %s", err)
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Role `%s` not found in the shop.", roleName))
		return
	}

	if item == nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Error("Role not found in shop")
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Role `%s` not found in the shop.", roleName))
		return
	}

	err = item.UpdateShopItem(roleName, roleDesc, ROLE, roleCost, roleDuration, roleRenewable)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "roleDesc": roleDesc, "roleCost": roleCost, "roleDuration": roleDuration, "roleRenewable": roleRenewable}).Errorf("Failed to update role in shop: %s", err)
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Failed to update role `%s` in the shop: %s", roleName, err))
		return
	}

	log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "roleDesc": roleDesc, "roleCost": roleCost, "roleDuration": roleDuration, "roleRenewable": roleRenewable}).Info("Role updated in shop")
	discmsg.SendNonEphemeralResponse(s, i, p.Sprintf("Role `%s` has been updated in the shop.", roleName))
}

// listShopItems lists the items in the shop.
func listShopItems(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> shop.listShopItems")
	defer log.Trace("<-- shop.listShopItems")

	p := discmsg.GetPrinter(language.AmericanEnglish)

	shop := GetShop(i.GuildID)
	items := shop.Items

	if len(items) == 0 {
		log.WithFields(log.Fields{"guildID": i.GuildID}).Debug("no items found")
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("No items found in the shop."))
		return
	}

	shopItems := make([]*discordgo.MessageEmbedField, 0, len(items))
	for _, item := range items {
		sb := strings.Builder{}
		sb.WriteString(p.Sprintf("Description: %s\n", item.Description))
		sb.WriteString(p.Sprintf("Cost: %d", item.Price))
		if item.Duration != "" {
			duration, _ := disctime.ParseDuration(item.Duration)
			sb.WriteString(p.Sprintf("\nDuration: %s\n", disctime.FormatDuration(duration)))
			// sb.WriteString(p.Sprintf("Auto-Rewable: %t", item.AutoRenewable))
		}
		shopItems = append(shopItems, &discordgo.MessageEmbedField{
			Name:   p.Sprintf("%s %s", unicode.FirstToUpper(item.Type), item.Name),
			Value:  sb.String(),
			Inline: false,
		})
	}

	embeds := []*discordgo.MessageEmbed{
		{
			Title:  "Shop Items",
			Fields: shopItems,
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
		log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": i.Member.User.ID, "error": err}).Error("unable to send list of shop items")
		return
	}

	log.WithFields(log.Fields{"guildID": i.GuildID, "numItems": len(items)}).Info("shop items listed")
}

// shop routes the shop commands to the proper handers.
func shop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> shop.shop")
	defer log.Trace("<-- shop.shop")

	if status == discord.STOPPING || status == discord.STOPPED {
		discmsg.SendEphemeralResponse(s, i, "The system is currently shutting down.")
		return
	}

	p := discmsg.GetPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options
	switch options[0].Name {
	case "buy":
		buyFromShop(s, i)
	case "update":
		updatePurchaseFromShop(s, i)
	case "purchases":
		listPurchasesFromShop(s, i)
	case "list":
		listShopItems(s, i)
	default:
		msg := p.Sprint("Command `\\shop\\%s` is not recognized.", options[0].Name)
		discmsg.SendEphemeralResponse(s, i, msg)
	}
}

// buyFromShop routes the buy shop item commands to the proper handers.
func buyFromShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> shop.buyFromShop")
	defer log.Trace("<-- shop.buyFromShop")

	p := discmsg.GetPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options
	switch options[0].Options[0].Name {
	case "role":
		buyRoleFromShop(s, i)
	default:
		msg := p.Sprint("Command `%s\\%s` is not recognized.", options[0].Name, options[0].Options[0].Name)
		discmsg.SendEphemeralResponse(s, i, msg)
	}
}

// buyRoleFromShop buys a role from the shop.
func buyRoleFromShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> shop.buyRoleFromShop")
	defer log.Trace("<-- shop.buyRoleFromShop")

	p := discmsg.GetPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options

	// Get the role details
	role := options[0].Options[0]
	roleName := role.Options[0].StringValue()
	// None of the following are configurable at the present
	roleRenew := false

	// Verify the role exists on the server
	guildRole := guild.GetGuildRole(s, i.GuildID, roleName)
	if guildRole == nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Error("role not found on server")
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Role `%s` not found on the server.", roleName))
		return
	}

	// Make sure the member doesn't already have the role
	if guild.MemberHasRole(s, i.GuildID, i.Member.User.ID, guildRole) {
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("You already have the `%s` role.", roleName))
		return
	}

	// Make sure the role is still available in the shop
	shopItem := GetShopItem(i.GuildID, roleName, ROLE)
	if shopItem == nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Error("failed to read role from shop")
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Role `%s` not found in the shop.", roleName))
		return
	}

	// Make sure the role hasn't already been purchased
	purchase, _ := readPurchase(i.GuildID, i.Member.User.ID, roleName, ROLE)
	if purchase != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Error("role already purchased")
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("You have already purchased role `%s`.", roleName))
		return
	}

	// Purchase the role
	purchase, err := shopItem.Purchase(i.Member.User.ID, roleRenew)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "memberID": i.Member.User.ID, "error": err}).Errorf("failed to purchase role")
		discmsg.SendEphemeralResponse(s, i, unicode.FirstToUpper(err.Error()))
		return
	}

	// Assign the role to the user. If the role can't be assigned, then undo the purchase of the role.
	err = guild.AssignRole(s, i.GuildID, i.Member.User.ID, roleName)
	if err != nil {
		purchase.Return()
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "memberID": i.Member.User.ID, "error": err}).Error("failed to assign role")
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Failed to assign role `%s` to you: %s", roleName, err))
		return
	}

	log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": i.Member.User.ID, "roleName": roleName}).Info("role purchased")
	discmsg.SendNonEphemeralResponse(s, i, p.Sprintf("Role `%s` has been purchased.", roleName))
}

// upatePurchaseFromShop routes the update purchase shop item commands to the proper handers.
func updatePurchaseFromShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> shop.updatePurchaseFromShop")
	defer log.Trace("<-- shop.updatePurchaseFromShop")

	p := discmsg.GetPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options
	switch options[0].Options[0].Name {
	case "role":
		updateRolePurchaseFromShop(s, i)
	default:
		msg := p.Sprint("Command `%s\\%s` is not recognized.", options[0].Name, options[0].Options[0].Name)
		discmsg.SendEphemeralResponse(s, i, msg)
	}
}

// updateRolePurchaseFromShop updates a role purchase from the shop.
func updateRolePurchaseFromShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> shop.updateRolePurchaseFromShop")
	defer log.Trace("<-- shop.updateRolePurchaseFromShop")

	p := discmsg.GetPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options

	// Get the role details
	role := options[0].Options[0]
	roleName := role.Options[0].StringValue()
	roleRenew := role.Options[1].BoolValue()

	purchase, err := readPurchase(i.GuildID, i.Member.User.ID, roleName, ROLE)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "error": err}).Error("Failed to read purchase from shop")
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Purchase for role `%s` not found.", roleName))
		return
	}

	if purchase == nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Error("role purchase not found")
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("You have not purchases the %s role.", roleName))
		return
	}

	purchase.AutoRenew = roleRenew
	err = writePurchase(purchase)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "roleRenew": roleRenew}).Errorf("Failed to update role purchase: %s", err)
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Failed to update role `%s` purchase: %s", roleName, err))
		return
	}

	log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "roleRenew": roleRenew}).Info("Role purchase updated")
	discmsg.SendNonEphemeralResponse(s, i, p.Sprintf("Role`%s` purchase has been updated.", roleName))
}

// listPurchasesFromShop lists the purchases made by the member in the shop.
func listPurchasesFromShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> shop.listPurchasesFromShop")
	defer log.Trace("<-- shop.listPurchasesFromShop")

	p := discmsg.GetPrinter(language.AmericanEnglish)

	purchases := GetAllPurchases(i.GuildID, i.Member.User.ID)

	if len(purchases) == 0 {
		log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": i.Member.User.ID}).Debug("no purchases found")
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("You haven't made any purchases from the shop!"))
		return
	}

	purchasesMsg := make([]*discordgo.MessageEmbedField, 0, len(purchases))

	for _, purchase := range purchases {
		sb := strings.Builder{}
		sb.WriteString(p.Sprintf("Description: %s\n", purchase.Item.Description))
		sb.WriteString(p.Sprintf("Price: %d", purchase.Item.Price))
		if !purchase.ExpiresOn.IsZero() {
			if purchase.ExpiresOn.Before(time.Now()) {
				sb.WriteString(p.Sprintf("\nExpired On: %s\n", purchase.ExpiresOn.Format("Jan 02 2006")))
			} else {
				sb.WriteString(p.Sprintf("\nExpires On: %s\n", purchase.ExpiresOn.Format("Jan 02 2006")))
				sb.WriteString(p.Sprintf("Auto-Renew: %t\n", purchase.AutoRenew))
			}
		}
		purchasesMsg = append(purchasesMsg, &discordgo.MessageEmbedField{
			Name:   p.Sprintf("%s %s", unicode.FirstToUpper(purchase.Item.Type), purchase.Item.Name),
			Value:  sb.String(),
			Inline: false,
		})
	}

	embeds := []*discordgo.MessageEmbed{
		{
			Title:  "Purchases",
			Fields: purchasesMsg,
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
		log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": i.Member.User.ID, "error": err}).Error("unable to send shop purchases")
		return
	}

	log.WithFields(log.Fields{"guildID": i.GuildID, "numItems": len(purchases)}).Debug("shop purchases listed")
}
