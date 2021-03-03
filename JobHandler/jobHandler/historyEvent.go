package jobHandler

type HistoryEvent struct {
	NumWorkers uint
	WorkerId   int
	Loss       float64
	Time       float64
	Epoch      int
}

type FunctionResponse struct {
	Loss     float64 `json:"loss"`
	Accuracy float64 `json:"accuracy"`
	WorkerId int     `json:"worker_id"`
}
