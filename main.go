package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type Peer struct {
	ID   string
	Addr string
	Port string
}

var peers = map[string]Peer{
	"node-01": {
		ID:   "node-01",
		Addr: "http://localhost:6001",
		Port: "6001",
	},
	"node-02": {
		ID:   "node-02",
		Addr: "http://localhost:6002",
		Port: "6002",
	},
	"node-03": {
		ID:   "node-03",
		Addr: "http://localhost:6003",
		Port: "6003",
	},
	"node-04": {
		ID:   "node-04",
		Addr: "http://localhost:6004",
		Port: "6004",
	},
}

var nodeID string
var leaderID string

func getNodeID() string {

	if len(os.Args) < 2 {
		panic(errors.New("not enough arguments"))
	}

	nodeID := os.Args[1]

	if _, ok := peers[nodeID]; !ok {
		panic(errors.New("invalid node-id"))
	}

	return nodeID

}

func main() {

	nodeID = getNodeID()

	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	/*
			    1. ping
		    2. electation
		    3. leader-elected

	*/

	r.POST("/ping", func(ctx *gin.Context) {
		ctx.JSONP(http.StatusOK, gin.H{})
	})

	r.POST("/electation", func(ctx *gin.Context) {

		ctx.JSONP(http.StatusOK, gin.H{})

		go elect()

	})

	r.POST("/leader-elected", func(ctx *gin.Context) {

		var data map[string]string

		if err := ctx.ShouldBindJSON(&data); err != nil {
			return
		}

		li, ok := data["leaderID"]
		if !ok {
			fmt.Println("expected leader in request body")
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "expected leader in request body",
			})
		}

		leaderID = li

		fmt.Println("leader elected ", leaderID)

		ctx.JSONP(http.StatusOK, gin.H{})

		go pingContinuslyLeader()

	})

	go r.Run(fmt.Sprintf(":%s", peers[nodeID].Port))

	waitAllPeersToBeUp()

	elect()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	<-c

}

func compunicateWithPeer(url string, payload map[string]string) error {

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("revived non-ok status code")
	}

	return nil
}

func isRankHigher(ourId string, nodeID string) bool {
	return strings.Compare(nodeID, ourId) == 1
}

func elect() {

	isHigherRankNodeAvailable := false

	for _, peer := range peers {

		if !isRankHigher(nodeID, peer.ID) {
			continue
		}

		err := compunicateWithPeer(fmt.Sprintf("%s/electation", peer.Addr), map[string]string{})
		if err != nil {
			continue
		}

		isHigherRankNodeAvailable = true

	}

	if !isHigherRankNodeAvailable {

		fmt.Println("I'm the leader: ", nodeID)

		for _, peer := range peers {

			if peer.ID == nodeID {
				continue
			}

			_ = compunicateWithPeer(fmt.Sprintf("%s/leader-elected", peer.Addr), map[string]string{
				"leaderID": nodeID,
			})

		}
	}

}

func pingContinuslyLeader() {

	for {

		err := compunicateWithPeer(fmt.Sprintf("%s/ping", peers[leaderID].Addr), map[string]string{})
		if err != nil {
			// leader is down
			elect()
			break
		}

		time.Sleep(10 * time.Second)
	}
}

func waitAllPeersToBeUp() {

	for _, peer := range peers {

		for {
			err := compunicateWithPeer(fmt.Sprintf("%s/ping", peer.Addr), map[string]string{})
			if err == nil {
				break
			}
		}
	}
}
