package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"
	. "ve489/util"
)

var num = flag.Int("n", 20, "Input how many times")
var ip = flag.String("i", "10.3.63.2", "Server IP")
var port = flag.Int("p", 8002, "Server Port")

func main() {
	flag.Parse()

	// read the file
	data, err := ioutil.ReadFile("../shakespeare.txt")
	if err != nil {
		fmt.Println("ReadFile error:", err)
		return
	}

	// connect with the server
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", *ip, *port))
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Dial complete")

	// send the file content in a loop
	txSeqNum, count := false, 0
	for i := 0; i < len(data); i++ {
		// _, err = conn.Write([]byte(fmt.Sprintf("Message%d", Bool2Int(txSeqNum))))
		msg := make([]byte, 2)
		msg[0], msg[1] = data[i], Bool2Byte(txSeqNum)

		// send one message = one byte + ack
		_, err = conn.Write([]byte(string(msg)))
		if err != nil {
			fmt.Println("conn.Write err:", err)
		}
		fmt.Printf("Sent Message %d (Seq: %d)\n", count, Bool2Int(txSeqNum))

		err = conn.SetReadDeadline(time.Now().Add(1500 * time.Millisecond))
		if err != nil {
			fmt.Println("conn.SetReadDeadline err:", err)
		}

		// read the ack
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)

		// if timeout, resend the ack
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				fmt.Printf("Waiting for ACK timeout, will resend Message %d (Seq: %d)\n", count, Bool2Int(txSeqNum))
				i --
			} else {
				return
			}
		} else {
			fmt.Println("Received from Server:", string(buf[:n]))
			count++
			txSeqNum = !txSeqNum
		}

		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("Client exits")
}
