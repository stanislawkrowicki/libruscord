package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"os"
)

const (
	githubRepo        = "https://github.com/stanislawkrowicki/libruscord"
	embedColor        = 0x00C09A
	sourceEmbedTitle  = "Link do repozytorium"
	lessonsEmbedTitle = "Lekcje na dziś"
	envLogin          = "LIBRUS_LOGIN"
	envPassword       = "LIBRUS_PASSWORD"
)

func createLessonsEmbed(lessons []LessonEntity) discordgo.MessageEmbed {
	var fields []*discordgo.MessageEmbedField

	for _, lesson := range lessons {
		name := fmt.Sprintf("%s: %s", lesson.Number, lesson.Subject)
		value := fmt.Sprintf("%s - %s [Dołącz](%s)", lesson.Teacher, lesson.Group, lesson.URL)
		field := discordgo.MessageEmbedField{Name: name, Value: value}
		fields = append(fields, &field)
	}

	return discordgo.MessageEmbed{
		Title:  lessonsEmbedTitle,
		Color:  embedColor,
		Fields: fields,
	}
}

func SourceCode(s *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := discordgo.MessageEmbed{
		Title:       sourceEmbedTitle,
		Color:       embedColor,
		Description: fmt.Sprintf("[GitHub](%s)", githubRepo),
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{&embed},
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
