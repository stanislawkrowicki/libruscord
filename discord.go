package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"os"
)

const (
	githubRepo  = "https://github.com/stanislawkrowicki/libruscord"
	embedColor  = 16711935
	embedTitle  = "Lekcje na dziś"
	envLogin    = "LIBRUS_LOGIN"
	envPassword = "LIBRUS_PASSWORD"
)

func createLessonsEmbed(lessons []LessonEntity) discordgo.MessageEmbed {
	content := ""

	for _, lesson := range lessons {
		content += fmt.Sprintf("%s: %s (%s) - %s [Dołącz](%s)\n\n", lesson.Number, lesson.Subject, lesson.Teacher, lesson.Group, lesson.URL)
	}

	return discordgo.MessageEmbed{
		URL:         githubRepo,
		Title:       embedTitle,
		Color:       embedColor,
		Description: content,
	}
}

func SourceCode(s *discordgo.Session, i *discordgo.InteractionCreate) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: githubRepo,
		},
	})
}

func FetchTodayLessons(s *discordgo.Session, i *discordgo.InteractionCreate) {
	_ = godotenv.Load()

	librusSession, err := Login(os.Getenv(envLogin), os.Getenv(envPassword))
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Brrt. Nie udało mi się zalogować :/ (%v)", err),
			},
		})
		return
	}

	lessons, err := librusSession.GetLessons()
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Hmm, coś poszło nie tak.. Błąd pobierania planu lekcji. (%v)", err),
			},
		})
		return
	}
	if len(lessons) == 0 {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Hurra! Wygląda na to, że dzisiaj nie ma żadnych lekcji.",
			},
		})
		return
	}

	embed := createLessonsEmbed(lessons)

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{&embed},
		},
	})
}
