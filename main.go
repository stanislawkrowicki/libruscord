package main

import (
	"flag"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "github",
			Description: "Kod źródłowy bota",
		},
		{
			Name:        "lekcje",
			Description: "Dzisiejsze lekcje",
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"github": SourceCode,
		"lekcje": FetchTodayLessons,
	}
)

var s *discordgo.Session

func init() {
	flag.Parse()
	_ = godotenv.Load()

	var err error
	s, err = discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatalf("Nieprawidłowy token bota.")
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot w gotowości!")
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Nie udało się nawiązać połączenia z Discord API. %v", err)
		return
	}

	for _, v := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Błąd przy tworzeniu komendy '%v': %v", v.Name, err)
		}
	}

	defer func(s *discordgo.Session) {
		_ = s.Close()
	}(s)

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Wyłączam...")
}
