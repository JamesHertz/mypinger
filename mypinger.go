package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	mesurer "mypinger/final"
	"time"

	//"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	//"github.com/libp2p/go-libp2p-core/network"
	//	"github.com/libp2p/go-libp2p-core/peer"
	//	"github.com/libp2p/go-libp2p-core/peerstore"
	//"github.com/multiformats/go-multiaddr"
)

type Host host.Host

var pError = func(args ...interface{}) {
	fmt.Print("Error: ")
	fmt.Println(args)
}
var (
	ADD     = newOpp("add", "adds a new peer to how peer list")
	PING    = newOpp("ping", "chooses a peer and connect to it")
	LIST    = newOpp("list", "list all the peers")
	UPLOAD  = newOpp("upload", "uploads the online peers")
	RECORDS = newOpp("records", "list all the records")
	PROTO   = newOpp("proto", "run the pre version of the ipfs protocol")
	HELP    = newOpp("help", "display all the option")
	EXIT    = newOpp("exit", "shuts down the program")
)

var MENU = []MenuOpp{
	ADD,
	PING,
	LIST,
	RECORDS,
	PROTO,
	HELP,
	EXIT,
}

// file where p2pADd of the nodes alive are written
const P_FILE = ".peers.info"

// the node's multiAdd
var p2pMutiAdd string

// used for input
var BUF = bufio.NewReader(os.Stdin)

func getLine() string {
	line, _ := BUF.ReadString('\n')
	return strings.Trim(line, "\n\r ")
}

func runAddPeer(h Host) {
	fmt.Print("peerAdd: ")
	peer := getLine()
	if err := addPeer(h, peer); err != nil {
		pError(err)
	} else {
		fmt.Println("Peer added!")
	}
}

func runHelp() {
	fmt.Println("Opps: ")
	for _, opp := range MENU {
		fmt.Printf("%s - %s\n", opp.option, opp.info)
	}
}

func runList(h Host) {
	var msg string
	for i, p := range h.Peerstore().Peers() {
		msg = fmt.Sprintf("Peer[%d] = %s", i, p)
		if p == h.ID() {
			msg = fmt.Sprintf("%s - You", msg)
		}
		fmt.Println(msg)
	}
}

func runRecords() {
	file, err := os.ReadFile(FILE_NAME)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No records!")
		} else {
			pError(err)
			return
		}
	}
	records := strings.Split(string(file), "\n")
	for i, r := range records {
		fmt.Printf("r[%d]: %s\n", i, r)
	}

}

func tmp(h Host) {
	peers := h.Peerstore().Peers()
	for _, p := range peers {
		if p != h.ID() {
			s, err := h.Peerstore().SupportsProtocols(p, PID)
			fmt.Printf("s=%v, err = %v", s, err)
		}
	}
}

type upEntry struct {
	multiadd string
	in       bool
}

func runUpload(h Host) {
	peers, err := readPFile()
	if err != nil {
		pError(err)
		return
	}

	onPeers := make(map[string]*upEntry)
	var offPeers []peer.ID

	for _, p := range peers {
		if p != p2pMutiAdd {
			k := strings.Split(p, "/p2p/")[1]
			onPeers[k] = &upEntry{multiadd: p, in: false}
		}
	}

	for _, p := range h.Peerstore().Peers() {
		if p == h.ID() {
			continue
		}

		if ptr, in := onPeers[string(p)]; !in {
			offPeers = append(offPeers, p)
		} else {
			ptr.in = true
		}
	}

	// tries to remove a peer but we can't do it :(
	for _, p := range offPeers {

		h.Network().ClosePeer(p)
		//h.Peerstore().RemovePeer(p)
		fmt.Printf("%s removed!\n", p)
	}

	for v, p := range onPeers {
		if !p.in {
			addPeer(h, p.multiadd)
			fmt.Printf("%s added!\n", v)
		}
	}

}

func run(h Host) {
	runHelp()
	var (
		over bool = false
		opp  string
	)
	for !over {
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
			//
		case PROTO.option:
			mesurer.LaunchProto(h)
		case UPLOAD.option:
			runUpload(h)

		case "":

		case "tmp":
			tmp(h)
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

// init methods

func initP(h Host) {
	peers, err := readPFile()
	if err == nil {
		for _, p := range peers {
			addPeer(h, p)
		}
	}
	peers = append(peers, p2pMutiAdd)
	if err = writePFile(peers); err != nil {
		pError(err)
		os.Exit(0)
	}
}

// takes out the peerAdd from the P_FILE
func exitP() {
	peers, err := readPFile()
	if err != nil {
		pError(err)
		return
	}
	var newPeers []string
	for _, p := range peers {
		if p != p2pMutiAdd {
			newPeers = append(newPeers, p)
		}
	}
	if err = writePFile(newPeers); err != nil {
		pError(err)
	}
}

// reads the P_FILE and slice of the peerAdd
func readPFile() ([]string, error) {
	tmp, err := os.ReadFile(P_FILE)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(tmp), "\n"), nil
}

// truncs the P_FILE and write a new set of add
func writePFile(peers []string) error {
	data := strings.Join(peers, "\n")
	data = strings.Trim(data, "\n\r ")
	if err := os.WriteFile(P_FILE, []byte(data), 0660); err != nil {
		return err
	}
	return nil
}

// main
func main() {
	rand.Seed(int64(time.Now().Nanosecond())) // no need to do this in ipfs
	var (
		wait, noProtocol, noRegist bool
		dst                        string
	)

	flag.BoolVar(&wait, "w", false, "Know If the menu should be displayed or not")
	flag.BoolVar(&noProtocol, "np", false, "To not regist the default protocol")
	flag.StringVar(&dst, "d", "", "dest peer multiaddr")
	flag.BoolVar(&noRegist, "nr", false, "do not regist the peer and do not upload all the online peers")

	flag.Parse()

	if dst != "" && wait {
		fmt.Println("Can't provide -w and -d at once")
		flag.CommandLine.PrintDefaults()
		os.Exit(0)
	}

	// set up
	h, err := makeHost(0, 0)

	if err != nil {
		panic(err)
	}

	p2pMutiAdd, err := getHostInfo(h)

	if err != nil {
		panic(err)
	}


	// if they want to leave the default config which is to regist the node
	// and upload online nodes let's do it
	if !noRegist {
		initP(h)
		go func() {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
			<-ch
			fmt.Println("Received sign shutting down")

			exitP()
			os.Exit(0)
		}()
	}

	fmt.Println("p2pMutiAdd =", p2pMutiAdd)
	if err != nil {
		panic(err)
	}

	if !noProtocol {
		h.SetStreamHandler(PID, mesurer.ProtoHandlerFunc)
	}

	if wait {
		fmt.Println("To connect run ./mypinger -d", p2pMutiAdd)
		select {}
	} else if dst != "" {
		if err := addPeer(h, dst); err != nil {
			panic(err)
		}
		runSender(h)
	} else {
		run(h)
	}

	if !noRegist {
		exitP()
	}
}
