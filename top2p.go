package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

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

const (					// think abou the name :?
	PID = "/tmp/1.0.0" // /ipfs/tmp/1.0.0
	FILE_NAME = "Records.txt"
)
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

func openFile() *os.File{

	file, err := os.OpenFile(FILE_NAME, os.O_WRONLY | os.O_APPEND, 0600)
	if err != nil{
		// ask later
		if !os.IsNotExist(err){
			log.Panic(err)
		}
		file, err = os.OpenFile(FILE_NAME, os.O_CREATE | os.O_WRONLY, 0600)
		if err != nil{
			log.Panic(err)
		}
	}

	return file  
}
// change later
func pingPeer(peerID peer.ID, h Host) error {

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

	RTT := time.Since(t)

	file := openFile() 


	msg := fmt.Sprintf("From: %s to %s - nRTT: %d", h.ID(), peerID, RTT)

	if aux, _ := file.Stat(); aux.Size() != 0{
		msg = "\n" + msg
	}

	// later
	
	// ignore all these erros :>
	if _, err = file.WriteString(msg); err != nil{
		return err
	}

	if err = file.Sync(); err != nil{
		return err
	}

	if err = file.Close(); err != nil{
		return err
	}

	// look at this later :)

	// think about this thing ...
	if buf[0] != L {
		return fmt.Errorf("not received %d received %d", L, buf[0]) // not of your business
	}

	log.Println("Ponged")

	if err = s.Close(); err != nil {
		return err
	}

	return nil
}


//CHANGE LATER
func runSender(h Host) {

	// change this later
	//addPeer(h, dest)
	peers := h.Peerstore().Peers()
	pSize := len(peers)

	if pSize <= 1{
		log.Println("no peers added")
		return
	}

	target := rand.Intn(pSize)
	for peers[target] == h.ID(){
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

	// change this later
	//var peerID = peers[target]
	
	if err := pingPeer(peers[target], h); err != nil {
		log.Fatal(err)
		return
	}

	log.Println("connection closed")
}


