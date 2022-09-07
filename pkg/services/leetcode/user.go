package leetcode

import (
	"context"
	"errors"
	"github.com/abdybaevae/leetcodestats/pkg/entities"
	"github.com/abdybaevae/leetcodestats/pkg/repositories"
	"github.com/rs/zerolog/log"
	"net/url"
	"strings"
	"time"
)

type AddUserArgs struct {
	TgUserId    int64
	TgUserName  string
	ProfileLink string
	ChatId      int64
}

type UserService interface {
	AddUser(ctx context.Context, args AddUserArgs) (err error)
	GetUserByTgIdAndChatId(ctx context.Context, tgUserId int64, chatId int64) (foundUser *entities.UserEntity, err error)
}

func NewUserService(userRepo repositories.UserRepo, statService StatsService, leetcodeClient LeetCodeClient) UserService {
	return &userServiceImpl{
		userRepo,
		statService,
		leetcodeClient,
	}
}

type userServiceImpl struct {
	userRepo       repositories.UserRepo
	statService    StatsService
	leetcodeClient LeetCodeClient
}

var ErrInvalidProfileUrl = errors.New("invalid profile link provided")
var ErrUserWithProfileNotFound = errors.New("user for given profile wasn't found")
var ErrUniqueConstraintViolation = errors.New("unique constraint violation error")

const pgConstraintViolationText = "duplicate key value violates unique constraint"

func (s *userServiceImpl) AddUser(ctx context.Context, args AddUserArgs) error {
	profileLink := strings.Trim(args.ProfileLink, " ")
	if len(profileLink) == 0 {
		return ErrInvalidProfileUrl
	}
	if profileLink[len(profileLink)-1] == '/' {
		profileLink = profileLink[:len(profileLink)-1]
	}
	// parse ltUserId from profileLink
	u, err := url.Parse(profileLink)
	if err != nil {
		return err
	}
	if u.Host != "leetcode.com" {
		return ErrInvalidProfileUrl
	}
	tokens := strings.Split(u.Path, "/")
	if len(tokens) == 0 {
		return ErrInvalidProfileUrl
	}
	ltUserId := tokens[len(tokens)-1]
	log.Info().Msgf("found ltUserId is %s", ltUserId)
	if len(ltUserId) == 0 {
		return ErrInvalidProfileUrl
	}

	// check ltUserId in leetcode
	if _, err = s.leetcodeClient.GetUserStats(ctx, ltUserId); err != nil {
		return ErrUserWithProfileNotFound
	}

	// init user in database
	userEntity := &entities.UserEntity{
		TgUserId:   args.TgUserId,
		TgUserName: args.TgUserName,
		LtUserName: ltUserId,
		ChatId:     args.ChatId,
		CreatedAt:  time.Now().Format(time.RFC3339Nano),
	}
	if err := s.userRepo.AddUser(ctx, userEntity); err != nil {
		if strings.Index(err.Error(), pgConstraintViolationText) != -1 {
			return ErrUniqueConstraintViolation //todo fixup
		}
		return err
	}
	return nil
}
func (s *userServiceImpl) GetUserByTgIdAndChatId(ctx context.Context, tgUserId int64, chatId int64) (foundUser *entities.UserEntity, err error) {
	return s.userRepo.GetUserByTgIdAndChatId(ctx, tgUserId, chatId)
}
