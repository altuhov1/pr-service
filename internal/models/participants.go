package models

import "time"

type PRStatus string

type User struct {
	UserIds  string `json:"userIds"`
	IsActive bool   `json:"isActive"`
}
type Team struct {
	Name  string   `json:"name"`
	Users []string `json:"UsersId"`
}
type PullRequest struct {
	Status        string    `json:"status"`
	InspectorsIds []string  `json:"inspectorsIds"`
	IsMerged      bool      `json:"isMerged"`
	MergedAt      time.Time `json:"mergedAt"`
	CreatedAt     time.Time `json:"createdAt"`
}
