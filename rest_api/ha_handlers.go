package main

import (
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
var haChannelAdd = make(chan HosterHaNodeStruct, 20)
var haChannelRemove = make(chan HosterHaNodeStruct, 20)

func init() {
	go addHaNode(haChannelAdd)
	go removeHaNode(haChannelRemove)
}

func addHaNode(haChannelAdd chan HosterHaNodeStruct) {
	for msg := range haChannelAdd {
		haHostsDb = append(haHostsDb, msg)
	}
}

func removeHaNode(haChannelRemove chan HosterHaNodeStruct) {
	haHosts := []HosterHaNodeStruct{}
	for msg := range haChannelAdd {
		for _, v := range haHostsDb {
			if msg != v {
				haHosts = append(haHosts, v)
			}
		}
	}
	haHostsDb = haHosts
}

func handleHaManagerRegistration(fiberContext *fiber.Ctx) error {
	tagCustomError = ""

	hosterNode := new(NodeStruct)
	if err := fiberContext.BodyParser(hosterNode); err != nil {
		tagCustomError = err.Error()
		fiberContext.Status(fiber.StatusUnprocessableEntity)
		return fiberContext.JSON(fiber.Map{"error": err.Error()})
	}

	hosterHaNode := HosterHaNodeStruct{}
	hosterHaNode.IsManager = false
	hosterHaNode.IsCandidate = false
	hosterHaNode.IsWorker = true
	hosterHaNode.NodeInfo = *hosterNode
	haChannelAdd <- hosterHaNode

	return fiberContext.JSON(fiber.Map{"message": "done", "context": hosterHaNode})
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
	workers := []HosterHaNodeStruct{}
	for _, v := range haHostsDb {
		if v.IsWorker {
			workers = append(workers, v)
		}
	}
	return fiberContext.JSON(workers)
}

func handleHaShareManager(fiberContext *fiber.Ctx) error {
	manager := HosterHaNodeStruct{}
	for _, v := range haHostsDb {
		if v.IsManager {
			manager = v
		}
	}
	return fiberContext.JSON(manager)
}

func handleHaShareCandidates(fiberContext *fiber.Ctx) error {
	candidates := []HosterHaNodeStruct{}
	for _, v := range haHostsDb {
		if v.IsCandidate {
			candidates = append(candidates, v)
		}
	}
	return fiberContext.JSON(candidates)
}
