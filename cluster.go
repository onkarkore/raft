/*
	Author : Onkar Kore
	This is a cluster interface to create differenet 
	server and send and recive meassages between them.
*/

package cluster

import (
	zmq "github.com/pebbe/zmq4"
	"fmt"
	"os"
	"strconv"
	"bufio"
	"io"
	"strings"	
	"encoding/gob"
	"bytes"
)

var (
	hostaddress []string
	configFile string
)



const (
	BROADCAST = -1
)


/* 
   Message format 
	RPid  - id of receiver 
	MsgId - unique message id (optional)
	Msg   - Actual data
*/
type Envelope struct {
	RPid int
	MsgId int64
	Msg string
	Term int
}


/*
   Server Data format
	ServerSocket - New Socket for server
	ServerID     - This server id
	ServerAdd    - This server address
	PeersId []   - This server peers
	Outboxd      - This server outbox
	Inboxd 	     - This server inbox
*/
type ServerData struct {
	ServerSocket *zmq.Socket	
	ServerID int
	ServerAdd string
	PeersId [] int
	PeersAdd [] string
	Outboxd chan *Envelope	
	Inboxd chan *Envelope
	ClientSocket [] *zmq.Socket
	servermapping map[int]int
}


type Server interface {
	Pid() int
	Peers() []int
	Outbox() chan *Envelope
	Inbox() chan *Envelope
	IndexOfServer() int
}


func (e ServerData) IndexOfServer() int {
	return e.servermapping[e.ServerID]
}


/* Returns Pid of this server */
func (e ServerData) Pid() int {
	return e.ServerID
}


func PeersAddress() []string {
	return hostaddress
}

/* Returns Peers of this server */
func (ser ServerData) Peers() []int {
	var av = []int{}

	f, err := os.OpenFile(configFile, os.O_CREATE|os.O_RDONLY,0600)
        if err != nil {
    	    print_error(err)
    	}
    	bf := bufio.NewReader(f)
	count:=0
    	for {
        	switch line, err := bf.ReadString('\n'); err {
        	case nil:
			line = line[:len(line)-1]

			if !strings.Contains(line,"tcp") {
				continue
			}

			parts := strings.Split(line, ":")
			value,_:=strconv.Atoi(parts[2])
			av = append(av,value)
			hostaddress = append(hostaddress,parts[0]+":"+parts[1])
			count++
        	case io.EOF:
        	    if line > "" {
        	        fmt.Println(line)
        	    }
		    f.Close()
        	    return av
	        default:
	           	print_error(err)
        	}
    	}
	f.Close()
	return av
}


/* Returns Outbox of this server */
func (s ServerData) Outbox() chan *Envelope {	
	return s.Outboxd
}


/* Returns inbox of this server */
func (s ServerData) Inbox() chan *Envelope {
	return s.Inboxd
}


/* Returns new socket of this server */
func CreateSocket() *zmq.Socket{
	var server *zmq.Socket
	server, _ = zmq.NewSocket(zmq.PULL)
	return server
}


/* Returns new socket of this server */
func CreateClientSocket() *zmq.Socket{
	client,_ := zmq.NewSocket(zmq.PUSH)
	return client
}


func TotalNumberOfServers(cfile string) int {
	
	configFile = cfile
	var s  ServerData
	var Peers = s.Peers()
	return len(Peers)
}

/* Create n number of server objects and initialize its properties */
func CreateServer(cfile string) [] ServerData{

	configFile = cfile	
	
	var serv [] ServerData
	var s  ServerData
	var Peers = s.Peers()
	var hostaddr =  PeersAddress()	
		
	for num:=0;num<1;num++{		
		s.ServerID = Peers[num]
		s.ServerAdd = hostaddr[num]+":"+strconv.Itoa(Peers[num])
		s.PeersId = Peers

		s.ServerSocket =  CreateSocket()
		s.ServerSocket.Bind(s.ServerAdd)

		s.Outboxd = make(chan * Envelope)	
		s.Inboxd = make(chan * Envelope)

		s.servermapping = make(map[int]int)

		for i:=0;i<len(Peers);i++{
			tmp:=CreateClientSocket()
			tmp.Connect(hostaddr[num]+":"+strconv.Itoa(Peers[i]))
			s.ClientSocket = append(s.ClientSocket,tmp)
			s.servermapping[Peers[i]]=i
		}
		serv = append(serv,s)
		fmt.Println("I: echo service is ready at ", s.ServerAdd)	
	}
	return serv
}

/* Receive messages from other server and send back to sender */
func ReceiveMsg(inbox chan *Envelope,server2 ServerData){
	
	for  {			
		receivemsg, err := server2.ServerSocket.RecvBytes(0)
		print_error(err)
		var r Envelope
		
		r1:=bytes.NewBuffer(receivemsg)
		decoder:=gob.NewDecoder(r1)
		decoder.Decode(&r)


		go addinbox(server2,r)
		if err != nil {
			print_error(err)
			break 
		}
		server2.ServerSocket.SendMessage(&r)
	}
}

/* Fill inbox of this server */
func addinbox(server2 ServerData,e Envelope){
	server2.Inbox() <- &e
}


/* Send message to other servers */
func SendMsgtoServers(outbox chan *Envelope,server1 ServerData){
	sentmsg_closed := false
	for {
	        if (sentmsg_closed) { return }
		select {
        		case cakeName, strbry_ok := <-outbox:
            			if (!strbry_ok) {
			                sentmsg_closed = true
			                fmt.Println(" Outbox channel closed!")	
			        } else {
										
              				var e Envelope
					e=*cakeName
					
					if e.RPid==-1 {
						var peers = server1.PeersId

						for peers_count:=0;peers_count<len(peers);peers_count++{
							if peers[peers_count]==server1.ServerID{
								continue
							}
							var e1 Envelope
							e1.RPid = peers[peers_count]
							e1.Msg  = e.Msg
							e1.Term = e.Term
					
					
							w:=new(bytes.Buffer)
							encoder:=gob.NewEncoder(w)
							encoder.Encode(e1)

							server1.ClientSocket[peers_count].SendBytes(w.Bytes(),0)	
						}
		
					} else {
						if e.RPid==server1.ServerID{
						} else {
							w:=new(bytes.Buffer)
							encoder:=gob.NewEncoder(w)
							encoder.Encode(e)
							server1.ClientSocket[server1.servermapping[e.RPid]].SendBytes(w.Bytes(),0)
						}
				}				        	            	    	
			}   
		 	
		}   
    	}   				
}

func print_error(e error) {
	if e != nil {
		panic(e)
	}
}


