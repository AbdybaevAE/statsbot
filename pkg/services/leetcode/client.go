package leetcode

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
)

var (
	ErrUnexpectedLeetCodeStatusCode = errors.New("unexpected response status code from leetcode")
	ErrNoSuchUser                   = errors.New("No such user")
)

type UserStatsRes struct {
	EasySolved        int
	EasySubmissions   int
	MediumSolved      int
	MediumSubmissions int
	HardSolved        int
	HardSubmissions   int
}

type LeetCodeClient interface {
	GetUserStats(ctx context.Context, userId string) (*UserStatsRes, error)
}

func NewLeetCodeClient() LeetCodeClient {
	return &leetCodeClientImpl{}
}

const query = `{"query":"{\n  matchedUser(username: \"%s\") {\n    submitStats: submitStatsGlobal {\n      acSubmissionNum {\n        difficulty\n        count\n        submissions\n      }\n    }\n  }\n}","variables":{}}
`

type leetCodeClientImpl struct {
}
type leetCodeUserStatsResponse struct {
	Data struct {
		MatchedUser struct {
			SubmitStats struct {
				AcSubmissionNum []struct {
					Difficulty  string `json:"difficulty"`
					Count       int    `json:"count"`
					Submissions int    `json:"submissions"`
				} `json:"acSubmissionNum"`
			} `json:"submitStats"`
		} `json:"matchedUser"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (s *leetCodeClientImpl) GetUserStats(ctx context.Context, userId string) (*UserStatsRes, error) {
	log.Info().Msgf("looking for user stats %s", userId)
	req, err := http.NewRequest(http.MethodPost, "https://leetcode.com/graphql", strings.NewReader(fmt.Sprintf(query, userId)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	log.Info().Msgf("status code is ", res.StatusCode)
	if res.StatusCode != 200 {
		return nil, ErrUnexpectedLeetCodeStatusCode
	}
	var response leetCodeUserStatsResponse
	log.Info().Interface("response", response)
	//fmt.Println(res.Body)
	//result, err := io.ReadAll(res.Body)
	//if err != nil {
	//	fmt.Println("error 79", err)
	//}
	//fmt.Println(string(result))
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}
	if len(response.Errors) != 0 {
		return nil, ErrNoSuchUser
	}
	var stats UserStatsRes
	for _, v := range response.Data.MatchedUser.SubmitStats.AcSubmissionNum {
		switch v.Difficulty {
		case "Easy":
			stats.EasySolved = v.Count
			stats.EasySubmissions = v.Submissions
			break
		case "Medium":
			stats.MediumSolved = v.Count
			stats.MediumSubmissions = v.Submissions
			break
		case "Hard":
			stats.HardSolved = v.Count
			stats.HardSubmissions = v.Submissions
			break
		default:
			break
		}
	}
	return &stats, nil
}
