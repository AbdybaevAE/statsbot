package entities

type UserEntity struct {
	UserId     int64  `db:"user_id"`
	TgUserId   int64  `db:"tg_user_id"`
	TgUserName string `db:"tg_username"`
	LtUserName string `db:"lt_username"`
	CreatedAt  string `db:"created_at"`
	ChatId     int64  `db:"chat_id"`
}
