package jobHandler

type HistoryEvent struct {
	NumWorkers uint
	NumServers uint
	WorkerId   int
	Loss       float64
	Accuracy   float64
	Time       float64
	Epoch      int
	Cost       float64
}

type FunctionResponse struct {
	Loss     float64 `json:"loss"`
	Accuracy float64 `json:"accuracy"`
	WorkerId int     `json:"worker_id"`
}
