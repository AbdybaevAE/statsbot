package repositories

import (
	"context"
	"fmt"
	"github.com/abdybaevae/leetcodestats/pkg/entities"
	"github.com/jmoiron/sqlx"
)

type UserRepo interface {
	AddUser(ctx context.Context, user *entities.UserEntity) error
	GetAllUsers(ctx context.Context) ([]entities.UserEntity, error)
	GetUserById(ctx context.Context, userId int64) (foundUser *entities.UserEntity, err error)
	GetUserByTgIdAndChatId(ctx context.Context, tgId int64, chatId int64) (foundUser *entities.UserEntity, err error)
}
type useRepoImpl struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) UserRepo {
	return &useRepoImpl{
		db: db,
	}
}

const addUserQuery = `insert into users (tg_user_id, tg_username, lt_username, created_at, chat_id) values (:tg_user_id, :tg_username, :lt_username, :created_at, :chat_id)`

func (s *useRepoImpl) AddUser(ctx context.Context, user *entities.UserEntity) error {
	res, err := s.db.NamedExecContext(ctx, addUserQuery, user)
	if err != nil {
		return err
	}
	fmt.Println(res)
	return nil
}

const getAllUsersQuery = `select * from users`

func (s *useRepoImpl) GetAllUsers(ctx context.Context) ([]entities.UserEntity, error) {
	var userEntities []entities.UserEntity
	if err := s.db.SelectContext(ctx, &userEntities, getAllUsersQuery); err != nil {
		return nil, err
	}
	return userEntities, nil
}

const getUserByIdQuery = `select * from users where user_id = $1`

func (s *useRepoImpl) GetUserById(ctx context.Context, userId int64) (*entities.UserEntity, error) {
	userEntity := &entities.UserEntity{}
	if err := s.db.GetContext(ctx, userEntity, getUserByIdQuery, userId); err != nil {
		return nil, err
	}
	return userEntity, nil
}

const getUserByTgIdAndChatIdQuery = `select * from users where tg_user_id = $1 and chat_id = $2`

func (s *useRepoImpl) GetUserByTgIdAndChatId(ctx context.Context, tgId int64, chatId int64) (*entities.UserEntity, error) {
	foundUser := &entities.UserEntity{}
	if err := s.db.GetContext(ctx, foundUser, getUserByTgIdAndChatIdQuery, tgId, chatId); err != nil {
		return foundUser, err
	}
	return foundUser, nil
}
