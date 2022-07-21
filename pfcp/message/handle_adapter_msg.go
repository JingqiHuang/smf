// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
// Copyright 2019 free5GC.org
//
// SPDX-License-Identifier: Apache-2.0

package message

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/omec-project/pfcp"
	"github.com/omec-project/smf/logger"
)

func ProcessPfcpAssociationRsp(rsp *http.Response) pfcp.PFCPAssociationSetupResponse {
	bodyBytes, err := io.ReadAll(rsp.Body)
	if err != nil {
		logger.PfcpLog.Fatalln(err)
	}
	rspBodyString := string(bodyBytes)
	logger.PfcpLog.Infof("rspBodyString ", rspBodyString)

	uMsg := UdpPodPfcpMsg{}
	pMsg := pfcp.PFCPAssociationSetupResponse{}

	json.Unmarshal(bodyBytes, &uMsg)
	msg := uMsg.Msg.Body

	msgString, _ := json.Marshal(msg)
	json.Unmarshal(msgString, &pMsg)
	fmt.Printf("ProcessPfcpAssociationRsp pMsg after unmarshal val %v\n", pMsg)
	fmt.Printf("ProcessPfcpAssociationRsp pMsg after unmarshal type %T\n", pMsg)
	// return pMsg here
	return pMsg
}

func ProcessHeartbeatRsp(rsp *http.Response) pfcp.HeartbeatResponse {
	bodyBytes, err := io.ReadAll(rsp.Body)
	if err != nil {
		logger.PfcpLog.Fatalln(err)
	}
	rspBodyString := string(bodyBytes)
	logger.PfcpLog.Infof("rspBodyString ", rspBodyString)

	uMsg := UdpPodPfcpMsg{}
	pMsg := pfcp.HeartbeatResponse{}

	json.Unmarshal(bodyBytes, &uMsg)
	msg := uMsg.Msg.Body

	msgString, _ := json.Marshal(msg)
	json.Unmarshal(msgString, &pMsg)
	fmt.Printf("ProcessHeartbeatRsp pMsg after unmarshal val %v\n", pMsg)
	fmt.Printf("ProcessHeartbeatRsp pMsg after unmarshal type %T\n", pMsg)
	// return pMsg here
	return pMsg
}

func ProcessPfcpAssociationReleaseRsp(rsp *http.Response) pfcp.PFCPAssociationReleaseResponse {

	pMsg := pfcp.PFCPAssociationReleaseResponse{}
	// bodyBytes, err := io.ReadAll(rsp.Body)
	// if err != nil {
	// 	logger.PfcpLog.Fatalln(err)
	// }
	// rspBodyString := string(bodyBytes)
	// logger.PfcpLog.Infof("rspBodyString ", rspBodyString)

	// uMsg := UdpPodPfcpMsg{}

	// json.Unmarshal(bodyBytes, &uMsg)
	// msg := uMsg.Msg.Body

	// msgString, _ := json.Marshal(msg)
	// json.Unmarshal(msgString, &pMsg)
	fmt.Printf("ProcessPfcpAssociationReleaseRsp pMsg after unmarshal val %v\n", pMsg)
	fmt.Printf("ProcessPfcpAssociationReleaseRsp pMsg after unmarshal type %T\n", pMsg)
	// return pMsg here
	return pMsg
}
func ProcessPfcpSessionEstablishmentRsp(rsp *http.Response) pfcp.PFCPSessionEstablishmentResponse {
	pMsg := pfcp.PFCPSessionEstablishmentResponse{}
	// bodyBytes, err := io.ReadAll(rsp.Body)
	// if err != nil {
	// 	logger.PfcpLog.Fatalln(err)
	// }
	// rspBodyString := string(bodyBytes)
	// logger.PfcpLog.Infof("rspBodyString ", rspBodyString)

	// uMsg := UdpPodPfcpMsg{}

	// json.Unmarshal(bodyBytes, &uMsg)
	// msg := uMsg.Msg.Body

	// msgString, _ := json.Marshal(msg)
	// json.Unmarshal(msgString, &pMsg)
	fmt.Printf("ProcessPfcpSessionEstablishmentRsp pMsg after unmarshal val %v\n", pMsg)
	fmt.Printf("ProcessPfcpSessionEstablishmentRsp pMsg after unmarshal type %T\n", pMsg)
	// return pMsg here
	return pMsg
}
func ProcessPfcpSessionModificationRsp(rsp *http.Response) pfcp.PFCPSessionModificationResponse {
	pMsg := pfcp.PFCPSessionModificationResponse{}
	// bodyBytes, err := io.ReadAll(rsp.Body)
	// if err != nil {
	// 	logger.PfcpLog.Fatalln(err)
	// }
	// rspBodyString := string(bodyBytes)
	// logger.PfcpLog.Infof("rspBodyString ", rspBodyString)

	// uMsg := UdpPodPfcpMsg{}

	// json.Unmarshal(bodyBytes, &uMsg)
	// msg := uMsg.Msg.Body

	// msgString, _ := json.Marshal(msg)
	// json.Unmarshal(msgString, &pMsg)
	fmt.Printf("ProcessPfcpSessionModificationRsp pMsg after unmarshal val %v\n", pMsg)
	fmt.Printf("ProcessPfcpSessionModificationRsp pMsg after unmarshal type %T\n", pMsg)
	// return pMsg here
	return pMsg
}

func ProcessPfcpSessionDeletionRsp(rsp *http.Response) pfcp.PFCPSessionDeletionResponse {
	pMsg := pfcp.PFCPSessionDeletionResponse{}
	// bodyBytes, err := io.ReadAll(rsp.Body)
	// if err != nil {
	// 	logger.PfcpLog.Fatalln(err)
	// }
	// rspBodyString := string(bodyBytes)
	// logger.PfcpLog.Infof("rspBodyString ", rspBodyString)

	// uMsg := UdpPodPfcpMsg{}

	// json.Unmarshal(bodyBytes, &uMsg)
	// msg := uMsg.Msg.Body

	// msgString, _ := json.Marshal(msg)
	// json.Unmarshal(msgString, &pMsg)
	fmt.Printf("ProcessPfcpSessionDeletionRsp pMsg after unmarshal val %v\n", pMsg)
	fmt.Printf("ProcessPfcpSessionDeletionRsp pMsg after unmarshal type %T\n", pMsg)
	// return pMsg here
	return pMsg
}
