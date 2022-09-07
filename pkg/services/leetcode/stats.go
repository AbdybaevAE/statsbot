package leetcode

import (
	"context"
	"fmt"
	"github.com/abdybaevae/leetcodestats/pkg/entities"
	"github.com/abdybaevae/leetcodestats/pkg/repositories"
	"log"
	"sync"
	"time"
)

type StatItem struct {
	EasySolved        int `json:"easy_solved"`
	EasySubmissions   int `json:"easy_submissions"`
	MediumSolved      int `json:"medium_solved"`
	MediumSubmissions int `json:"medium_submissions"`
	HardSolved        int `json:"hard_solved"`
	HardSubmissions   int `json:"hard_submissions"`
}
type StatDiff struct {
	LtUserName string
	ChatId     int64
	Before     StatItem
	After      StatItem
}
type OnUpdateStatHandler func(stat []StatDiff)

type StatsService interface {
	Run(ctx context.Context, handler OnUpdateStatHandler) (err error)
	InitStats(ctx context.Context, tgUserId int64) (currUserStats *UserStatsRes, err error)
}
type service struct {
	leetCodeClient LeetCodeClient
	statRepo       repositories.StatRepo
	userRepo       repositories.UserRepo
}

func NewStatsService(client LeetCodeClient, statRepo repositories.StatRepo, userRepo repositories.UserRepo) StatsService {
	return &service{
		leetCodeClient: client,
		statRepo:       statRepo,
		userRepo:       userRepo,
	}
}

func (s *service) Run(ctx context.Context, handler OnUpdateStatHandler) error {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			allUsers, err := s.userRepo.GetAllUsers(ctx)
			if err != nil {
				log.Printf("cannot access users from database cause of err: %s", err)
			}
			var results []StatDiff
			var lock sync.Mutex
			var wg sync.WaitGroup
			for _, user := range allUsers {
				go func(userEntity entities.UserEntity) {
					defer wg.Done()
					stat, err := s.leetCodeClient.GetUserStats(ctx, userEntity.LtUserName)
					if err != nil {
						log.Printf("cannot retrieve users info %s because of err: %s", userEntity, err)
						return
					}
					statEntity := entities.StatEntity{
						UserId:            userEntity.UserId,
						Hash:              s.getTodayStatHash(userEntity.LtUserName),
						EasySolved:        stat.EasySolved,
						EasySubmissions:   stat.EasySubmissions,
						MediumSolved:      stat.MediumSolved,
						MediumSubmissions: stat.MediumSubmissions,
						HardSolved:        stat.HardSolved,
						HardSubmissions:   stat.HardSubmissions,
					}
					if err = s.statRepo.SaveStat(ctx, statEntity); err != nil {
						log.Printf("Cannot save user stats cause of err :%s", err)
						return
					}
					lock.Lock()
					defer lock.Unlock()
					yesterdayStat, err := s.statRepo.GetStatByHash(ctx, s.getYesterdayStatHash(userEntity.LtUserName))
					if err != nil {
						log.Printf("cannot retrieve yesterday stat")
						return
					}
					results = append(results, StatDiff{
						LtUserName: userEntity.LtUserName,
						ChatId:     userEntity.ChatId,
						Before: StatItem{
							EasySolved:        yesterdayStat.EasySolved,
							EasySubmissions:   yesterdayStat.EasySubmissions,
							MediumSolved:      yesterdayStat.MediumSolved,
							MediumSubmissions: yesterdayStat.MediumSubmissions,
							HardSolved:        yesterdayStat.MediumSolved,
							HardSubmissions:   yesterdayStat.HardSubmissions,
						},
						After: StatItem{
							EasySolved:        statEntity.EasySolved,
							EasySubmissions:   statEntity.EasySubmissions,
							MediumSolved:      statEntity.MediumSolved,
							MediumSubmissions: statEntity.MediumSubmissions,
							HardSolved:        statEntity.MediumSolved,
							HardSubmissions:   statEntity.HardSubmissions,
						},
					})
				}(user)
				wg.Add(1)
			}
			wg.Wait()
			if len(results) != 0 {
				go handler(results)
			}
		case <-ctx.Done():
			return nil
		}
	}
}
func (s *service) InitStats(ctx context.Context, userId int64) (*UserStatsRes, error) {
	user, err := s.userRepo.GetUserById(ctx, userId)
	if err != nil {
		return nil, err
	}
	stats, err := s.leetCodeClient.GetUserStats(ctx, user.LtUserName)
	fmt.Println("stats from leetcode direct call", stats)
	if err != nil {
		return nil, err
	}
	statEntity := entities.StatEntity{
		UserId:            userId,
		EasySubmissions:   stats.EasySubmissions,
		EasySolved:        stats.EasySolved,
		MediumSolved:      stats.MediumSolved,
		MediumSubmissions: stats.MediumSubmissions,
		HardSolved:        stats.HardSolved,
		HardSubmissions:   stats.HardSubmissions,
		Hash:              s.getTodayStatHash(user.LtUserName),
	}
	return stats, s.statRepo.SaveStat(ctx, statEntity)
}
func (s *service) generateStatTime(time time.Time, ltId string) string {
	return fmt.Sprintf("everyday:%s:%s", ltId, time.Format("2006-01-02"))
}
func (s *service) getYesterdayStatHash(ltId string) string {
	return s.generateStatTime(time.Now().AddDate(0, 0, -1), ltId)
}
func (s *service) getTodayStatHash(ltId string) string {
	return s.generateStatTime(time.Now(), ltId)
}
