/*
	Author : Onkar Kore
	This is a raft interface for leader election.
*/

package cluster

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	s                []ServerData
	
)

/*
   raft Data format
	Term 		- Current term of each server
	IsPrimary    	- true if this server is leader else false
	ServerNumber    - address of this server
	LastVotedTo   	- address of sever last voted by this server
	NumberOfVotes   - Number of votes received by this server
	IsVoted		- true if server voted to candidate else false
	ServerType	- 1 leader, 2 candidate, 3 follower	
*/
type RaftData struct {
	Term          int
	IsPrimary     bool
	ServerNumber  int
	LastVotedTo   int
	NumberOfVotes int	
	IsVoted       bool
	ServerType    int 
	server ServerData
	ElectionTimeOut  time.Duration /* Election Timeout */
	HeartbeatTimeOut time.Duration /* Heartbeat Timeout */
	receivedVoteFrom map[int]bool 
}

/*
	This function is used to create cluster and allocate 
	term to each server.
*/
func AllocateRaft(conFile string,r *RaftData)  {

	/* Create cluster */
	s = CreateServer(conFile)
	configFile = conFile
	t:=readTimeoutsFromFile()	
	totalServers:=TotalNumberOfServers(conFile)
	

	/* Initialize raft data structure of each server */
	for i := 0; i < 1; i++ {
		r.Term = 0
		r.IsPrimary = false
		r.ServerNumber = s[i].ServerID
		r.LastVotedTo = 0
		r.NumberOfVotes = 1
		r.ServerType = 3
		r.server = s[i]
		r.IsVoted = false
		r.receivedVoteFrom = make (map[int]bool)
		r.ElectionTimeOut = t.ElectionTimeOut
		r.HeartbeatTimeOut = t.HeartbeatTimeOut 
	}

	
	/* Start send and receive message service of each server */
	for num := 0; num < 1; num++ {
		go SendMsgtoServers(s[num].Outbox(), s[num])
		go ReceiveMsg(s[num].Inbox(), s[num])
	}

	go raftWorking(r, totalServers)
}



/*
	This function is used to read last term of each server while 
	booting at the start if present. 
*/
func initializeFromFile(r *RaftData)  {	
	var e error

	_, e = os.Stat("log")
	if e != nil {
		os.Mkdir("log", 0777)
	}

	lfile := "log" + "/log_for_servernumber_" + strconv.Itoa(r.ServerNumber)
	if _, e = os.Stat(lfile); e != nil {
		_, err := os.Create(lfile)
		print_error(err)
	}

	logfile, err := os.OpenFile(lfile, os.O_RDWR, 0600)
	print_error(err)

	buf := bufio.NewReader(logfile)
	if err != nil {
		
	} else {
		for {

			line, err := buf.ReadString('\n')
			if err != nil {
				break
			}
			line = line[:len(line)-1]
			if strings.Contains(line, "TERM") {
				continue
			}

			parts := strings.Split(line, " ")

			r.Term, e = strconv.Atoi(parts[0])
			print_error(e)

			leader,e := strconv.Atoi(parts[1])
			print_error(e)
			if leader == r.ServerNumber {
				r.IsPrimary = true
				r.ServerType = 1
			}
		}
	}
	logfile.Close()
}

/*
	Read HeartBeat and Election Timeout from cluster.conf file
*/
func readTimeoutsFromFile() RaftData{

	var t RaftData

	configfile, err := os.OpenFile(configFile, os.O_RDWR, 0600)
	print_error(err)

	buf := bufio.NewReader(configfile)
	if err != nil {
		
	} else {
		for {
			line, err := buf.ReadString('\n')
			if err != nil {
				break
			}
			line = line[:len(line)-1]
			if strings.Contains(line, "HeartBeatTimeOut") {
				parts := strings.Split(line, " ")
				value, _ := strconv.Atoi(parts[1])
				t.HeartbeatTimeOut = time.Duration(value)
			}
			if strings.Contains(line, "ElectionTimeOut") {
				parts := strings.Split(line, " ")
				value, _ := strconv.Atoi(parts[1])
				t.ElectionTimeOut = time.Duration(value)
			}
		}
	}
	configfile.Close()
	return t
}

/*
	This fuction is used to select leader of a cluster.
*/
func raftWorking(r *RaftData, totalServer int) {

	initializeFromFile(r)
	for {
		/* This is the working mechanism for leader */
		if r.IsPrimary {
			select {
			case msg, ok := <-r.server.Inbox():
				if !ok {
				} else {
					var m Envelope
					m = *msg
					str := strings.Split(m.Msg, " ")

					if str[0] == "Request" {
						if r.Term < m.Term {
							r.Term = m.Term
							r.IsPrimary = false
							r.ServerType = 3
							r.IsVoted = true
							go sentReply(r, str[1])
						}
					}

					if str[0] == "HeartBeat" {
						if r.Term < m.Term {
							r.Term = m.Term
							r.IsPrimary = false
							r.IsVoted = false
							r.ServerType = 3
						}
					}

				}
			/* Leader send heartbeat messgaes to followers after each HeartBeatTime Interval */
			case <-time.After(r.HeartbeatTimeOut * time.Millisecond):
				go sentHeartbeatMessage(r)
			}
		}

		/* This is the working mechanism for followers */
		if !r.IsPrimary {
			select {
			case msg, ok := <-r.server.Inbox():
				if !ok {
				} else {
					var m Envelope
					m = *msg
					str := strings.Split(m.Msg, " ")

					/* If message is of type Request then check term of this server with sender term and reply back 						   if this server term is less than sender's term */
					if str[0] == "Request" {
						if r.Term < m.Term && !r.IsVoted {
							r.Term = m.Term
							r.LastVotedTo,_ = strconv.Atoi(str[1])
							r.IsVoted = true
							go sentReply(r, str[1])
						}
					}

					/* If message type is Reply then calculate number of votes and if number of votes greater than 							(number of servers/2) then assign this server as a leader */
					if str[0] == "Reply" {
					
						rid,_ := strconv.Atoi(str[1])
						if !r.receivedVoteFrom[rid] {
							r.NumberOfVotes = r.NumberOfVotes + 1
							r.receivedVoteFrom[rid] = true
						}
						
						if (totalServer%2==1 && r.NumberOfVotes > (totalServer/2)) || (totalServer%2==0 && r.NumberOfVotes > (totalServer/2)-1) {
							r.IsPrimary = true
							r.ServerType = 1

							lfile := "log" + "/log_for_servernumber_" + strconv.Itoa(r.ServerNumber)
							logfile, _ := os.OpenFile(lfile, os.O_RDWR, 0600)
							logfile.Write([]byte(strconv.Itoa(r.Term)+" "+strconv.Itoa(r.ServerNumber)+ "\n"))
							logfile.Close()

						}
					}
					if str[0] == "HeartBeat" {
						r.IsVoted = false
						////fmt.Println("OK")						
					}
				}
			/* One of the server become candidate after electointimeout and requesr for votes */
			case <-time.After(r.ElectionTimeOut * time.Millisecond):
					r.Term = r.Term + 1
					r.LastVotedTo = r.ServerNumber
					r.NumberOfVotes = 1
					r.ServerType = 2
					r.IsVoted = false
					for key, _ := range r.receivedVoteFrom { 
						delete(r.receivedVoteFrom,key)
					}

					go sentRequest(r)
			}
		}
	}
}

/*
	This function is used to send request message like vote me.
*/
func sentRequest(r *RaftData) {
	var e Envelope
	e.RPid = -1
	e.Term = r.Term
	e.Msg = "Request " + strconv.Itoa(r.ServerNumber)
	r.server.Outbox() <- &e
}

/*
	This function is used to send reply message like vote is granted.
*/
func sentReply(r *RaftData, toAdd string) {

	id, _ := strconv.Atoi(toAdd)
	var e Envelope
	e.RPid = id
	e.Term = r.Term
	e.Msg = "Reply " + strconv.Itoa(r.ServerNumber)
	r.server.Outbox() <- &e
}

/*
	This function is used by leader to send heartbeat messages.
*/
func sentHeartbeatMessage(r *RaftData) {
	var e Envelope
	e.RPid = -1
	e.Term = r.Term
	e.Msg = "HeartBeat " + strconv.Itoa(r.ServerNumber)
	r.server.Outbox() <- &e
}


