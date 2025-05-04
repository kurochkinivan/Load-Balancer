package entity

type Client struct {
	ID            int64  `json:"id"`
	IPAddress     string `json:"ip_address"`
	Name          string `json:"name"`
	Capacity      int    `json:"capacity"`
	RatePerSecond int    `json:"rate_per_second"`
}
