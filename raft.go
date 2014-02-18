/*
	Author : Onkar Kore
	This is a raft interface for leader election.
*/

package cluster

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	s                [30]ServerData
	ElectionTimeOut  time.Duration /* Election Timeout */
	HeartbeatTimeOut time.Duration /* Heartbeat Timeout */
	IsLeaderSelected bool
	skip             int
	term             int
	leader           int
	IsVoteRequested  bool
)

/*
   raft Data format
	Term 		- Current term of each server
	IsPrimary    	- true if this server is leader else false
	ServerNumber    - address of this server
	LastVotedTo   	- address of sever last voted by this server
	NumberOfVotes   - Number of votes received by this server
	SentHeartbeat 	- Testing flag to start or stop heartbeating message
*/
type raftData struct {
	Term          int
	IsPrimary     bool
	ServerNumber  int
	LastVotedTo   int
	NumberOfVotes int
	SentHeartbeat bool
}

/*
	This function is used to create cluster and allocate 
	term to each server.
*/
func AllocateRaft() [30]raftData {

	/* Create cluster */
	s = CreateConnection()

	ReadTimeoutsFromFile()

	var r [30]raftData
	var sd ServerData
	var Peers = sd.Peers()

	IsLeaderSelected = false
	skip = 0
	term = 0
	IsVoteRequested = true

	/* Initialize raft data structure of each server */
	for i := 0; i < len(Peers); i++ {
		r[i].Term = 0
		r[i].IsPrimary = false
		r[i].ServerNumber = s[i].ServerID
		r[i].LastVotedTo = 0
		r[i].NumberOfVotes = 1
		r[i].SentHeartbeat = true

	}

	/* Read last status of leader while initial booting from file if present. */
	var e error
	lfile := "currentLeader"
	if _, e = os.Stat(lfile); e != nil {
		_, err := os.Create(lfile)
		print_error(err)
	}

	logfile, err := os.OpenFile(lfile, os.O_RDWR, 0600)
	print_error(err)

	buf := bufio.NewReader(logfile)
	if e != nil {
	} else {
		for {

			line, err := buf.ReadString('\n')
			if err != nil {
				break
			}
			line = line[:len(line)-1]
			if strings.Contains(line, "LEADER") {
				continue
			}
			leader, e = strconv.Atoi(line)
			print_error(e)
		}
	}

	/* Start send and receive message service of each cluster */
	for num := 0; num < len(Peers); num++ {
		go SendMsgtoServers(s[num].Outbox(), s[num])
		go ReceiveMsg(s[num].Inbox(), s[num])
	}

	return r
}

/*
	This function is used to read last term of each server while 
	booting at the start if present. 
*/
func InitializeFromFile(r raftData) {
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
	if e != nil {
		//logfile.Write([]byte("TERM\n"))
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
			r.Term, e = strconv.Atoi(line)
			print_error(e)

			if r.ServerNumber == leader {
				r.IsPrimary = true
				IsLeaderSelected = true
				term = r.Term
			}
		}
	}
}

/*
	Read HeartBeat and Election Timeout from cluster.conf file
*/
func ReadTimeoutsFromFile() {

	configfile, err := os.OpenFile("cluster.conf", os.O_RDWR, 0600)
	print_error(err)

	var e error
	buf := bufio.NewReader(configfile)
	if e != nil {
		//logfile.Write([]byte("TERM\n"))
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
				HeartbeatTimeOut = time.Duration(value)
			}
			if strings.Contains(line, "ElectionTimeOut") {
				parts := strings.Split(line, " ")
				value, _ := strconv.Atoi(parts[1])
				ElectionTimeOut = time.Duration(value)
			}
		}
	}
}

/*
	This fuction is used to select leader of a cluster.
*/
func RaftWorking(r raftData, totalServer int) {

	r.Term = term
	if r.ServerNumber == leader {
		r.IsPrimary = true
		IsLeaderSelected = true
		fmt.Println("\n", "Leader server id ", r.ServerNumber, "Term of server", r.Term, "\n")
	}

	for {
		/* This is the working mechanism for leader */
		if r.IsPrimary {
			select {
			case msg, ok := <-s[r.ServerNumber-2011].Inbox():
				if !ok {
					fmt.Println("Inbox channel closed!")
				} else {
					var m Envelope
					m = *msg
					str := strings.Split(m.Msg, " ")

					if str[0] == "Request" {
						if r.Term < m.Term {
							r.Term = m.Term
							go SentReply(r, str[1])
						}
					}
				}
			/* Leader send heartbeat messgaes to followers after each HeartBeatTime Interval */
			case <-time.After(HeartbeatTimeOut * time.Millisecond):
				/* This part contains flags which are used to stop sending heartbeat messages
				   to get the effect like server is dead.					 
				*/
				skip = skip + 1
				if skip < 15 {

				} else {
					r.IsPrimary = false
					skip = 0
					r.SentHeartbeat = false
					IsVoteRequested = true
				}

				if r.SentHeartbeat {
					go SentHeartbeatMessage(r)
				}
			}
		}

		/* This is the working mechanism for followers */
		if !r.IsPrimary {
			select {
			case msg, ok := <-s[r.ServerNumber-2011].Inbox():
				if !ok {
					fmt.Println("Inbox channel closed!")
				} else {
					var m Envelope
					m = *msg
					str := strings.Split(m.Msg, " ")

					/* If message is of type Request then check term of this server with sender term and reply back 						   if this server term is less than sender's term */
					if str[0] == "Request" {
						if r.Term < m.Term && r.SentHeartbeat {
							r.Term = m.Term
							go SentReply(r, str[1])
						}
					}

					/* If message type is Reply then calculate number of votes and if number of votes greater than 							   (number of servers/2) then assign this server as a leader */
					if str[0] == "Reply" {
						r.NumberOfVotes = r.NumberOfVotes + 1
						if r.NumberOfVotes > (totalServer/2) && !IsLeaderSelected {
							r.IsPrimary = true
							IsLeaderSelected = true
							fmt.Println("Leader server id ", r.ServerNumber, "Term of server", r.Term, "\n")

							lfile := "log" + "/log_for_servernumber_" + strconv.Itoa(r.ServerNumber)
							logfile, _ := os.OpenFile(lfile, os.O_RDWR, 0600)
							logfile.Write([]byte(strconv.Itoa(r.Term) + "\n"))

							file, _ := os.OpenFile("currentLeader", os.O_RDWR, 0600)
							file.Write([]byte(strconv.Itoa(r.ServerNumber) + "\n"))
						}
					}
					if str[0] == "HeartBeat" {
						//fmt.Println("OK")						
					}
				}
			/* One of the server become candidate after electointimeout and requesr for votes */
			case <-time.After(ElectionTimeOut * time.Millisecond):
				IsLeaderSelected = false
				if r.SentHeartbeat && IsVoteRequested {
					r.Term = r.Term + 1
					r.LastVotedTo = r.ServerNumber
					r.NumberOfVotes = 1
					IsVoteRequested = false
					fmt.Println("Candidate server id ", r.ServerNumber)
					go SentRequest(r)
				}
			}
		}
	}
}

/*
	This function is used to send request message like vote me.
*/
func SentRequest(r raftData) {
	var e Envelope
	e.RPid = -1
	e.Term = r.Term
	e.Msg = "Request " + strconv.Itoa(r.ServerNumber)
	s[r.ServerNumber-2011].Outbox() <- &e
}

/*
	This function is used to send reply message like vote is granted.
*/
func SentReply(r raftData, toAdd string) {

	id, _ := strconv.Atoi(toAdd)
	var e Envelope
	e.RPid = id
	e.Term = r.Term
	e.Msg = "Reply " + strconv.Itoa(r.ServerNumber)
	s[r.ServerNumber-2011].Outbox() <- &e
}

/*
	This function is used by leader to send heartbeat messages.
*/
func SentHeartbeatMessage(r raftData) {
	var e Envelope
	e.RPid = -1
	e.Term = r.Term
	e.Msg = "HeartBeat " + strconv.Itoa(r.ServerNumber)
	s[r.ServerNumber-2011].Outbox() <- &e
}

func print_error(e error) {
	if e != nil {
		panic(e)
	}
}
