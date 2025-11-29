package models

type Wheel struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Items []WheelItem `json:"items"`
}

type WheelItem struct {
	ID     string  `json:"id"`
	Option string  `json:"option"`
	Color  string  `json:"color"`
	Weight float64 `json:"weight"`
}
