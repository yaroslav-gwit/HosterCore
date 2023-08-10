package main

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
)

type NodeStruct struct {
	Hostname string `json:"hostname"`
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type HosterHaNodeStruct struct {
	IsManager   bool       `json:"is_manager"`
	IsCandidate bool       `json:"is_candidate"`
	IsWorker    bool       `json:"is_worker"`
	LastPing    int64      `json:"last_ping"`
	NodeInfo    NodeStruct `json:"node_info"`
}

type HaConfigJsonStruct struct {
	NodeType   string       `json:"node_type"`
	Candidates []NodeStruct `json:"candidates"`
}

var haHostsDb []HosterHaNodeStruct

func handleHaManagerRegistration(fiberContext *fiber.Ctx) error {
	tagCustomError = ""

	hosterNode := NodeStruct{}
	if err := fiberContext.BodyParser(hosterNode); err != nil {
		tagCustomError = err.Error()
		fiberContext.Status(fiber.StatusUnprocessableEntity)
		return fiberContext.JSON(fiber.Map{"error": err.Error()})
	}

	hosterHaNode := HosterHaNodeStruct{}
	hosterHaNode.IsManager = false
	hosterHaNode.NodeInfo = hosterNode
	haHostsDb = append(haHostsDb, hosterHaNode)

	jsonOutput, _ := json.Marshal(haHostsDb)
	return fiberContext.JSON(fiber.Map{"message": "done", "context": jsonOutput})
}

func handleHaPing(fiberContext *fiber.Ctx) error {
	return fiberContext.JSON(fiber.Map{"message": "pong"})
}

func handleHaTerminate(fiberContext *fiber.Ctx) error {
	return fiberContext.JSON(fiber.Map{"message": "done"})
}

func handleHaPromote(fiberContext *fiber.Ctx) error {
	return fiberContext.JSON(fiber.Map{"message": "done"})
}

func handleHaShareWorkers(fiberContext *fiber.Ctx) error {
	return fiberContext.JSON(fiber.Map{"message": "done"})
}

func handleHaShareManager(fiberContext *fiber.Ctx) error {
	return fiberContext.JSON(fiber.Map{"message": "done"})
}

func handleHaShareCandidates(fiberContext *fiber.Ctx) error {
	return fiberContext.JSON(fiber.Map{"message": "done"})
}
