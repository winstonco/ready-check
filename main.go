// https://github.com/bwmarrin/discordgo/blob/master/examples/airhorn/main.go

package main

import (
	"container/list"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	guildId = flag.String("g", "", "Test guild ID")
	token   = flag.String("t", "", "Bot token")
)

var discord *discordgo.Session

func init() {
	flag.Parse()
}

func init() {
	var err error
	discord, err = discordgo.New("Bot " + *token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}
}

type LFG struct {
	Game      string
	Time      string
	NumPeople uint8
	CreatedBy string
	SaidYes   *list.List
	SaidNo    *list.List
}

func (lfg *LFG) ClearName(username string) {
	for e := lfg.SaidYes.Front(); e != nil; e = e.Next() {
		if e.Value == username {
			lfg.SaidYes.Remove(e)
		}
	}
	for e := lfg.SaidNo.Front(); e != nil; e = e.Next() {
		if e.Value == username {
			lfg.SaidNo.Remove(e)
		}
	}
}

func (lfg *LFG) AddYes(username string) {
	lfg.ClearName(username)
	if lfg.CreatedBy != username {
		lfg.SaidYes.PushBack(username)
	}
}

func (lfg *LFG) AddNo(username string) {
	lfg.ClearName(username)
	lfg.SaidNo.PushBack(username)
}

func (lfg *LFG) GenerateEmbed() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       0x00ffff,
		Title:       lfg.Title(),
		Description: lfg.Desc(),
	}
}

func (lfg *LFG) Title() string {
	return fmt.Sprintf("lfg: %s\n", lfg.CreatedBy)
}

func (lfg *LFG) Desc() string {
	res := fmt.Sprintf("Event: %s\n", lfg.Game)
	res += fmt.Sprintf("Time: %s\n", lfg.Time)
	if lfg.NumPeople > 0 {
		if lfg.NumPeople == 1 {
			res += fmt.Sprintf("Requires %d more person", lfg.NumPeople)
		} else {
			res += fmt.Sprintf("Requires %d more people", lfg.NumPeople)
		}

		if lfg.SaidYes.Len() >= int(lfg.NumPeople) {
			res += " | Ready!\n"
		} else {
			res += " | Not enough people!\n"
		}
	}
	if lfg.SaidYes.Len() > 0 {
		res += "---\n"
	}
	for e := lfg.SaidYes.Front(); e != nil; e = e.Next() {
		res += fmt.Sprintf("%s is ready!\n", e.Value)
	}
	return res
}

var lfgMessages = make(map[string]*LFG)

func getUserName(member *discordgo.Member) string {
	if member.Nick != "" {
		return member.Nick
	}
	return member.User.Username
}

var msgComponents = []discordgo.MessageComponent{
	discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Emoji: discordgo.ComponentEmoji{
					Name: "👍",
				},
				Label:    "Ready!",
				Style:    discordgo.SuccessButton,
				CustomID: "yes",
			},
			discordgo.Button{
				Emoji: discordgo.ComponentEmoji{
					Name: "👎",
				},
				Label:    "Not Ready!",
				Style:    discordgo.DangerButton,
				CustomID: "no",
			},
		},
	},
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
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
			},
		},
	}
	componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"yes": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			messageID := i.Message.ID
			log.Printf("handler messageID: %s\n", messageID)
			lfg := lfgMessages[messageID]
			if lfg == nil {
				log.Println("lfg is nil")
				return
			}
			lfg.AddYes(getUserName(i.Member))
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Components: msgComponents,
					Embeds:     []*discordgo.MessageEmbed{lfg.GenerateEmbed()},
				},
			})
			if err != nil {
				panic(err)
			}
		},
		"no": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			messageID := i.Message.ID
			log.Printf("handler messageID: %s\n", messageID)
			lfg := lfgMessages[messageID]
			if lfg == nil {
				log.Println("lfg is nil")
				return
			}
			lfg.AddNo(getUserName(i.Member))

			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Components: msgComponents,
					Embeds:     []*discordgo.MessageEmbed{lfg.GenerateEmbed()},
				},
			})
			if err != nil {
				panic(err)
			}
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"lfg": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options

			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			newLFG := &LFG{
				Game:      "Any",
				Time:      "Now",
				NumPeople: 0,
				CreatedBy: getUserName(i.Member),
				SaidYes:   list.New(),
				SaidNo:    list.New(),
			}

			if option, ok := optionMap["event-name"]; ok {
				// Option values must be type asserted from interface{}.
				// Discordgo provides utility functions to make this simple.
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

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})

			message, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content:    "@frostfire",
				Components: msgComponents,
				Embeds:     []*discordgo.MessageEmbed{newLFG.GenerateEmbed()},
			})
			if err != nil {
				fmt.Println("Error creating follow-up message: ", err)
				return
			}
			if message == nil {
				log.Println("message is nil")
				return
			}
			lfgMessages[message.ID] = newLFG
			log.Printf("message messageID: %s\n", message.ID)
		},
	}
)

func init() {
	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionMessageComponent:
			if h, ok := componentHandlers[i.MessageComponentData().CustomID]; ok {
				h(s, i)
			}
		}
	})
}

func main() {
	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	// Open the websocket and begin listening.
	err := discord.Open()
	if err != nil {
		log.Fatalf("Error opening Discord session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := discord.ApplicationCommandCreate(discord.State.User.ID, *guildId, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer discord.Close()

	// Wait here until CTRL-C or other term signal is received.
	log.Println("ReadyCheck is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	discord.Close()
}
