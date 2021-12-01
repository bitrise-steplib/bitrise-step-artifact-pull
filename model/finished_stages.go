package model

type FinishedStages []Stage

type Stage struct {
	Name      string     `json:"name"`
	Workflows []Workflow `json:"workflows"`
}

type Workflow struct {
	ExternalId string `json:"external_id"`
	Name       string `json:"name"`
}
