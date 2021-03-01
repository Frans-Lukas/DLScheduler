package structs

type Job struct {
	Budget     float64 `json:"budget"`
	TargetLoss float64 `json:"targetLoss"`
	ImageUrl   string `json:"imageUrl"`
	History    []HistoryEvent
}