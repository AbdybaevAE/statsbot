package main

import (
	"context"
	"fmt"
	"github.com/abdybaevae/leetcodestats/database"
	"github.com/abdybaevae/leetcodestats/pkg/repositories"
	"github.com/abdybaevae/leetcodestats/pkg/services/leetcode"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	ctx := context.Background()
	if err := godotenv.Load(); err != nil {
		log.Warn().Msg("cannot load config file")
	}
	db, err := database.Connect(ctx)
	if err != nil {
		log.Fatal().Msgf("cannot connect to database due to err: %s", err)
	}
	statRepo := repositories.NewStatRepo(db)
	userRepo := repositories.NewUserRepo(db)
	leetcodeClient := leetcode.NewLeetCodeClient()
	statService := leetcode.NewStatsService(leetcodeClient, statRepo, userRepo)
	userService := leetcode.NewUserService(userRepo, statService, leetcodeClient)
	bot, err := leetcode.NewTelegramBot(userService, statService)
	if err != nil {
		log.Fatal().Msgf("cannot init bot due to %v", err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err = bot.RunBackgroundUserCommandsHandler(ctx); err != nil {
			log.Fatal().Msgf("cannot init bot due to %w", err)
		}
	}()
	wg.Add(1)
	go func() {
		log.Fatal().Err(statService.Run(ctx, handleUpdates(bot)))
		defer wg.Done()
	}()
	wg.Add(1)
	go func() {

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Welcome to new server!")
		})

		// listen to port
		http.ListenAndServe(":"+os.Getenv("PORT"), nil)

	}()
	wg.Wait()
}

var welcome = []string{"Statistics time!", "3", "2", "1"}

func handleUpdates(bot *leetcode.TelegramBot) leetcode.OnUpdateStatHandler {

	return func(items []leetcode.StatDiff) {
		diffByChatId := map[int64][]leetcode.StatDiff{}
		for _, stat := range items {
			diffByChatId[stat.ChatId] = append(diffByChatId[stat.ChatId], stat)
		}
		for chatId, items := range diffByChatId {
			for _, mes := range welcome {
				msg := tgbotapi.NewMessage(chatId, mes)
				if _, err := bot.Send(msg); err != nil {
					log.Err(err).Msg("Cannot send stat message...")
				}
				time.Sleep(1)
			}

			for _, stat := range items {
				before := stat.Before
				after := stat.After
				message := fmt.Sprintf(`
For the past day @%s have following progress:
Easy solved - %d, submissions - %d
Medium solved - %d, submissions - %d
Hard solved - %d, submissions - %d`, stat.LtUserName,
					after.EasySolved-before.EasySolved, after.EasySubmissions-before.EasySubmissions,
					after.MediumSolved-before.MediumSolved, after.MediumSubmissions-before.MediumSubmissions,
					after.HardSolved-before.HardSolved, after.HardSubmissions-before.HardSubmissions)
				msg := tgbotapi.NewMessage(stat.ChatId, message)
				if _, err := bot.Send(msg); err != nil {
					log.Err(err).Msg("Cannot send stat message...")
				}
			}
		}

	}
}
