package slashcommands

import (
	"fmt"
	"log"
	"ready-check/internal/lfg"

	"github.com/bwmarrin/discordgo"
)

var LfgCommand *discordgo.ApplicationCommand = &discordgo.ApplicationCommand{
	Name:        "lfg",
	Description: "Create lfg",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "event-name",
			Description: "Name of the game / activity",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "time",
			Description: "Proposed start time (e.g. \"in 15\", \"at 1\")",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "number-of-people",
			Description: "The number of extra players needed",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionMentionable,
			Name:        "notify",
			Description: "The role/member to ping",
			Required:    false,
		},
	},
}

var LfgHandler func(s *discordgo.Session, i *discordgo.InteractionCreate) = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	var username string
	if i.Member.Nick != "" {
		username = i.Member.Nick
	} else {
		username = i.Member.User.Username
	}

	newLFG := &lfg.LFG{
		Game:      "Any",
		Time:      "Now",
		NumPeople: 0,
		CreatedBy: username,
		SaidYes:   make([]string, 0),
		SaidNo:    make([]string, 0),
		SaidYesIn: make([]lfg.YesInType, 0),
	}

	if option, ok := optionMap["event-name"]; ok {
		if option.StringValue() != "" {
			newLFG.Game = option.StringValue()
		}
	}

	if option, ok := optionMap["time"]; ok {
		if option.StringValue() != "" {
			newLFG.Time = option.StringValue()
		}
	}

	if option, ok := optionMap["number-of-people"]; ok {
		if option.IntValue() != 0 {
			newLFG.NumPeople = uint8(option.IntValue())
		}
	}

	var toMention lfg.Mentionable
	if option, ok := optionMap["notify"]; ok {
		user := option.UserValue(s)
		if user != nil {
			newLFG.ToNotify = user
			toMention = user
		}
		role := option.RoleValue(s, i.GuildID)
		if role != nil {
			newLFG.ToNotify = role
			toMention = role
		}
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	message, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf("%s\n", toMention.Mention()),
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						CustomID: "yes",
						Emoji: discordgo.ComponentEmoji{
							Name: "👍",
						},
						Label: "Yes!",
						Style: discordgo.SuccessButton,
					},
					discordgo.Button{
						CustomID: "no",
						Emoji: discordgo.ComponentEmoji{
							Name: "👎",
						},
						Label: "No!",
						Style: discordgo.DangerButton,
					},
					discordgo.Button{
						CustomID: "yes-in",
						Emoji: discordgo.ComponentEmoji{
							Name: "⏳",
						},
						Label: "Ready in...",
						Style: discordgo.SecondaryButton,
					},
				},
			},
		},
		Embeds: []*discordgo.MessageEmbed{newLFG.GenerateEmbed()},
	})
	if err != nil {
		log.Println("Error creating follow-up message: ", err)
		return
	}
	if message == nil {
		log.Println("message is nil")
		return
	}
	lfg.MessageIDLFGMap[message.ID] = newLFG
	// log.Printf("message messageID: %s\n", message.ID)
}
