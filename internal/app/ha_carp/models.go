package main

type CarpInfo struct {
	Interface string `json:"interface"`
	Vhid      int    `json:"vhid"`
	Advskew   int    `json:"advskew"`
	Advbase   int    `json:"advbase"`
	Pass      string `json:"pass"`
}
