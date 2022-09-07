package repositories

import (
	"context"
	"github.com/abdybaevae/leetcodestats/pkg/entities"
	"github.com/jmoiron/sqlx"
)

type StatRepo interface {
	SaveStat(ctx context.Context, stat entities.StatEntity) error
	GetStatByHash(ctx context.Context, hash string) (*entities.StatEntity, error)
}
type statRepoImpl struct {
	db *sqlx.DB
}

func NewStatRepo(db *sqlx.DB) StatRepo {
	return &statRepoImpl{db}
}

const SaveStatQuery = `
	insert into stats (
		user_id, 
		stat_hash, 
		stat_easy_solved, 
		stat_easy_submissions, 
		stat_medium_solved, 
		stat_medium_submissions, 
		stat_hard_solved, 
		stat_hard_submissions ) 
	values (
	        :user_id, 
	        :stat_hash, 
	        :stat_easy_solved, 
	        :stat_easy_submissions, 
	        :stat_medium_solved, 
	        :stat_medium_submissions, 
	        :stat_hard_solved, 
	        :stat_hard_submissions )
`

func (s *statRepoImpl) SaveStat(ctx context.Context, statEntity entities.StatEntity) error {
	if _, err := s.db.NamedExec(SaveStatQuery, statEntity); err != nil {
		return err
	}
	return nil
}

const getStatByHashQuery = `
	select * from stats 
		where stat_hash = $1
`

func (s *statRepoImpl) GetStatByHash(ctx context.Context, hash string) (*entities.StatEntity, error) {
	foundStatEntity := &entities.StatEntity{}
	if err := s.db.GetContext(ctx, foundStatEntity, getStatByHashQuery, hash); err != nil {
		return foundStatEntity, err
	}
	return foundStatEntity, nil
}
