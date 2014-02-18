/*
	Author : Onkar Kore
	This is a main program to create multiple server and test raft interface.
*/

package cluster

import (
	"fmt"
	"os"
	"testing"
	"time"
)

/*
	Create a cluster and start election.
	After each leader selection, kill that leader then new server become a candidate
	and start election. After killing half of the server, there is no leader.	
*/
func Test_main(t *testing.T) {

	if len(os.Args) < 1 {
		fmt.Printf("<usage> : go run server.go \n")
		return
	}

	fmt.Println("\n\n", "Test case 1 - Kill each leader one by one after leader selection", "\n\n")

	StartRaftTesting()
}

/*
	Majority failure case
*/
func Test_main1(t *testing.T) {

	if len(os.Args) < 1 {
		fmt.Printf("<usage> : go run server.go \n")
		return
	}

	fmt.Println("\n\n", "Test case 2 - Split leader selection after killing half of the server", "\n\n")

	StartRaftTesting()
}

func StartRaftTesting() {

	var s ServerData
	var Peers = s.Peers()

	raft := AllocateRaft()

	for num := 0; num < len(Peers); num++ {
		InitializeFromFile(raft[num])
	}

	for num := 0; num < len(Peers); num++ {
		go RaftWorking(raft[num], len(Peers))
	}
	select {
	case <-time.After(15 * time.Second):
	}
}
