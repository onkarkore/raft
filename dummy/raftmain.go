/*
	Author : Onkar Kore
	This is a main program to create multiple server and test raft interface.
*/

package main

import (
	"os"
	"time"
	raft "github.com/onkarkore/raft"
	"net/rpc"
	"net/http"
)

type Args struct {
}

type Results struct {
	Term int
	CurrentState int
}

type RPCMethods struct {
  
}

var r raft.RaftData

/*
	Create a cluster and start election.
	After each leader selection, kill that leader then new server become a candidate
	and start election. After killing half of the server, there is no leader.	
*/
func main() {

	rpc.Register(&RPCMethods{})
	rpc.HandleHTTP()
	go http.ListenAndServe(":"+os.Args[2], nil)
	
	StartRaftTesting()
}



func StartRaftTesting() {

	raft.AllocateRaft(os.Args[1],&r)
		
	for{	
		select {
			case <-time.After(5 * time.Second):
			
		}
	}
}


func (m *RPCMethods) GetTerm (args Args, results *Results) error {
	results.Term = r.Term
	results.CurrentState = r.ServerType
	return nil  // no error
}




