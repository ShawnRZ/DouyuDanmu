package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"regexp"
	"time"

	"nhooyr.io/websocket"
)

func send(ctx context.Context, c *websocket.Conn) {
	for {
		msgBytes := <-sendCh
		dataLen := len(msgBytes) + 9
		lenBytes := bytes.NewBuffer([]byte{})
		err := binary.Write(lenBytes, binary.LittleEndian, int32(dataLen))
		if err != nil {
			log.Fatalln(err)
		}
		sendBytes := []byte{0xb1, 0x02, 0x00, 0x00}
		endBytes := []byte{0x00}
		data := [][]byte{lenBytes.Bytes(), lenBytes.Bytes(), sendBytes, msgBytes, endBytes}
		dataBytes := bytes.Join(data, nil)
		log.Printf(" %d \n", dataBytes)
		log.Printf(" %d \n", len(dataBytes))
		err = c.Write(ctx, websocket.MessageText, dataBytes)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func recv(ctx context.Context, c *websocket.Conn) {
	for {
		_, msg, err := c.Read(ctx)
		if err != nil {
			log.Println(err)
			continue
		}
		recvCh <- msg[12:]
	}
}

func login(loginMsg string, joinMsg string) {
	sendCh <- []byte(loginMsg)
	<-recvCh
	sendCh <- []byte(joinMsg)
}

func print() {
	for {
		data := <-recvCh
		regName := regexp.MustCompile(`nn@=.*?/`)
		regTXT := regexp.MustCompile(`txt@=.*?/`)
		name := regName.FindAll(data, -1)
		txt := regTXT.FindAll(data, -1)
		if len(name) == 0 || len(txt) == 0 {
			continue
		}
		// fmt.Println(string(data))
		fmt.Println(string(name[0][4:len(name[0])-1]) + ": " + string(txt[0][5:len(txt[0])-1]))
	}
}

func heartickSend() {
	msg := "type@=mrkl/"
	for {
		sendCh <- []byte(msg)
		time.Sleep(time.Second * 15)
	}
}

var recvCh chan []byte
var sendCh chan []byte

func main() {

	roomID := "1126960"
	loginMsg := "type@=loginreq/roomid@=" + roomID + "/dfl@=sn@AA=105@ASss@AA=1/username@=visitor2926208/uid@=1093445248/ver@=20190610/aver@=218101901/ct@=0/"
	joinMsg := "type@=joingroup/rid@=" + roomID + "/gid@=1/"

	recvCh = make(chan []byte, 50)
	sendCh = make(chan []byte, 10)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	c, _, err := websocket.Dial(ctx, "wss://danmuproxy.douyu.com:8506/", nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close(websocket.StatusInternalError, "c.close")
	go recv(ctx, c)
	go send(ctx, c)

	login(loginMsg, joinMsg)
	go heartickSend()

	print()

}
