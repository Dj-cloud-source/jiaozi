package station

type StationResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type AdminStationResponse struct {
	ID     uint64 `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}
