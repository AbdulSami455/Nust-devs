package github

import "time"

type User struct {
	Login       string `json:"login"`
	Name        string `json:"name"`
	AvatarURL   string `json:"avatar_url"`
	Bio         string `json:"bio"`
	Location    string `json:"location"`
	Company     string `json:"company"`
	Blog        string `json:"blog"`
	Email       string `json:"email"`
	Followers   int    `json:"followers"`
	Following   int    `json:"following"`
	PublicRepos int    `json:"public_repos"`
}

type Repo struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	FullName        string    `json:"full_name"`
	Description     string    `json:"description"`
	HTMLURL         string    `json:"html_url"`
	Language        string    `json:"language"`
	StargazersCount int       `json:"stargazers_count"`
	ForksCount      int       `json:"forks_count"`
	Fork            bool      `json:"fork"`
	PushedAt        time.Time `json:"pushed_at"`
	License         *struct {
		SPDXID string `json:"spdx_id"`
		Name   string `json:"name"`
	} `json:"license"`
}

func (r Repo) LicenseName() string {
	if r.License == nil {
		return ""
	}
	if r.License.SPDXID != "" {
		return r.License.SPDXID
	}
	return r.License.Name
}

type ContributionDay struct {
	Date  string `json:"date"`
	Count int    `json:"contributionCount"`
}
