package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"time"
	//"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/libp2p/go-libp2p-core/host"
	//"github.com/libp2p/go-libp2p-core/network"
	//"github.com/libp2p/go-libp2p-core/peer"
//	"github.com/libp2p/go-libp2p-core/peerstore"
	//"github.com/multiformats/go-multiaddr"
)

type Host host.Host



var pError = func(args...interface{}){
	fmt.Print("Error: ")
	fmt.Println(args)
}
var (
	ADD = newOpp("add", "adds a new peer to how peer list")
	PING = newOpp("ping", "chooses a peer and connect to it")
	LIST = newOpp("list", "list all the peers")
	RECORDS = newOpp("records", "list all the records")
	HELP = newOpp("help", "display all the option")
	EXIT = newOpp("exit", "shuts down the program")
)

var MENU = []MenuOpp{
	ADD,
	PING,
	LIST,
	RECORDS,
	HELP,
	EXIT,
}


var p2pMutiAdd string

var BUF = bufio.NewReader(os.Stdin) 

func getLine() string{
	line, _ := BUF.ReadString('\n')
	return strings.Trim(line, "\n\r ")
}

func runReceiver(h Host) {
	// change this later :D
	fmt.Println("To connect run ./mypinger -d", p2pMutiAdd)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	<-ch
	fmt.Println("Received sign shutting down")

}


func runAddPeer(h Host){
	fmt.Print("peerAdd: ")
	peer := getLine()
	if err := addPeer(h, peer); err != nil{
		fmt.Println("error:", err.Error())
	}else{
		fmt.Println("Peer added!")
	}
}


func runHelp(){
	fmt.Println("Opps: ")
	for _, opp := range MENU {
		fmt.Printf("%s - %s\n", opp.option, opp.info)
	}
}

func runList(h Host){
	var msg string
	for i, p := range h.Peerstore().Peers(){
		msg = fmt.Sprintf("Peer[%d] = %s", i, p)	
		if p == h.ID(){
			msg = fmt.Sprintf("%s - You", msg)
		}
		fmt.Println(msg)
		// aside
		//fmt.Println("len = ", len([]byte(p)))
	}
}

func runRecords(){
	file, err := os.ReadFile(FILE_NAME)
	if err != nil{
		if os.IsNotExist(err){
			fmt.Println("No records!")
		}else{
			pError(err.Error())	
			return
		}
	}
	records := strings.Split(string(file), "\n")
	for i, r := range records{
		fmt.Printf("r[%d]: %s\n", i, r)
	}
	
}

func run(h Host) {
	runHelp()
	var (
		over bool = false
		opp string
	)
	for !over{
		fmt.Print(">> ")

		opp = getLine()

		switch opp {
		case ADD.option:
			runAddPeer(h)

		case PING.option:
			go runSender(h)

		case LIST.option:
			runList(h)

		case RECORDS.option:
			runRecords()

		case HELP.option:
			runHelp()
		

		case "":
		case EXIT.option:
			println("exit choosed")
			over = true

		default:
			fmt.Println("Invalid option")
		}
	}

	fmt.Println("Bye Bye!!")
}

// end menu
// addPeers
// display peers
// connect [choose a random peer and connect to it]
// list connections [file and do wonderful stuffs on it]


func main() {

	rand.Seed(int64(time.Now().Nanosecond())) // no need to do this in ipfs
	var( 
		wait, noProtocol bool
		dst string
	)

	flag.BoolVar(&wait, "w", false, "Know If the menu should be displayed or not")
	flag.BoolVar(&noProtocol, "np", false, "To not regist the default protocol")
	flag.StringVar(&dst, "d", "", "dest peer multiaddr")
	flag.Parse()

	if dst != "" && wait{
		fmt.Println("Can't provide -w and -d at once")
		flag.CommandLine.PrintDefaults()
		os.Exit(0)
	}

	// set up 
	h, err := makeHost(0, 0)


	if err != nil{
		panic(err)
	}
	add, err := getHostInfo(h)
	if err != nil {
		panic(err.Error())
	}	
	p2pMutiAdd = add

	// setup end

	fmt.Println("p2pMutiAdd =", p2pMutiAdd)
	if err != nil {
		panic(err)
	}

	if !noProtocol{
		h.SetStreamHandler(PID, handleStream)
	}

	if wait{
		runReceiver(h)
	}else if dst != "" {
		if err:= addPeer(h, dst); err != nil{
			panic(err)
		}
		runSender(h)
	}else {
		run(h)
	}
}
