Raft - Leader Election Implementation 

-----------------------------------------------------------------------------------------------------------

1. What is this?

	This project is used to create n servers and implemented raft algorithm on the top of it to 
	select a leader	for a cluster of n servers.
	
2. Working

	Cluster library which is used for
	
		 - To create n-number of servers
		 - To queue outgoing message in the outbox of each server
		 - To queue incoming message in the inbox of each server
		 
	Raft library which is used for leader selection based on maximum number of votes by peers.


3. How to run?

		- go get github.com/onkarkore/raft/
		- build /dummy/raftmain.go 
			go build raftmain.go
		- run python script to create config file for each server
			python configcreator.py <HearbeatTImeout> <EectionTimeout>
		- go test github.com/onkarkore/raft/


4. Testing
	
	Following test cases are checked
		 
		 Test case 1 - (Minority Failure)
			       > Start cluster. One of the server become candidate and request for votes.
			         When that server receives votes greater than (total no of servers/2), it will
			         become leader of that cluster.
			       > Then kill that leader after some then after election timeout new server become 
			         candidate and request for votes.
			       
		 Test case 2 - (Majority Failure)
			       > After killing more than half of server if total number of servers are odd or 
				 half of the servers if total number of servers are even then there is no leader 
				 for current cluster.
	
		Also, ensure that there is only one leader present at a time.

5. Configuration File (cluster.conf)

		This file contains list of all sever addresses.
		e.g. tcp://127.0.0.1:2001
		     tcp://127.0.0.1:2002

6. Log files

	log folder is created at the start for each server which contain term and leader server number.
	
		
7. References 

	- go language tutorial
	- raft paper
	- http://golang.org/pkg/net/rpc/

