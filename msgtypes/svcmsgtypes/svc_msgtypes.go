// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package svcmsgtypes

type SmfMsgType string

// List of Msgs
const (

	//N11 Service
	MsgTypeNone           SmfMsgType = "none"
	CreateSmContext       SmfMsgType = "CreateSmContext"
	UpdateSmContext       SmfMsgType = "UpdateSmContext"
	ReleaseSmContext      SmfMsgType = "ReleaseSmContext"
	NotifySmContextStatus SmfMsgType = "NotifySmContextStatus"
	RetrieveSmContext     SmfMsgType = "RetrieveSmContext"
	NsmfPDUSessionCreate  SmfMsgType = "Create"  //Create a PDU session in the H-SMF
	NsmfPDUSessionUpdate  SmfMsgType = "Update"  //Update a PDU session in the H-SMF or V- SMF
	NsmfPDUSessionRelease SmfMsgType = "Release" //Release a PDU session in the H-SMF

	//NNRF_NFManagement
	NnrfNFRegister           SmfMsgType = "NfRegister"
	NnrfNFDeRegister         SmfMsgType = "NfDeRegister"
	NnrfNFInstanceDeRegister SmfMsgType = "NnrfNFInstanceDeRegister"
	NnrfNFDiscoveryUdm       SmfMsgType = "NfDiscoveryUdm"
	NnrfNFDiscoveryPcf       SmfMsgType = "NfDiscoveryPcf"
	NnrfNFDiscoveryAmf       SmfMsgType = "NfDiscoveryAmf"

	//NUDM_
	SmSubscriptionDataRetrieval SmfMsgType = "SmSubscriptionDataRetrieval"

	//NPCF_
	SmPolicyAssociationCreate       SmfMsgType = "SmPolicyAssociationCreate"
	SmPolicyAssociationDelete       SmfMsgType = "SmPolicyAssociationDelete"
	SmPolicyUpdateNotification      SmfMsgType = "SmPolicyUpdateNotification"
	SmPolicyTerminationNotification SmfMsgType = "SmPolicyTerminationNotification"

	//AMF_
	N1N2MessageTransfer                    SmfMsgType = "N1N2MessageTransfer"
	N1N2MessageTransferFailureNotification SmfMsgType = "N1N2MessageTransferFailureNotification"

	//PFCP
	PfcpSessCreate  SmfMsgType = "PfcpSessCreate"
	PfcpSessModify  SmfMsgType = "PfcpSessModify"
	PfcpSessRelease SmfMsgType = "PfcpSessRelease"
)

var msgTypeText = map[SmfMsgType]string{
	//N11 Service
	MsgTypeNone:	"none",
	CreateSmContext:	"CreateSmContext",
	UpdateSmContext:	"UpdateSmContext",
	ReleaseSmContext:"ReleaseSmContext",
	NotifySmContextStatus :"NotifySmContextStatus",
	RetrieveSmContext     :"RetrieveSmContext",
	NsmfPDUSessionCreate  :"Create",  //Create a PDU session in the H-SMF
	NsmfPDUSessionUpdate  :"Update",  //Update a PDU session in the H-SMF or V- SMF
	NsmfPDUSessionRelease :"Release", //Release a PDU session in the H-SMF

	//NNRF_NFManagement
	NnrfNFRegister:"NfRegister",
	NnrfNFDeRegister:"NfDeRegister",
	NnrfNFInstanceDeRegister :"NnrfNFInstanceDeRegister",
	NnrfNFDiscoveryUdm :"NfDiscoveryUdm",
	NnrfNFDiscoveryPcf :"NfDiscoveryPcf",
	NnrfNFDiscoveryAmf :"NfDiscoveryAmf",

	//NUDM_
	SmSubscriptionDataRetrieval :"SmSubscriptionDataRetrieval",

	//NPCF_
	SmPolicyAssociationCreate :"SmPolicyAssociationCreate",
	SmPolicyAssociationDelete :"SmPolicyAssociationDelete",
	SmPolicyUpdateNotification:"SmPolicyUpdateNotification",
	SmPolicyTerminationNotification :"SmPolicyTerminationNotification",

	//AMF_
	N1N2MessageTransfer:"N1N2MessageTransfer",
	N1N2MessageTransferFailureNotification :"N1N2MessageTransferFailureNotification",

	//PFCP
	PfcpSessCreate  :"PfcpSessCreate",
	PfcpSessModify  :"PfcpSessModify",
	PfcpSessRelease :"PfcpSessRelease",
}

var msgTypeTextString = map[string]SmfMsgType {
	//N11 Service
	"none": MsgTypeNone,
	"CreateSmContext":CreateSmContext,
	"UpdateSmContext":UpdateSmContext,
	"ReleaseSmContext":ReleaseSmContext,
	"NotifySmContextStatus":NotifySmContextStatus,
	"RetrieveSmContext":RetrieveSmContext,
	"Create":NsmfPDUSessionCreate,  //Create a PDU session in the H-SMF
	"Update": NsmfPDUSessionUpdate,//Update a PDU session in the H-SMF or V- SMF
	"Release": NsmfPDUSessionRelease, //Release a PDU session in the H-SMF

	//NNRF_NFManagement
	"NfRegister":NnrfNFRegister,
	"NfDeRegister":NnrfNFDeRegister,
	"NnrfNFInstanceDeRegister": NnrfNFInstanceDeRegister,
	"NfDiscoveryUdm":NnrfNFDiscoveryUdm,
	"NfDiscoveryPcf":NnrfNFDiscoveryPcf,
	"NfDiscoveryAmf":NnrfNFDiscoveryAmf,

	//NUDM_
	"SmSubscriptionDataRetrieval":SmSubscriptionDataRetrieval,

	//NPCF_
	"SmPolicyAssociationCreate":SmPolicyAssociationCreate,
	"SmPolicyAssociationDelete":SmPolicyAssociationDelete,
	"SmPolicyUpdateNotification":SmPolicyUpdateNotification,
	"SmPolicyTerminationNotification":SmPolicyTerminationNotification,

	//AMF_
	"N1N2MessageTransfer":N1N2MessageTransfer,
	"N1N2MessageTransferFailureNotification": N1N2MessageTransferFailureNotification,

	//PFCP
	"PfcpSessCreate":PfcpSessCreate,
	"PfcpSessModify":PfcpSessModify,
	"PfcpSessRelease":PfcpSessRelease,
}

func SmfMsgTypeString(code SmfMsgType) string {
	return msgTypeText[code]
}

func SmfMsgTypeType(code string) SmfMsgType {
	return msgTypeTextString[code]
}

