package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/omec-project/smf/pfcp/message"
)

var seq uint32

func getSeqNumber() uint32 {
	return atomic.AddUint32(&seq, 1)
}

func rec_msg(w http.ResponseWriter, req *http.Request) {

	for headerName, headerValue := range req.Header {
		fmt.Printf("\t%s = %s\n", headerName, strings.Join(headerValue, ", "))
	}

	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Printf("server: could not read request body: %s\n", err)
	}

	// reconstruct pfcp msg
	fmt.Printf("server: request body: %s\n", reqBody)

	msgType := req.Header.Get("MsgType")
	fmt.Printf("server: msgType: %s\n", msgType)
	fmt.Printf("server: msgType type: %T\n", msgType)

	adapterId := getSeqNumber()

	var udpPodMsg message.UdpPodPfcpMsg
	json.Unmarshal(reqBody, &udpPodMsg)
	// logs for debug
	fmt.Println("--------------------------------------")
	fmt.Printf("server: udpPodMsg.Msg.Header: %v\n\n", udpPodMsg.Msg.Header)
	fmt.Printf("server: udpPodMsg.Msg.Header type: %T\n\n", udpPodMsg.Msg.Header)
	fmt.Printf("server: udpPodMsg.Msg.Body: %v\n\n", udpPodMsg.Msg.Body)
	fmt.Printf("server: udpPodMsg.Msg.Body type: %T\n\n", udpPodMsg.Msg.Body)
	fmt.Printf("server: adapterId: %v\n", adapterId)
	fmt.Println("--------------------------------------")

	// return marshalled pfcp message
	udpPodMsgJson, _ := json.Marshal(udpPodMsg)
	fmt.Println("udpPodMsgJson := ", udpPodMsgJson)
	fmt.Println("udpPodMsgJson := ", string(udpPodMsgJson))

	jsonRsp := []byte(udpPodMsgJson)
	w.Write(jsonRsp)

}

func main() {
	http.HandleFunc("/", rec_msg)
	http.ListenAndServe(":8090", nil)

}
