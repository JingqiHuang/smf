// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
// Copyright 2019 free5GC.org
//
// SPDX-License-Identifier: Apache-2.0

package message

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"sync/atomic"

	"github.com/omec-project/nas/nasMessage"
	"github.com/omec-project/openapi/models"
	"github.com/omec-project/pfcp"
	"github.com/omec-project/pfcp/pfcpType"
	"github.com/omec-project/pfcp/pfcpUdp"
	smf_context "github.com/omec-project/smf/context"
	"github.com/omec-project/smf/logger"
	"github.com/omec-project/smf/metrics"
	"github.com/omec-project/smf/msgtypes/pfcpmsgtypes"
	"github.com/omec-project/smf/pfcp/udp"
)

var seq uint32

var ip = "172.16.45.23"

var udp_adapter = true

func getSeqNumber() uint32 {
	return atomic.AddUint32(&seq, 1)
}

func init() {
	PfcpTxns = make(map[uint32]*pfcpType.NodeID)
}

var (
	PfcpTxns    map[uint32]*pfcpType.NodeID
	PfcpTxnLock sync.Mutex
)

func FetchPfcpTxn(seqNo uint32) (upNodeID *pfcpType.NodeID) {
	PfcpTxnLock.Lock()
	defer PfcpTxnLock.Unlock()
	if upNodeID = PfcpTxns[seqNo]; upNodeID != nil {
		delete(PfcpTxns, seqNo)
	}
	return upNodeID
}

func InsertPfcpTxn(seqNo uint32, upNodeID *pfcpType.NodeID) {
	PfcpTxnLock.Lock()
	defer PfcpTxnLock.Unlock()
	PfcpTxns[seqNo] = upNodeID
}

func SendHeartbeatRequest(upNodeID pfcpType.NodeID) error {
	pfcpMsg, err := BuildPfcpHeartbeatRequest()
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP Heartbeat Request failed: %v", err)
		return err
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:        pfcp.PfcpVersion,
			MP:             0,
			S:              pfcp.SEID_NOT_PRESENT,
			MessageType:    pfcp.PFCP_HEARTBEAT_REQUEST,
			SequenceNumber: getSeqNumber(),
		},
		Body: pfcpMsg,
	}

	addr := &net.UDPAddr{
		IP:   upNodeID.ResolveNodeIdToIp(),
		Port: pfcpUdp.PFCP_PORT,
	}

	InsertPfcpTxn(message.Header.SequenceNumber, &upNodeID)

	if udp_adapter {
		rsp, err := SendPfcpMsgToAdapter(upNodeID, message, addr, nil)
		if err != nil {
			logger.PfcpLog.Errorf("SendPfcpMsgToAdapter failed: %v", err)
			return err
		}
		pMsg := ProcessHeartbeatRsp(rsp)
		logger.PfcpLog.Infof("after ProcessHeartbeatRsp pMsg val %v\n", pMsg)
		logger.PfcpLog.Infof("after ProcessHeartbeatRsp pMsg type %T\n", pMsg)
	}

	if err := udp.SendPfcp(message, addr, nil); err != nil {
		FetchPfcpTxn(message.Header.SequenceNumber)
		return err
	}
	logger.PfcpLog.Debugf("sent pfcp heartbeat request seq[%d] to NodeID[%s]", message.Header.SequenceNumber,
		upNodeID.ResolveNodeIdToIp().String())
	return nil
}

func SendPfcpAssociationSetupRequest(upNodeID pfcpType.NodeID) {
	if net.IP.Equal(upNodeID.ResolveNodeIdToIp(), net.IPv4zero) {
		logger.PfcpLog.Errorf("PFCP Association Setup Request failed, invalid NodeId: %v", string(upNodeID.NodeIdValue))
		return
	}

	pfcpMsg, err := BuildPfcpAssociationSetupRequest()
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP Association Setup Request failed: %v", err)
		return
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:        pfcp.PfcpVersion,
			MP:             0,
			S:              pfcp.SEID_NOT_PRESENT,
			MessageType:    pfcp.PFCP_ASSOCIATION_SETUP_REQUEST,
			SequenceNumber: getSeqNumber(),
		},
		Body: pfcpMsg,
	}

	addr := &net.UDPAddr{
		IP:   upNodeID.ResolveNodeIdToIp(),
		Port: pfcpUdp.PFCP_PORT,
	}

	if udp_adapter {
		rsp, err := SendPfcpMsgToAdapter(upNodeID, message, addr, nil)
		if err != nil {
			logger.PfcpLog.Errorf("SendPfcpMsgToAdapter failed: %v", err)
			return
		}
		if rsp.StatusCode == http.StatusOK {
			// process response
			pMsg := ProcessPfcpAssociationRsp(rsp)
			logger.PfcpLog.Infof("after ProcessPfcpAssociationRsp pMsg val %v\n", pMsg)
			logger.PfcpLog.Infof("after ProcessPfcpAssociationRsp pMsg type %T\n", pMsg)
		}
	}
	udp.SendPfcp(message, addr, nil)

	logger.PfcpLog.Infof("Sent PFCP Association Request to NodeID[%s]", upNodeID.ResolveNodeIdToIp().String())
}

func SendPfcpAssociationSetupResponse(upNodeID pfcpType.NodeID, cause pfcpType.Cause) {
	pfcpMsg, err := BuildPfcpAssociationSetupResponse(cause)
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP Association Setup Response failed: %v", err)
		return
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:        pfcp.PfcpVersion,
			MP:             0,
			S:              pfcp.SEID_NOT_PRESENT,
			MessageType:    pfcp.PFCP_ASSOCIATION_SETUP_RESPONSE,
			SequenceNumber: 1,
		},
		Body: pfcpMsg,
	}

	addr := &net.UDPAddr{
		IP:   upNodeID.ResolveNodeIdToIp(),
		Port: pfcpUdp.PFCP_PORT,
	}

	if udp_adapter {
		_, err := SendPfcpMsgToAdapter(upNodeID, message, addr, nil)
		if err != nil {
			logger.PfcpLog.Errorf("in SendPfcpAssociationSetupResponse SendPfcpMsgToAdapter failed: %v", err)
			return
		}

		// do nothing for sending the rsp

		// if rsp.StatusCode == http.StatusOK {
		// 	// process response
		// 	pMsg := ProcessPfcpAssociationRsp(rsp)
		// 	logger.PfcpLog.Infof("after ProcessPfcpAssociationRsp pMsg val %v\n", pMsg)
		// 	logger.PfcpLog.Infof("after ProcessPfcpAssociationRsp pMsg type %T\n", pMsg)
		// }
	}

	udp.SendPfcp(message, addr, nil)
	logger.PfcpLog.Infof("Sent PFCP Association Response to NodeID[%s]", upNodeID.ResolveNodeIdToIp().String())
}

func SendPfcpAssociationReleaseRequest(upNodeID pfcpType.NodeID) {
	pfcpMsg, err := BuildPfcpAssociationReleaseRequest()
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP Association Release Request failed: %v", err)
		return
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:        pfcp.PfcpVersion,
			MP:             0,
			S:              pfcp.SEID_NOT_PRESENT,
			MessageType:    pfcp.PFCP_ASSOCIATION_RELEASE_REQUEST,
			SequenceNumber: 1,
		},
		Body: pfcpMsg,
	}

	addr := &net.UDPAddr{
		IP:   upNodeID.ResolveNodeIdToIp(),
		Port: pfcpUdp.PFCP_PORT,
	}

	if udp_adapter {
		rsp, err := SendPfcpMsgToAdapter(upNodeID, message, addr, nil)
		if err != nil {
			logger.PfcpLog.Errorf("in SendPfcpAssociationReleaseRequest SendPfcpMsgToAdapter failed: %v", err)
			return
		}

		if rsp.StatusCode == http.StatusOK {
			// process response
			pMsg := ProcessPfcpAssociationReleaseRsp(rsp)
			logger.PfcpLog.Infof("after ProcessPfcpAssociationReleaseRsp pMsg val %v\n", pMsg)
			logger.PfcpLog.Infof("after ProcessPfcpAssociationReleaseRsp pMsg type %T\n", pMsg)
		}
	}

	udp.SendPfcp(message, addr, nil)
	logger.PfcpLog.Infof("Sent PFCP Association Release Request to NodeID[%s]", upNodeID.ResolveNodeIdToIp().String())
}

func SendPfcpAssociationReleaseResponse(upNodeID pfcpType.NodeID, cause pfcpType.Cause) {
	pfcpMsg, err := BuildPfcpAssociationReleaseResponse(cause)
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP Association Release Response failed: %v", err)
		return
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:        pfcp.PfcpVersion,
			MP:             0,
			S:              pfcp.SEID_NOT_PRESENT,
			MessageType:    pfcp.PFCP_ASSOCIATION_RELEASE_RESPONSE,
			SequenceNumber: 1,
		},
		Body: pfcpMsg,
	}

	addr := &net.UDPAddr{
		IP:   upNodeID.ResolveNodeIdToIp(),
		Port: pfcpUdp.PFCP_PORT,
	}

	if udp_adapter {
		_, err := SendPfcpMsgToAdapter(upNodeID, message, addr, nil)
		if err != nil {
			logger.PfcpLog.Errorf("in SendPfcpAssociationReleaseResponse SendPfcpMsgToAdapter failed: %v", err)
			return
		}
		// do nothing for sending the rsp
		// if rsp.StatusCode == http.StatusOK {
		// 	// process response
		// 	pMsg := ProcessPfcpAssociationReleaseRsp(rsp)
		// 	logger.PfcpLog.Infof("after ProcessPfcpAssociationReleaseRsp pMsg val %v\n", pMsg)
		// 	logger.PfcpLog.Infof("after ProcessPfcpAssociationReleaseRsp pMsg type %T\n", pMsg)
		// }
	}

	udp.SendPfcp(message, addr, nil)
	logger.PfcpLog.Infof("Sent PFCP Association Release Response to NodeID[%s]", upNodeID.ResolveNodeIdToIp().String())
}

func SendPfcpSessionEstablishmentRequest(
	upNodeID pfcpType.NodeID,
	ctx *smf_context.SMContext,
	pdrList []*smf_context.PDR, farList []*smf_context.FAR, barList []*smf_context.BAR, qerList []*smf_context.QER) {
	pfcpMsg, err := BuildPfcpSessionEstablishmentRequest(upNodeID, ctx, pdrList, farList, barList, qerList)
	if err != nil {
		ctx.SubPfcpLog.Errorf("Build PFCP Session Establishment Request failed: %v", err)
		return
	}

	ip := upNodeID.ResolveNodeIdToIp()

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:         pfcp.PfcpVersion,
			MP:              1,
			S:               pfcp.SEID_PRESENT,
			MessageType:     pfcp.PFCP_SESSION_ESTABLISHMENT_REQUEST,
			SEID:            0,
			SequenceNumber:  getSeqNumber(),
			MessagePriority: 0,
		},
		Body: pfcpMsg,
	}

	upaddr := &net.UDPAddr{
		IP:   ip,
		Port: pfcpUdp.PFCP_PORT,
	}
	ctx.SubPduSessLog.Traceln("[SMF] Send SendPfcpSessionEstablishmentRequest")
	ctx.SubPduSessLog.Traceln("Send to addr ", upaddr.String())

	eventData := pfcpUdp.PfcpEventData{LSEID: ctx.PFCPContext[ip.String()].LocalSEID, ErrHandler: HandlePfcpSendError}
	if udp_adapter {
		rsp, err := SendPfcpMsgToAdapter(upNodeID, message, upaddr, eventData)
		if err != nil {
			logger.PfcpLog.Errorf("in SendPfcpSessionEstablishmentRequest SendPfcpMsgToAdapter failed: %v", err)
			// TODO: call eventData.ErrHandler here
			return
		}
		if rsp.StatusCode == http.StatusOK {
			// process response
			pMsg := ProcessPfcpSessionEstablishmentRsp(rsp)
			logger.PfcpLog.Infof("after ProcessPfcpAssociationReleaseRsp pMsg val %v\n", pMsg)
			logger.PfcpLog.Infof("after ProcessPfcpAssociationReleaseRsp pMsg type %T\n", pMsg)
			// TODO: handle pMsg using pfcp.Handler funcs

		}
	}
	eventData = pfcpUdp.PfcpEventData{LSEID: ctx.PFCPContext[ip.String()].LocalSEID, ErrHandler: HandlePfcpSendError}

	udp.SendPfcp(message, upaddr, eventData)
	ctx.SubPfcpLog.Infof("Sent PFCP Session Establish Request to NodeID[%s]", ip.String())
}

// Deprecated: PFCP Session Establishment Procedure should be initiated by the CP function
func SendPfcpSessionEstablishmentResponse(addr *net.UDPAddr) {
	pfcpMsg, err := BuildPfcpSessionEstablishmentResponse()
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP Session Establishment Response failed: %v", err)
		return
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:         pfcp.PfcpVersion,
			MP:              1,
			S:               pfcp.SEID_PRESENT,
			MessageType:     pfcp.PFCP_SESSION_ESTABLISHMENT_RESPONSE,
			SEID:            123456789123456789,
			SequenceNumber:  1,
			MessagePriority: 12,
		},
		Body: pfcpMsg,
	}
	// TODO: need ip -> nodeID
	/*
		if udp_adapter {
			// TODO: need ip -> NodeID transfer
			_, err := SendPfcpMsgToAdapter(upNodeID, message, addr, nil)
			if err != nil {
				logger.PfcpLog.Errorf("in SendPfcpAssociationReleaseRequest SendPfcpMsgToAdapter failed: %v", err)
				return
			}
			// do nothing afetr sending pfcp rsp
			// if rsp.StatusCode == http.StatusOK {
			// 	// process response
			// 	pMsg := ProcessPfcpSessionEstablishmentRsp(rsp)
			// 	logger.PfcpLog.Infof("after ProcessPfcpAssociationReleaseRsp pMsg val %v\n", pMsg)
			// 	logger.PfcpLog.Infof("after ProcessPfcpAssociationReleaseRsp pMsg type %T\n", pMsg)
			// }
		}
	*/
	udp.SendPfcp(message, addr, nil)
}

func SendPfcpSessionModificationRequest(upNodeID pfcpType.NodeID,
	ctx *smf_context.SMContext,
	pdrList []*smf_context.PDR, farList []*smf_context.FAR, barList []*smf_context.BAR, qerList []*smf_context.QER) (seqNum uint32) {
	pfcpMsg, err := BuildPfcpSessionModificationRequest(upNodeID, ctx, pdrList, farList, barList, qerList)
	if err != nil {
		ctx.SubPfcpLog.Errorf("Build PFCP Session Modification Request failed: %v", err)
		return
	}

	seqNum = getSeqNumber()
	nodeIDtoIP := upNodeID.ResolveNodeIdToIp().String()
	remoteSEID := ctx.PFCPContext[nodeIDtoIP].RemoteSEID
	message := pfcp.Message{
		Header: pfcp.Header{
			Version:         pfcp.PfcpVersion,
			MP:              1,
			S:               pfcp.SEID_PRESENT,
			MessageType:     pfcp.PFCP_SESSION_MODIFICATION_REQUEST,
			SEID:            remoteSEID,
			SequenceNumber:  seqNum,
			MessagePriority: 12,
		},
		Body: pfcpMsg,
	}

	upaddr := &net.UDPAddr{
		IP:   upNodeID.ResolveNodeIdToIp(),
		Port: pfcpUdp.PFCP_PORT,
	}

	eventData := pfcpUdp.PfcpEventData{LSEID: ctx.PFCPContext[nodeIDtoIP].LocalSEID, ErrHandler: HandlePfcpSendError}

	if udp_adapter {
		rsp, err := SendPfcpMsgToAdapter(upNodeID, message, upaddr, eventData)
		if err != nil {
			logger.PfcpLog.Errorf("in SendPfcpSessionModificationRequest SendPfcpMsgToAdapter failed: %v", err)
			return
		}
		if rsp.StatusCode == http.StatusOK {
			// process response
			pMsg := ProcessPfcpSessionModificationRsp(rsp)
			logger.PfcpLog.Infof("after ProcessPfcpSessionModificationRsp pMsg val %v\n", pMsg)
			logger.PfcpLog.Infof("after ProcessPfcpSessionModificationRsp pMsg type %T\n", pMsg)
		}
	}

	udp.SendPfcp(message, upaddr, eventData)
	ctx.SubPfcpLog.Infof("Sent PFCP Session Modify Request to NodeID[%s]", upNodeID.ResolveNodeIdToIp().String())
	return seqNum
}

// Deprecated: PFCP Session Modification Procedure should be initiated by the CP function
func SendPfcpSessionModificationResponse(addr *net.UDPAddr) {
	pfcpMsg, err := BuildPfcpSessionModificationResponse()
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP Session Modification Response failed: %v", err)
		return
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:         pfcp.PfcpVersion,
			MP:              1,
			S:               pfcp.SEID_PRESENT,
			MessageType:     pfcp.PFCP_SESSION_MODIFICATION_RESPONSE,
			SEID:            123456789123456789,
			SequenceNumber:  1,
			MessagePriority: 12,
		},
		Body: pfcpMsg,
	}
	// TODO -> need ip to nodeID
	udp.SendPfcp(message, addr, nil)
}

func SendPfcpSessionDeletionRequest(upNodeID pfcpType.NodeID, ctx *smf_context.SMContext) (seqNum uint32) {
	pfcpMsg, err := BuildPfcpSessionDeletionRequest(upNodeID, ctx)
	if err != nil {
		ctx.SubPfcpLog.Errorf("Build PFCP Session Deletion Request failed: %v", err)
		return
	}
	seqNum = getSeqNumber()
	nodeIDtoIP := upNodeID.ResolveNodeIdToIp().String()
	remoteSEID := ctx.PFCPContext[nodeIDtoIP].RemoteSEID
	message := pfcp.Message{
		Header: pfcp.Header{
			Version:         pfcp.PfcpVersion,
			MP:              1,
			S:               pfcp.SEID_PRESENT,
			MessageType:     pfcp.PFCP_SESSION_DELETION_REQUEST,
			SEID:            remoteSEID,
			SequenceNumber:  seqNum,
			MessagePriority: 12,
		},
		Body: pfcpMsg,
	}

	upaddr := &net.UDPAddr{
		IP:   upNodeID.ResolveNodeIdToIp(),
		Port: pfcpUdp.PFCP_PORT,
	}

	eventData := pfcpUdp.PfcpEventData{LSEID: ctx.PFCPContext[nodeIDtoIP].LocalSEID, ErrHandler: HandlePfcpSendError}
	if udp_adapter {
		rsp, err := SendPfcpMsgToAdapter(upNodeID, message, upaddr, eventData)
		if err != nil {
			logger.PfcpLog.Errorf("in SendPfcpSessionDeletionRequest SendPfcpMsgToAdapter failed: %v", err)
			return
		}
		if rsp.StatusCode == http.StatusOK {
			// process response
			pMsg := ProcessPfcpSessionDeletionRsp(rsp)
			logger.PfcpLog.Infof("after ProcessPfcpSessionDeletionRsp pMsg val %v\n", pMsg)
			logger.PfcpLog.Infof("after ProcessPfcpSessionDeletionRsp pMsg type %T\n", pMsg)
		}
	}
	udp.SendPfcp(message, upaddr, eventData)

	ctx.SubPfcpLog.Infof("Sent PFCP Session Delete Request to NodeID[%s]", upNodeID.ResolveNodeIdToIp().String())
	return seqNum
}

// Deprecated: PFCP Session Deletion Procedure should be initiated by the CP function
func SendPfcpSessionDeletionResponse(addr *net.UDPAddr) {
	pfcpMsg, err := BuildPfcpSessionDeletionResponse()
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP Session Deletion Response failed: %v", err)
		return
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:         pfcp.PfcpVersion,
			MP:              1,
			S:               pfcp.SEID_PRESENT,
			MessageType:     pfcp.PFCP_SESSION_DELETION_RESPONSE,
			SEID:            123456789123456789,
			SequenceNumber:  1,
			MessagePriority: 12,
		},
		Body: pfcpMsg,
	}
	// TODO: need ip -> nodeID transfer
	udp.SendPfcp(message, addr, nil)
}

func SendPfcpSessionReportResponse(addr *net.UDPAddr, cause pfcpType.Cause, pfcpSRflag pfcpType.PFCPSRRspFlags, seqFromUPF uint32, SEID uint64) {
	pfcpMsg, err := BuildPfcpSessionReportResponse(cause, pfcpSRflag)
	if err != nil {
		logger.PfcpLog.Errorf("Build PFCP Session Report Response failed: %v", err)
		return
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:        pfcp.PfcpVersion,
			MP:             0,
			S:              pfcp.SEID_PRESENT,
			MessageType:    pfcp.PFCP_SESSION_REPORT_RESPONSE,
			SequenceNumber: seqFromUPF,
			SEID:           SEID,
		},
		Body: pfcpMsg,
	}
	// TODO: need ip -> nodeID transfer
	udp.SendPfcp(message, addr, nil)
	logger.PfcpLog.Infof("Sent PFCP Session Report Response Seq[%d] to NodeID[%s]", seqFromUPF, addr.IP.String())
}

func SendHeartbeatResponse(addr *net.UDPAddr, seq uint32) {
	pfcpMsg := pfcp.HeartbeatResponse{
		RecoveryTimeStamp: &pfcpType.RecoveryTimeStamp{
			RecoveryTimeStamp: udp.ServerStartTime,
		},
	}

	message := pfcp.Message{
		Header: pfcp.Header{
			Version:        pfcp.PfcpVersion,
			MP:             0,
			S:              pfcp.SEID_NOT_PRESENT,
			MessageType:    pfcp.PFCP_HEARTBEAT_RESPONSE,
			SequenceNumber: seq,
		},
		Body: pfcpMsg,
	}
	// TODO: need ip -> nodeID transfer
	udp.SendPfcp(message, addr, nil)
	logger.PfcpLog.Infof("Sent PFCP Heartbeat Response Seq[%d] to NodeID[%s]", seq, addr.IP.String())
}

func HandlePfcpSendError(msg *pfcp.Message, pfcpErr error) {

	logger.PfcpLog.Errorf("send of PFCP msg [%v] failed, %v",
		pfcpmsgtypes.PfcpMsgTypeString(msg.Header.MessageType), pfcpErr.Error())
	metrics.IncrementN4MsgStats(smf_context.SMF_Self().NfInstanceID,
		pfcpmsgtypes.PfcpMsgTypeString(msg.Header.MessageType), "Out", "Failure", pfcpErr.Error())

	//Refresh SMF DNS Cache incase of any send failure(includes timeout)
	pfcpType.RefreshDnsHostIpCache()

	switch msg.Header.MessageType {
	case pfcp.PFCP_SESSION_ESTABLISHMENT_REQUEST:
		handleSendPfcpSessEstReqError(msg, pfcpErr)
	case pfcp.PFCP_SESSION_MODIFICATION_REQUEST:
		handleSendPfcpSessModReqError(msg, pfcpErr)
	case pfcp.PFCP_SESSION_DELETION_REQUEST:
		handleSendPfcpSessRelReqError(msg, pfcpErr)
	default:
		logger.PfcpLog.Errorf("Unable to send PFCP packet type [%v] and content [%v]",
			pfcpmsgtypes.PfcpMsgTypeString(msg.Header.MessageType), msg)
	}
}

func handleSendPfcpSessEstReqError(msg *pfcp.Message, pfcpErr error) {
	//Lets decode the PDU request
	pfcpEstReq, _ := msg.Body.(pfcp.PFCPSessionEstablishmentRequest)

	SEID := pfcpEstReq.CPFSEID.Seid
	smContext := smf_context.GetSMContextBySEID(SEID)
	smContext.SubPfcpLog.Errorf("PFCP Session Establishment send failure, %v", pfcpErr.Error())
	//N1N2 Request towards AMF
	n1n2Request := models.N1N2MessageTransferRequest{}

	//N1 Container Info
	n1MsgContainer := models.N1MessageContainer{
		N1MessageClass:   "SM",
		N1MessageContent: &models.RefToBinaryData{ContentId: "GSM_NAS"},
	}

	//N1N2 Json Data
	n1n2Request.JsonData = &models.N1N2MessageTransferReqData{PduSessionId: smContext.PDUSessionID}

	if smNasBuf, err := smf_context.BuildGSMPDUSessionEstablishmentReject(smContext,
		nasMessage.Cause5GSMRequestRejectedUnspecified); err != nil {
		smContext.SubPduSessLog.Errorf("Build GSM PDUSessionEstablishmentReject failed: %s", err)
	} else {
		n1n2Request.BinaryDataN1Message = smNasBuf
		n1n2Request.JsonData.N1MessageContainer = &n1MsgContainer
	}

	//Send N1N2 Reject request
	rspData, _, err := smContext.
		CommunicationClient.
		N1N2MessageCollectionDocumentApi.
		N1N2MessageTransfer(context.Background(), smContext.Supi, n1n2Request)
	smContext.ChangeState(smf_context.SmStateInit)
	smContext.SubCtxLog.Traceln("SMContextState Change State: ", smContext.SMContextState.String())
	if err != nil {
		smContext.SubPfcpLog.Warnf("Send N1N2Transfer failed")
	}
	if rspData.Cause == models.N1N2MessageTransferCause_N1_MSG_NOT_TRANSFERRED {
		smContext.SubPfcpLog.Warnf("%v", rspData.Cause)
	}
	smContext.SubPfcpLog.Errorf("PFCP send N1N2Transfer Reject initiated for id[%v], pduSessId[%v]", smContext.Identifier, smContext.PDUSessionID)

	//clear subscriber
	smf_context.RemoveSMContext(smContext.Ref)
}

func handleSendPfcpSessRelReqError(msg *pfcp.Message, pfcpErr error) {
	//Lets decode the PDU request
	pfcpRelReq, _ := msg.Body.(pfcp.PFCPSessionDeletionRequest)

	SEID := pfcpRelReq.CPFSEID.Seid
	smContext := smf_context.GetSMContextBySEID(SEID)
	if smContext != nil {
		smContext.SubPfcpLog.Errorf("PFCP Session Delete send failure, %v", pfcpErr.Error())
		//Always send success
		smContext.SBIPFCPCommunicationChan <- smf_context.SessionReleaseSuccess
	}
}

func handleSendPfcpSessModReqError(msg *pfcp.Message, pfcpErr error) {
	//Lets decode the PDU request
	pfcpModReq, _ := msg.Body.(pfcp.PFCPSessionModificationRequest)

	SEID := pfcpModReq.CPFSEID.Seid
	smContext := smf_context.GetSMContextBySEID(SEID)
	smContext.SubPfcpLog.Errorf("PFCP Session Modification send failure, %v", pfcpErr.Error())

	smContext.SBIPFCPCommunicationChan <- smf_context.SessionUpdateTimeout
}

type UdpPodMsgType int

type UdpPodPfcpMsg struct {
	SEID     string          `json:"seid" yaml:"seid" bson:"seid"`
	SmfIp    string          `json:"smfIp" yaml:"smfIp" bson:"smfIp"`
	UpNodeID pfcpType.NodeID `json:"upNodeID" yaml:"upNodeID" bson:"upNodeID"`
	// message type contains in Msg.Header
	Msg       pfcp.Message `json:"msg" yaml:"msg" bson:"msg"`
	Addr      *net.UDPAddr `json:"addr" yaml:"addr" bson:"addr"`
	EventData interface{}  `json:"eventData" yaml:"eventData" bson:"eventData"`
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func SeidConv(seid uint64) (seidStr string) {
	seidStr = strconv.FormatUint(seid, 16)
	return seidStr
}

func SendPfcpMsgToAdapter(upNodeID pfcpType.NodeID, msg pfcp.Message, addr *net.UDPAddr, eventData interface{}) (*http.Response, error) {
	// localVarHeaderParams := make(map[string]string)

	// get SEID
	seid := msg.Header.SEID
	// get IP
	ip_str := GetLocalIP()
	udpPodMsg := &UdpPodPfcpMsg{
		UpNodeID:  upNodeID,
		SEID:      SeidConv(seid),
		SmfIp:     ip_str,
		Msg:       msg,
		Addr:      addr,
		EventData: eventData,
	}
	if eventData != nil {
		// udpPodMsg.EventData = eventData.(pfcpUdp.PfcpEventData)
		udpPodMsg.EventData = nil
	}
	serverPort := 8090

	// BUG: TODO: cannot mashalled to json if eventData is not nil, result of json.Marshall is empty
	udpPodMsgJson, _ := json.Marshal(udpPodMsg)
	fmt.Println("udpPodMsgJson := ", udpPodMsgJson)
	fmt.Println("udpPodMsgJson := ", string(udpPodMsgJson))

	// change the IP here
	fmt.Printf("send to http://%s:%d\n", ip, serverPort)
	requestURL := fmt.Sprintf("http://%s:%d", ip, serverPort)
	jsonBody := []byte(udpPodMsgJson)

	bodyReader := bytes.NewReader(jsonBody)
	req, err := http.NewRequest(http.MethodPost, requestURL, bodyReader)
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := http.Client{
		Timeout: 30 * time.Second,
	}
	// waiting for http response
	rsp, err := client.Do(req)
	fmt.Printf("client: rsp: %v\n", rsp)
	fmt.Printf("client: rsp type: %T\n", rsp)

	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
	}

	return rsp, err
}
