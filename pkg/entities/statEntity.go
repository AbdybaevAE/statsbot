package entities

type StatEntity struct {
	Id                string `db:"stat_id"`
	UserId            int64  `db:"user_id"`
	Hash              string `db:"stat_hash"`
	EasySolved        int    `db:"stat_easy_solved"`
	EasySubmissions   int    `db:"stat_easy_submissions"`
	MediumSolved      int    `db:"stat_medium_solved"`
	MediumSubmissions int    `db:"stat_medium_submissions"`
	HardSolved        int    `db:"stat_hard_solved"`
	HardSubmissions   int    `db:"stat_hard_submissions"`
}
