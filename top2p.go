package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"errors"
	// github packages
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

/*
type Record struct{
	Sender [34]byte
	Receiver [34]byte
	Rtt int64 // in us
}
*/

const ( // think abou the name :?
	PID       = "/tmp/1.0.0" // /ipfs/tmp/1.0.0
	FILE_NAME = "Records.txt"
)

var (
	errHasNoPeers = errors.New("has no peer")
	errDoNotSupportProtocol = errors.New("peer doesn't support protocol")
)	

var noSupportedPeers = make(map[peer.ID]bool)



/*
	Idea:
	type Record struct{
		Sender [34]byte
		Receiver [34]byte
		RTT int64
	}
*/

func handleStream(s network.Stream) {

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
		log.Fatal("Error: ", err.Error())
		return
	}

	log.Println("connection closed")

}

func openFile() *os.File {

	file, err := os.OpenFile(FILE_NAME, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		// ask later
		if !os.IsNotExist(err) {
			log.Panic(err)
		}
		file, err = os.OpenFile(FILE_NAME, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			log.Panic(err)
		}
	}

	return file
}

func SupportsProtocol(peerID peer.ID, h Host) bool{

	protos, _ :=  h.Peerstore().GetProtocols(peerID)
	for _, proto := range protos{
		if proto == PID{
			return true
		} 
	}

	return false
}

// change later
func pingPeer(peerID peer.ID, h Host) error {

	// look at this later it might be useful h.Peerstore().RecordLatency()
	// check if peer suppports protocol
	// I'm gonna do something related to this
	/*
	if !SupportsProtocol(peerID, h){
		return errDoNotSupportProtocol
	}
	*/

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
	if buf[0] != L {
		// think about this later :>
		return fmt.Errorf("not received %d received %d", L, buf[0]) // not of your business
	}

	log.Println("Ponged")

	// should I ignore
	if err = s.Close(); err != nil {
		return err
	}

	return nil
}

func chooseAndPing(h Host) error {
	peers := h.Peerstore().Peers()
	pSize := len(peers)

	if pSize <= (1 + len(noSupportedPeers)){
		return errHasNoPeers 
	}

	target := rand.Intn(pSize)

	for peers[target] == h.ID() {
		target = rand.Intn(pSize)
	}

	/*
		if peers[target] == h.ID(){
			if(target < pSize - 1){
				target++
			}else{
				target--
			}
		}
	*/


	if err := pingPeer(peers[target], h); err != nil {
		noSupportedPeers[peers[target]] = true
		// if doesn't support the protocol do nothing by now
		// wait and continue
		return err
	}

	log.Println("connection closed")
	return nil
}

//CHANGE LATER
// requires time > 0
func runSender(h Host) {

	if err := chooseAndPing(h); err != nil{
		log.Fatal(err)
	}
}
