Raft - Leader Election Implementation 

-----------------------------------------------------------------------------------------------------------

1. What is this?

	This project contains a library called "raft.go" which is used for leader selection in a cluster. 

2. Files present

	(a) cluster.go 
	
		This is a library which is used for
		 - To create n-number of servers 
		 - To queue outgoing message in the outbox of each server
		 - To queue incoming message in the inbox of each server
		 
	(b) raft.go

		This is a raft library file.
		Following things are to be done in this interface
		 - Create new socket for each server
		 - Set all properties related to each server
		 - Set raft structure for each server
		 - Send request, reply and heartbeat messages between servers.
		 - Select a leader based on maximum number of votes.

	(c) raft_test.go
	
		This is a test file to check working of raft library.
		Following test cases are checked
		 Test case 1 - Start cluster. One of the server become candidate and request for votes.
			       When that server receives votes greater than (total no of servers/2), t will
			       become leader of that cluster.
			       Then kill that leader after some time by stopping to send heartbeat messages
			       and reply messages (accepted vote message) for other candidates.
			       This process continues till half of the server dead then new candidate is ready
			       to become server but it will not get required number of votes.
		 Test case 2 - Split leader selection after killing half of the server.
			       Two servers become candidate at the same time but half of the servers are already
			       dead so no one become leader.

	(d) cluster.conf
	
		This file contains list of all sever addresses.
		e.g. tcp://127.0.0.1:2001
		     tcp://127.0.0.1:2002
		Also contain Heartbeat and Election timouts.


3. Log files

	log folder is created at the start which contain last term of each server.
	currentLeader file contains id of current leader of a cluster.
		
4. How to run?
	- go get github.com/onkarkore/raft/
	- go test github.com/onkarkore/raft/

5. References 

	- go language tutorial
	- raft paper
