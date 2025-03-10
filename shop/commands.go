package shop

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rbrabson/goblin/discord"
	"github.com/rbrabson/goblin/guild"
	"github.com/rbrabson/goblin/internal/discmsg"
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
							},
						},
					},
				},
				{
					Name:        "list",
					Description: "Lists the items in the shop.",
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
					Name:        "list",
					Description: "Lists the items in the shop that may be purchased.",
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
		msg := p.Sprint("Command \"%s\" is not recognized.", options[0].Name)
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
		msg := p.Sprint("Command \"%s\\%s\" is not recognized.", options[0].Name, options[0].Options[0].Name)
		discmsg.SendEphemeralResponse(s, i, msg)
	}
}

// addRoleToShop adds a role to the shop.
func addRoleToShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> shop.addRoleToShop")
	defer log.Trace("<-- shop.addRoleToShop")

	p := discmsg.GetPrinter(language.AmericanEnglish)

	options := i.ApplicationCommandData().Options

	// Get the role details, hadndlng optional paramateters
	role := options[0].Options[0]
	roleName := role.Options[0].StringValue()
	roleCost := int(role.Options[1].IntValue())
	// None of the following are configurable at the present
	roleDesc := roleName
	roleDuration := time.Duration(0)
	roleRenewable := false

	shop := GetShop(i.GuildID)
	_, err := shop.AddShopItem(roleName, roleDesc, SHOP, roleCost, roleDuration, roleRenewable)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "roleDesc": roleDesc, "roleCost": roleCost, "roleDuration": roleDuration, "roleRenewable": roleRenewable}).Errorf("Failed to add role to shop: %s", err)
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Failed to add role \"%s\" to the shop: %s", roleName, err))
		return
	}

	log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "roleDesc": roleDesc, "roleCost": roleCost, "roleDuration": roleDuration, "roleRenewable": roleRenewable}).Info("Role added to shop")
	discmsg.SendNonEphemeralResponse(s, i, p.Sprintf("Role \"%s\" has been added to the shop.", roleName))
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
		msg := p.Sprint("Command \"%s\\%s\" is not recognized.", options[0].Name, options[0].Options[0].Name)
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
	err := shop.RemoveShopItem(roleName, SHOP)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Errorf("Failed to remove role from shop: %s", err)
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Failed to remove role \"%s\" from the shop: %s", roleName, err))
		return
	}

	log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Info("Role removed from shop")
	discmsg.SendNonEphemeralResponse(s, i, p.Sprintf("Role \"%s\" has been removed from the shop.", roleName))
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
		msg := p.Sprint("Command \"%s\\%s\" is not recognized.", options[0].Name, options[0].Options[0].Name)
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
	var roleDuration time.Duration
	if len(role.Options) > 3 {
		roleDuration = time.Duration(role.Options[3].IntValue())
	}
	var roleRenewable bool
	if len(role.Options) > 4 {
		roleRenewable = role.Options[4].BoolValue()
	}

	item, err := readShopItem(i.GuildID, roleName, SHOP)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Errorf("Failed to read role from shop: %s", err)
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Role \"%s\" not found in the shop.", roleName))
		return
	}

	if item == nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Error("Role not found in shop")
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Role \"%s\" not found in the shop.", roleName))
		return
	}

	err = item.UpdateShopItem(roleName, roleDesc, SHOP, roleCost, roleDuration, roleRenewable)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "roleDesc": roleDesc, "roleCost": roleCost, "roleDuration": roleDuration, "roleRenewable": roleRenewable}).Errorf("Failed to update role in shop: %s", err)
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Failed to update role \"%s\" in the shop: %s", roleName, err))
		return
	}

	log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "roleDesc": roleDesc, "roleCost": roleCost, "roleDuration": roleDuration, "roleRenewable": roleRenewable}).Info("Role updated in shop")
	discmsg.SendNonEphemeralResponse(s, i, p.Sprintf("Role \"%s\" has been updated in the shop.", roleName))
}

// listShopItems lists the items in the shop.
func listShopItems(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> shop.listShopItems")
	defer log.Trace("<-- shop.listShopItems")

	p := discmsg.GetPrinter(language.AmericanEnglish)

	shop := GetShop(i.GuildID)
	items := shop.Items

	sb := strings.Builder{}
	for _, item := range items {
		sb.WriteString(p.Sprintf("`%s`", item.Name))
		sb.WriteString(p.Sprintf(", %s,", item.Type))
		sb.WriteString(p.Sprintf(" (%s)", item.Description))
		sb.WriteString(p.Sprintf(" $%d", item.Price))
		if item.Duration != 0 {
			sb.WriteString(p.Sprintf(", %s", item.Duration))
			sb.WriteString(p.Sprintf(", %t", item.AutoRenewable))
		}
		sb.WriteString("\n")
	}

	discmsg.SendEphemeralResponse(s, i, sb.String())
	log.WithFields(log.Fields{"guildID": i.GuildID, "numItems": len(items)}).Info("Shop items listed")
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
	case "list":
		listPurchasesFromShop(s, i)
	default:
		msg := p.Sprint("Command \"\\shop\\%s\" is not recognized.", options[0].Name)
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
		msg := p.Sprint("Command \"%s\\%s\" is not recognized.", options[0].Name, options[0].Options[0].Name)
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

	purchase, _ := readPurchase(i.GuildID, i.Member.User.ID, roleName, SHOP)
	if purchase != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Error("role already purchased")
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("You have already purchased role \"%s\".", roleName))
		return
	}

	shopItem := GetShopItem(i.GuildID, roleName, SHOP)
	if shopItem == nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName}).Error("Failed to read role from shop")
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Role \"%s\" not found in the shop.", roleName))
		return
	}

	// Purchase the role
	_, err := shopItem.Purchase(i.Member.User.ID, roleRenew)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "memberID": i.Member.User.ID, "error": err}).Errorf("failed to purchase role")
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Failed to purchase role \"%s\"", roleName))
		return
	}

	log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "memberID": i.Member.User.ID}).Info("Role purchased")
	discmsg.SendNonEphemeralResponse(s, i, p.Sprintf("Role \"%s\" has been purchased.", roleName))
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
		msg := p.Sprint("Command \"%s\\%s\" is not recognized.", options[0].Name, options[0].Options[0].Name)
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

	purchase, err := readPurchase(i.GuildID, i.Member.User.ID, roleName, SHOP)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "error": err}).Error("Failed to read purchase from shop")
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Purchase for role \"%s\" not found.", roleName))
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
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Failed to update role \"%s\" purchase: %s", roleName, err))
		return
	}

	log.WithFields(log.Fields{"guildID": i.GuildID, "roleName": roleName, "roleRenew": roleRenew}).Info("Role purchase updated")
	discmsg.SendNonEphemeralResponse(s, i, p.Sprintf("Role \"%s\" purchase has been updated.", roleName))
}

// listPurchasesFromShop lists the purchases made by the member in the shop.
func listPurchasesFromShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Trace("--> shop.listPurchasesFromShop")
	defer log.Trace("<-- shop.listPurchasesFromShop")

	p := discmsg.GetPrinter(language.AmericanEnglish)

	purchases, err := readPurchases(i.GuildID, i.Member.User.ID)
	if err != nil {
		log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": i.Member.User.ID, "error": err}).Error("failed to read purchases from shop")
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("Failed to read purchases from the shop: %s", err))
		return
	}

	if len(purchases) == 0 {
		log.WithFields(log.Fields{"guildID": i.GuildID, "memberID": i.Member.User.ID}).Debug("no purchases found")
		discmsg.SendEphemeralResponse(s, i, p.Sprintf("You haven't made any purchases from the shop!"))
		return
	}

	sb := strings.Builder{}
	for _, purchase := range purchases {
		sb.WriteString(p.Sprintf("`%s`", purchase.Item.Name))
		sb.WriteString(p.Sprintf(", %s,", purchase.Item.Type))
		sb.WriteString(p.Sprintf(" (%s)", purchase.Item.Description))
		sb.WriteString(p.Sprintf(" $%d", purchase.Item.Price))
		if purchase.Item.Duration != 0 {
			sb.WriteString(p.Sprintf(", %s", purchase.Item.Duration))
			sb.WriteString(p.Sprintf(", %t", purchase.AutoRenew))
		}
		sb.WriteString("\n")
	}

	discmsg.SendEphemeralResponse(s, i, sb.String())
	log.WithFields(log.Fields{"guildID": i.GuildID, "numItems": len(purchases)}).Debug("shop purchases listed")
}
