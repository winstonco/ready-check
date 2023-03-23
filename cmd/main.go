// https://github.com/bwmarrin/discordgo/blob/master/examples/airhorn/main.go

package main

import (
	"ready-check/internal/config"
	"ready-check/internal/lfg"
	"ready-check/internal/slashcommands"

	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func init() {
	config.LoadConfig()
}

var discord *discordgo.Session

func init() {
	var err error
	discord, err = discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}
}

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
}

var (
	commands = []*discordgo.ApplicationCommand{
		slashcommands.LfgCommand,
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		slashcommands.LfgCommand.Name: slashcommands.LfgHandler,
	}
)

var componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"yes": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		messageID := i.Message.ID
		log.Printf("handler messageID: %s\n", messageID)
		lfg := lfg.MessageIDLFGMap[messageID]
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
	"yes-in": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: "ready-in-modal",
				Title:    "Ready in...",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "ready-in-input",
								Label:       "How long?",
								Style:       discordgo.TextInputShort,
								Placeholder: "10 min",
								Required:    true,
							},
						},
					},
				},
			},
		})
		if err != nil {
			panic(err)
		}
	},
	"no": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		messageID := i.Message.ID
		log.Printf("handler messageID: %s\n", messageID)
		lfg := lfg.MessageIDLFGMap[messageID]
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
		case discordgo.InteractionModalSubmit:
			messageID := i.Message.ID
			log.Printf("handler messageID: %s\n", messageID)
			lfg := lfg.MessageIDLFGMap[messageID]
			if lfg == nil {
				log.Println("lfg is nil")
				return
			}
			lfg.AddYesIn(getUserName(i.Member), i.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value)
			fmt.Println(i.ModalSubmitData())
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
		cmd, err := discord.ApplicationCommandCreate(discord.State.User.ID, os.Getenv("GUILD_ID"), v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	log.Println("Updating status...")
	err = discord.UpdateGameStatus(0, "/lfg")
	if err != nil {
		log.Fatalf("Error updating game status: %v", err)
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
