package main

type CarpInfo struct {
	Status    string `json:"status"`
	Interface string `json:"interface"`
	Vhid      int    `json:"vhid"`
	Advskew   int    `json:"advskew"`
	Advbase   int    `json:"advbase"`
	Pass      string `json:"pass"`
}
