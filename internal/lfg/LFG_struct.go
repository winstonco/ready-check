package lfg

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var MessageIDLFGMap = make(map[string]*LFG)

type Mentionable interface {
	Mention() string
}

type YesInType struct {
	name string
	time string
}

type LFG struct {
	Game      string
	Time      string
	NumPeople uint8
	ToNotify  Mentionable
	CreatedBy string
	SaidYes   []string
	SaidYesIn []YesInType
	SaidNo    []string
}

func (lfg *LFG) ClearName(username string) {
	for i, e := range lfg.SaidYes {
		if e == username {
			lfg.SaidYes = append(lfg.SaidYes[:i], lfg.SaidYes[i+1:]...)
		}
	}
	for i, e := range lfg.SaidNo {
		if e == username {
			lfg.SaidNo = append(lfg.SaidNo[:i], lfg.SaidNo[i+1:]...)
		}
	}
	for i, e := range lfg.SaidYesIn {
		if e.name == username {
			lfg.SaidYesIn = append(lfg.SaidYesIn[:i], lfg.SaidYesIn[i+1:]...)
		}
	}
}

func (lfg *LFG) AddYes(username string) {
	lfg.ClearName(username)
	if lfg.CreatedBy != username {
		lfg.SaidYes = append(lfg.SaidYes, username)
	}
}

func (lfg *LFG) AddNo(username string) {
	lfg.ClearName(username)
	lfg.SaidNo = append(lfg.SaidNo, username)
}

func (lfg *LFG) AddYesIn(username string, time string) {
	lfg.ClearName(username)
	if lfg.CreatedBy != username {
		lfg.SaidYesIn = append(lfg.SaidYesIn, YesInType{
			name: username,
			time: time,
		})
	}
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

		if len(lfg.SaidYes) >= int(lfg.NumPeople) {
			res += " | Ready!\n"
		} else {
			res += " | Not enough people!\n"
		}
	}
	if len(lfg.SaidYes) > 0 || len(lfg.SaidYesIn) > 0 {
		res += "---\n"
	}
	for _, e := range lfg.SaidYes {
		res += fmt.Sprintf("%s is ready!\n", e)
	}
	for _, e := range lfg.SaidYesIn {
		res += fmt.Sprintf("%s is ready in %s!\n", e.name, e.time)
	}
	return res
}
