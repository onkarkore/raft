/*
	Author : Onkar Kore
	This is a test program for testing raft packege.
*/

package cluster

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"net/rpc"
	"time"
	"testing"
	"bufio"
	"strings"
)


type Args struct {
}

type Results struct {
	Term int
	CurrentState int
}

/*
	Create a cluster and start election.
	After each leader selection, kill that leader then new server become a candidate
	and start election. After killing half of the server, there is no leader.	
*/
func Test_main(t *testing.T) {

	test:=1
	test1:=0

	var args Args
	var serverport[] int

	/* Read config file for number of servers. */
	configfile, _ := os.OpenFile("cluster.conf", os.O_RDWR, 0600)
	i:=0
	var e error
	buf := bufio.NewReader(configfile)
	if e != nil {
		
	} else {
		for {
			line, err := buf.ReadString('\n')
			if err != nil {
				break
			}
			line = line[:len(line)-1]
			parts := strings.Split(line, ":")
			value,_ := strconv.Atoi(parts[2])
			serverport=append(serverport,value)
			i = i+1
		}
	}
	configfile.Close()
	
	totalnumber:=i
	totalkilled := 0
	leadercount := 0

	var cmd [] *exec.Cmd
	cmd = make([]*exec.Cmd,totalnumber) 	
	
	port:=1234

	/* Create n processes for n servers */
	for i:=0;i<totalnumber;i++{
		filename := "cluster"+strconv.Itoa(serverport[i])+".conf"
		cmd[i] = exec.Command("./dummy/raftmain", filename, strconv.Itoa(port))
		cmd[i].Stdout = os.Stdout
		cmd[i].Stderr = os.Stderr	
		cmd[i].Start()	
		port=port+1
	}

	var client[] *rpc.Client
	var results[] Results
	var killed[] int
	var rpcport[] int
		
	client = make ([]*rpc.Client,totalnumber)
	results = make ([]Results,totalnumber)
	rpcport = make ([]int,totalnumber)
	port=1234
	
	time.Sleep(3*time.Second)

	/* Create rpc sockets for n servers */
	for i1:=0;i1<totalnumber;i1++{	
		var err error			
		client[i1], err = rpc.DialHTTP("tcp", "127.0.0.1:"+strconv.Itoa(port))
		if err!=nil{
			fmt.Println("Error",err)			
		}
		rpcport[i1]=port
		port=port+1
	} 

	fmt.Println("Minority Failure testing")

    	for {
		select {
		case <-time.After(2 * time.Second):
			
			fmt.Println("------------------------------------------------------------------------------------------------")
			fmt.Println(" {Server_no 	Term 		CurrentStatus}")

			for i:=0;i<totalnumber;i++{				
				client[i].Call("RPCMethods.GetTerm", &args, &results[i])
				if results[i].CurrentState==1{
					fmt.Print(" {",i+1,results[i].Term," Leader","} ")
				}
				if results[i].CurrentState==2{
					fmt.Print(" {",i+1,results[i].Term," Candidate","} ")
				}
				if results[i].CurrentState==3{
					fmt.Print(" {",i+1,results[i].Term," Follower","} ")
				}
			}
			fmt.Println("\n-----------------------------------------------------------------------------------------------")
			
			
			if test%2==0{
				if (totalnumber%2==1&&totalkilled > totalnumber/2)||(totalnumber%2==0&&totalkilled > (totalnumber/2)-1){
						fmt.Println("Majority Failure testing")
						
						if test1==2{
							fmt.Println("PASS")
							for j:=0;j<totalnumber;j++{
								client[j].Close()					
								err:=cmd[j].Process.Kill()
								if err!=nil{
								}
								cmd[j].Process.Wait()
							}
							os.Exit(0)						
						}
						if test1%2==0{
							
						} else {
							fmt.Println("\nServer ",killed[0]+1," Started")
							filename := "cluster"+strconv.Itoa(serverport[killed[0]])+".conf"
							cmd[killed[0]] = exec.Command("./dummy/raftmain", filename, strconv.Itoa(rpcport[killed[0]]))
							cmd[killed[0]].Stdout = os.Stdout
							cmd[killed[0]].Stderr = os.Stderr	
							cmd[killed[0]].Start()
							time.Sleep(2*time.Second)
							client[killed[0]],_ = rpc.DialHTTP("tcp", "127.0.0.1:"+strconv.Itoa(rpcport[killed[0]]))
							totalkilled = totalkilled -1
							killed=killed[1:len(killed)]
						}
						test1=test1+1
						
				}	

				for i:=0;i<totalnumber;i++{								
					if results[i].CurrentState==1{	

						fmt.Println("\n\n Current Leader is ",i+1)						
						client[i].Close()					
						
						err:=cmd[i].Process.Kill()
						if err!=nil{
						}
						cmd[i].Process.Wait()
						fmt.Println("\n Killed server ",i+1,"\n")
						results[i].Term=0
						results[i].CurrentState=0
						totalkilled=totalkilled+1
						killed = append(killed,i)
						leadercount = leadercount + 1
					}									
				}
				
				if leadercount>1{
					t.Error("More than two leaders")
					fmt.Println("FAIL")
					for j:=0;j<totalnumber;j++{
						client[j].Close()					
						err:=cmd[j].Process.Kill()
						if err!=nil{
						}
						cmd[j].Process.Wait()
					}
					os.Exit(0)
				} else {
					leadercount = 0
				}

			}
			test=test+1
		}
    	}
}

