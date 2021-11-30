package model

import "time"

type FinishedStages struct {
	Stages []Stage
}

type Stage struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Workflows []Workflow `json:"workflows"`
}

type Workflow struct {
	ExternalId string    `json:"external_id"`
	FinishedAt time.Time `json:"finished_at"`
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	StartedAt  time.Time `json:"started_at"`
	Status     string    `json:"status"`
}
