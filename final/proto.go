package mesurer

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	//github packages
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

// msr stands for mesurer
const (
	PID       = "/ipfs/msr/1.0.0"
	FILE_NAME = "records.txt" // look at this later :>
)

// erros
var (
	errHasNoPeers = errors.New("has no peers")
)

type Host = host.Host

type proto struct {
	host.Host
	target int
}

func ProtoHandlerFunc(s network.Stream) {

	log.Println("Received a connection")

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	buf := make([]byte, 1)
	_, err := rw.Read(buf)

	if err != nil {
		log.Fatal(err.Error())
		return
	}

	log.Println("Pinged")

	_, err = rw.Write(buf)

	if err != nil {
		log.Fatal(err.Error())
		return
	}

	rw.Flush()

	if err = s.Close(); err != nil {
		// you don't wanna use fatal or maybe you do :>
		log.Fatal("Error: ", err.Error())
		return
	}

	log.Println("connection closed")
}

// ask later if you shoudl keep the file always open
func openFile() *os.File {
	file, err := os.OpenFile(FILE_NAME, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		log.Panic(err)
	}
	return file
}

func (h *proto) pingPeer() error {

	// look at this later it might be useful h.Peerstore().RecordLatency()
	// check if peer suppports protocol
	// I'm gonna do something related to this
	/*
		if !SupportsProtocol(peerID, h){
			return errDoNotSupportProtocol
		}
	*/

	peerID, err := h.getTarget()
	if err != nil {
		return err
	}
	s, err := h.NewStream(context.Background(), peerID, PID)

	if err != nil {
		return err
	}

	log.Println("Connection established")

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	var L uint8 = uint8(rand.Uint32())

	buf := []byte{L}

	t := time.Now()
	_, err = rw.Write(buf)

	if err != nil {
		return err
	}

	rw.Flush()

	_, err = rw.Read(buf)

	if err != nil {
		return err
	}

	// ask about this later
	RTT := time.Since(t)

	file := openFile()

	msg := fmt.Sprintf("From: %s to %s - nRTT: %d", h.ID(), peerID, RTT)

	if aux, _ := file.Stat(); aux.Size() != 0 {
		msg = "\n" + msg
	}

	// later

	// ignore all these erros :>
	file.WriteString(msg)

	file.Sync()

	file.Close()

	// look at this later :)
	// think about this later :>
	if buf[0] != L {
		// think about this later :>
		return fmt.Errorf("not received %d received %d", L, buf[0]) // not of your business
	}

	log.Println("Ponged")

	// should I ignore
	if err = s.Close(); err != nil {
		return err
	}

	log.Println("connection closed")
	return nil
}

func (p *proto) getTarget() (peer.ID, error) {

	// you're gonna have to do more work over here

	peers := p.Peerstore().Peers()
	pSize := len(peers)

	if pSize <= 1 {
		return "", errHasNoPeers
	}

	incT := func() {
		p.target = (p.target + 1) % pSize
	}
	// new algo by now
	if peers[p.target] == p.ID() {
		incT()
	}
	/*
		target := rand.Intn(pSize)

		for peers[target] == p.ID() {
			target = rand.Intn(pSize)
		}
	*/

	target := p.target
	incT()
	return peers[target], nil
}

func createFile() {

	file, err := os.Open(FILE_NAME)
	if err != nil && os.IsNotExist(err) {

		file, err = os.Create(FILE_NAME)
		if err != nil {
			log.Fatal("Problems create file :>")
		}
	}
	file.Close()
}


func run(p *proto) {
	createFile()
	var err error
	var msg string
	for {

		msg = fmt.Sprintf("target = %d ::", p.target)
		if err = p.pingPeer(); err != nil {
			// decide what to do
			log.Println(msg,err)
			time.Sleep(30 * time.Second)
			continue
		}
		log.Println("+1 regs")
		// by now
		time.Sleep(15 * time.Second)
	}
}

// maybe I will have to change it later
func LaunchProto(host host.Host) {
	host.SetStreamHandler(PID, ProtoHandlerFunc)
	proto := proto{host, 0}
	fmt.Println("Sending back thread")
	go run(&proto)
}
