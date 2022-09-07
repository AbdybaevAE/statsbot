package leetcode

import (
	"context"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

type TelegramBot struct {
	usersService UserService
	statService  StatsService
	*tgbotapi.BotAPI
}

func NewTelegramBot(usersService UserService, statService StatsService) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_API_TOKEN"))
	if err != nil {
		return nil, err
	}
	bot.Debug = true
	return &TelegramBot{
		usersService: usersService,
		statService:  statService,
		BotAPI:       bot,
	}, nil
}

const WelcomeStatsMessage = `
User @%s was successfully added for tracking!
Current user's statistics:
Easy / Medium / Hard solved - %d / %d / %d.
Easy / Medium / Hard submissions - %d / %d / %d.`

func (s *TelegramBot) RunBackgroundUserCommandsHandler(ctx context.Context) error {
	bot := s.BotAPI
	go func() {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		updates := bot.GetUpdatesChan(u)
		for {
			select {
			case update := <-updates:
				if update.Message == nil { // ignore any non-Message updates
					continue
				}
				if !update.Message.IsCommand() { // ignore any non-command Messages
					continue
				}

				// Create a new MessageConfig. We don't have text yet,
				// so we leave it empty.
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

				// Extract the command from the Message./
				switch update.Message.Command() {
				case "track_user":
					tokens := strings.Split(update.Message.Text, " ")
					if len(tokens) < 1 {
						msg.Text = "Invalid user profile's link provided."
						break
					}
					profileLink := tokens[1]
					log.Info().Msgf("Received profile link - %s", profileLink)
					userFrom := update.SentFrom()
					if userFrom == nil {
						msg.Text = "Cannot setup process :("
						break
					}
					args := AddUserArgs{
						TgUserId:    userFrom.ID,
						TgUserName:  userFrom.UserName,
						ChatId:      update.Message.Chat.ID,
						ProfileLink: profileLink,
					}
					if err := s.usersService.AddUser(ctx, args); err != nil {
						if errors.Is(err, ErrInvalidProfileUrl) {
							msg.Text = "Invalid user profile's link provided."
						} else if errors.Is(err, ErrUserWithProfileNotFound) {
							msg.Text = "Cannot find user for given profile link."
						} else if errors.Is(err, ErrUniqueConstraintViolation) {
							msg.Text = "Given user already exists"
						} else {
							msg.Text = "Something went wrong :("
						}
						break
					}
					createdUser, err := s.usersService.GetUserByTgIdAndChatId(ctx, userFrom.ID, update.Message.Chat.ID)
					if err != nil {
						msg.Text = "Something went wrong :("
					}
					stats, err := s.statService.InitStats(ctx, createdUser.UserId)
					if err != nil {
						// handle error
						msg.Text = "Something went wrong :(" // todo fix
						fmt.Println("error unknown", err)
						break
					}
					msg.Text = fmt.Sprintf(WelcomeStatsMessage,
						createdUser.LtUserName,
						stats.EasySolved,
						stats.MediumSolved,
						stats.HardSolved,
						stats.EasySubmissions,
						stats.MediumSubmissions,
						stats.HardSubmissions)
				default:
					msg.Text = "I don't know that command"
				}

				if _, err := bot.Send(msg); err != nil {
					log.Err(err).Msg("Cannot send message back...")
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil

}
func (s *TelegramBot) SendStats(ctx context.Context, items []StatItem) error {
	return nil
}
