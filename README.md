# ReadyCheck

### [Add to your Discord server](https://discord.com/api/oauth2/authorize?client_id=1086770261614931998&permissions=2048&scope=bot)

### What does it do?

#### `/lfg`

The `/lfg` command sends an LFG invitation message to the channel, which other members can respond to in various ways.

The LFG is highly customizable, with options for event name, start time, and number of people required. It also comes with the option to ping a role/server member.

| Option           | Type        | Description                                      | Default Value |
| ---------------- | ----------- | ------------------------------------------------ | ------------- |
| event-name       | string      | Name of the game / activity                      | "Any"         |
| time             | string      | Proposed start time (e.g. "in 15 min", "at 1pm") | "Now"         |
| number-of-people | uint8       | The number of _extra_ players needed             | 0             |
| notify           | Mentionable | The role or member to ping                       | nil           |

---

ReadyCheck is written entirely in Go and uses the package [DiscordGo](https://pkg.go.dev/github.com/bwmarrin/discordgo).
