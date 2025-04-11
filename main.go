package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
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
	// "node-05": {
	// 	ID:   "node-05",
	// 	Addr: "http://localhost:6005",
	// 	Port: "6005",
	// }, "node-06": {
	// 	ID:   "node-06",
	// 	Addr: "http://localhost:6006",
	// 	Port: "6006",
	// },
}

func getNodeID() string {
	if len(os.Args) < 2 {
		panic(errors.New("missing node-id"))
	}
	id := os.Args[1]

	if _, ok := peers[id]; !ok {
		panic(errors.New("invalid node-id"))
	}

	return id
}

var nodeID string

func main() {

	nodeID = getNodeID()

	fmt.Println("Node ID: ", nodeID)

	و
	r := gin.New()

	r.POST("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})

	r.POST("/electation", func(c *gin.Context) {
		c.JSON(200, gin.H{})

		go elect("electation-request")
	})

	r.POST("/leader-elected", func(c *gin.Context) {

		var data map[string]string
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		leaderID, ok := data["leader"]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Leader ID not provided"})
			return
		}

		fmt.Printf("Leader elected: %s\n", leaderID)
		c.JSON(200, gin.H{})

		go pingContinuslyLeader(leaderID)

	})

	go r.Run(fmt.Sprintf(":%s", peers[nodeID].Port))

	waitAllPeersBeUp()

	elect("startup")

	c := make(chan bool)
	<-c
}

func waitAllPeersBeUp() {

	for peerID, peer := range peers {

		if peerID == nodeID {
			continue
		}

		err := compunicateWithPeer(fmt.Sprintf("%s/ping", peer.Addr), map[string]string{})

		for err != nil {
			time.Sleep(1 * time.Second)
			err = compunicateWithPeer(fmt.Sprintf("%s/ping", peer.Addr), map[string]string{})
		}

		fmt.Println("ping received from peer: ", peer.ID)

	}

}

func compunicateWithPeer(url string, data map[string]string) error {

	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Create an HTTP client with a timeout
	client := &http.Client{
		Timeout: 5 * time.Second, // Set timeout to 5 seconds
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response: %d", resp.StatusCode)
	}

	return nil
}

func elect(from string) {
	fmt.Sprintln("start electation")
	isHighestRankedNodeAvailable := false
	for _, peer := range peers {

		if !isRankHigher(nodeID, peer.ID) {
			continue
		}

		err := compunicateWithPeer(fmt.Sprintf("%s/electation", peer.Addr), map[string]string{})
		if err == nil {
			isHighestRankedNodeAvailable = true
		}
	}

	if !isHighestRankedNodeAvailable {

		fmt.Printf("I am the leader: %s %s\n", nodeID, from)
		for _, peer := range peers {

			if peer.ID == nodeID {
				continue
			}

			_ = compunicateWithPeer(fmt.Sprintf("%s/leader-elected", peer.Addr), map[string]string{
				"leader": nodeID,
			})

		}
	}

}

func isRankHigher(myNodeID, id string) bool {
	return strings.Compare(id, myNodeID) == 1
}

func pingContinuslyLeader(leaderID string) {

	for {
		err := compunicateWithPeer(fmt.Sprintf("%s/ping", peers[leaderID].Addr), map[string]string{})
		if err != nil {
			break
		}
		time.Sleep(5 * time.Second)
	}
	fmt.Printf("Leader %s is down\n", leaderID)
	elect("leader-down")
}
