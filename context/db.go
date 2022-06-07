	// db changes
	

// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
// Copyright 2019 free5GC.org
//
// SPDX-License-Identifier: Apache-2.0
//

package context

import (
	"encoding/json"
	"fmt"
	"github.com/badhrinathpa/MongoDBLibrary"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/free5gc/smf/logger"
	// "github.com/free5gc/openapi/models"

	"github.com/omec-project/idgenerator"
	"net"
	"time"
	"github.com/omec-project/smf/msgtypes/svcmsgtypes"
	"strconv"
)

const (
	SmContextDataColl = "smf.data.smContext"
	GTPTunnelDataColl = "smf.data.gtpTunnel"
	TransactionDataCol = "smf.data.transaction"
	SeidSmContextCol = "smf.data.seidSmContext"
	NodeInDBCol = "smf.data.nodeInDB"
)

// Transaction
type TransactionInDB struct {
	startTime time.Time
	endTime time.Time
	TxnId              uint32
	Priority           uint32
	Req                interface{}
	Rsp                interface{}
	Ctxt               interface{}
	// MsgType            svcmsgtypes.SmfMsgType
	MsgType				string
	CtxtKey            string
	Err                error
	// Status             chan bool
	NextTxnId          uint32
	// TxnFsmLog          *logrus.Entry
}

func (t *TransactionInDB) StartTime() time.Time {
	return t.startTime
}

func (t *TransactionInDB) EndTime() time.Time {
	return t.endTime
}


// type DataPathPoolInDB map[int64]DataPathInDB
type DataPathPoolInDB map[int64]*DataPathInDB

// UPTunnel
type UPTunnelInDB struct {
	PathIDGenerator *idgenerator.IDGenerator
	DataPathPool    DataPathPoolInDB
	ANInformation   struct {
		IPAddress net.IP
		TEID      uint32
	}
}


type DataPathInDB struct {
	// meta data
	Activated         bool
	IsDefaultPath     bool
	Destination       Destination
	HasBranchingPoint bool
	// Data Path Double Link List
	FirstDPNode DataPathNodeInDB
	// FirstDPNode *DataPathNodeInDB
}

// type DataPathNodeInDB  struct {
// 	// UPF UPF
// 	UPF *UPF
// 	// DataPathToAN *DataPathDownLink
// 	// DataPathToDN map[string]*DataPathUpLink //uuid to DataPathLink

// 	// UpLinkTunnel previously GTPTunnel
// 	UpLinkTunnelTEID   uint32
// 	// DownLinkTunnel previously GTPTunnel
// 	DownLinkTunnelTEID uint32

// 	UpLinkTunnel GTPTunnelInDB
// 	DownLinkTunnel GTPTunnelInDB

// 	// for UE Routing Topology
// 	// for special case:
// 	// branching & leafnode

// 	// InUse                bool
// 	IsBranchingPoint bool
// 	// DLDataPathLinkForPSA *DataPathUpLink
// 	// BPUpLinkPDRs         map[string]*DataPathDownLink // uuid to UpLink
// }

type GTPTunnelInDB struct {
	// SrcEndPoint previously DataPathNode
	SrcEndPointUpLinkTunnelTEID  uint32 
	SrcEndPointDownLinkTunnelTEID  uint32 
	// SrcEndPointUPF UPF
	SrcEndPointUPF *UPF
	SrcEndPointIsBranchingPoint bool

	// DestEndPoint
	DestEndPointUpLinkTunnelTEID  uint32 
	DestEndPointDownLinkTunnelTEID uint32
	// DestEndPointUPF UPF
	DestEndPointUPF *UPF
	DestEndPointIsBranchingPoint bool

	TEID uint32
	PDR  map[string]*PDR
}


func SetupSmfCollection() {
	fmt.Println("db - SetupSmfCollection!!")
	// MongoDBLibrary.SetMongoDB("sdcore", "mongodb://mongodb")
	// _, err := MongoDBLibrary.CreateIndex(SmContextDataColl, "supi")
	// if err != nil {
	// 	logger.CtxLog.Errorf("Create index failed on Supi field.")
	// }
	MongoDBLibrary.SetMongoDB("sdcore", "mongodb://mongodb")
	_, err := MongoDBLibrary.CreateIndex(SmContextDataColl, "ref")
	if err != nil {
		logger.CtxLog.Errorf("Create index failed on ref field.")
	}
	// MongoDBLibrary.SetMongoDB("sdcore", "mongodb://mongodb")
	// _, err = MongoDBLibrary.CreateIndex(SmContextDataColl, "pduSessionID")
	// if err != nil {
	// 	logger.CtxLog.Errorf("Create index failed on PDUSessionID field.")
	// }

	MongoDBLibrary.SetMongoDB("sdcore", "mongodb://mongodb")
	_, err = MongoDBLibrary.CreateIndex(GTPTunnelDataColl, "teid")
	if err != nil {
		logger.CtxLog.Errorf("Create index failed on teid field.")
	}
	MongoDBLibrary.SetMongoDB("sdcore", "mongodb://mongodb")
	_, err = MongoDBLibrary.CreateIndex(TransactionDataCol, "txnId")
	if err != nil {
		logger.CtxLog.Errorf("Create index failed on TxnId field.")
	}
	
	MongoDBLibrary.SetMongoDB("sdcore", "mongodb://mongodb")
	_, err = MongoDBLibrary.CreateIndex(SeidSmContextCol, "seid")
	if err != nil {
		logger.CtxLog.Errorf("Create index failed on TxnId field.")
	}
	
}


func (smContext *SMContext) MarshalJSON() ([]byte, error) {
	type Alias SMContext
	// UPTunnel: GTPTunnel -> TEID

	// SBIPFCPCommunicationChan omit, and make a new one when needed

	// ActiveTxn: transaction -> TxnId
	fmt.Println("db - in MarshalJSON ")
	fmt.Println("db - in MarshalJSON smContext.ActiveTxn.NextTxn %v", smContext.ActiveTxn.NextTxn)
	fmt.Println("db - in MarshalJSON smContext.ActiveTxn.Req %v", smContext.ActiveTxn.Req)
	fmt.Println("db - in MarshalJSON smContext.ActiveTxn.Rsp %v", smContext.ActiveTxn.Rsp)
	fmt.Println("db - in MarshalJSON smContext.ActiveTxn.Ctxt %v", smContext.ActiveTxn.Ctxt)
	fmt.Println("db - end of logging")
	var msgTypeVal string
	msgTypeVal = svcmsgtypes.SmfMsgTypeString(smContext.ActiveTxn.MsgType)
	fmt.Println("db - in MarshalJSON ", msgTypeVal)
	activeTxnVal := TransactionInDB{
		startTime: smContext.ActiveTxn.StartTime(),
		endTime: smContext.ActiveTxn.EndTime(),
		TxnId: smContext.ActiveTxn.TxnId,
		Priority: smContext.ActiveTxn.Priority,
		// Req: smContext.ActiveTxn.Req,
		// Rsp: smContext.ActiveTxn.Rsp,
		// Ctxt: smContext.ActiveTxn.Ctxt,
		MsgType: msgTypeVal,
		CtxtKey: smContext.ActiveTxn.CtxtKey,
		Err: smContext.ActiveTxn.Err,
		// Status: smContext.ActiveTxn.Status,
		// NextTxnId: smContext.ActiveTxn.NextTxn.TxnId,
		// TxnFsmLog: smContext.ActiveTxn.TxnFsmLog,
	}
	if smContext.ActiveTxn.Req != nil {
		activeTxnVal.Req = smContext.ActiveTxn.Req
	}
	if smContext.ActiveTxn.Rsp != nil {
		activeTxnVal.Rsp = smContext.ActiveTxn.Rsp
	}
	if smContext.ActiveTxn.Ctxt != nil {
		activeTxnVal.Ctxt = smContext.ActiveTxn.Ctxt
	}
	// tmpVal := activeTxnVal.NextTxn
	// while tmpVal != nil {
	// 	activeTxnVal.NextTxnId = smContext.ActiveTxn.NextTxn.TxnId
	// }

	// store transaction in DB
	// StoreTxnInDB(&activeTxnVal)

	return json.Marshal(&struct {
		// UPTunnel
		ActiveTxn	TransactionInDB	`json:"activeTxn"`
		// Tunnel	UPTunnelInDB	`json:"upTunnel"`
		*Alias
	}{
		// UPTunnel: upTunnelVal
		ActiveTxn: activeTxnVal,
		// Tunnel: upTunnelVal,
		Alias:       (*Alias)(smContext),
	})
}

func (smContext *SMContext) UnmarshalJSON(data []byte) error {
	type Alias SMContext
	aux := &struct {
		ActiveTxn	TransactionInDB	`json:"activeTxn"`
		// Tunnel	UPTunnelInDB	`json:"upTunnel"`
		*Alias
	}{
		Alias: (*Alias)(smContext),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		fmt.Println("Err in customized unMarshall!!")
		return err
	}
	// recover ActiveTxn
	smContext.ActiveTxn.SetStartTime(aux.ActiveTxn.StartTime())
	smContext.ActiveTxn.SetEndTime(aux.ActiveTxn.EndTime())
	smContext.ActiveTxn.TxnId = aux.ActiveTxn.TxnId
	smContext.ActiveTxn.Priority = aux.ActiveTxn.Priority
	// smContext.ActiveTxn.Req = &aux.ActiveTxn.Req
	// smContext.ActiveTxn.Rsp = &aux.ActiveTxn.Rsp
	// smContext.ActiveTxn.Ctxt = &aux.ActiveTxn.Ctxt
	smContext.ActiveTxn.MsgType = svcmsgtypes.SmfMsgTypeType(aux.ActiveTxn.MsgType)
	smContext.ActiveTxn.CtxtKey = aux.ActiveTxn.CtxtKey
	smContext.ActiveTxn.Err = aux.ActiveTxn.Err
	// smContext.ActiveTxn.Status = aux.ActiveTxn.Status
	// // need to retrieve next txn in db
	// // smContext.ActiveTxn.NextTxnId = &aux.ActiveTxn.NextTxnId
	// smContext.ActiveTxn.TxnFsmLog = aux.ActiveTxn.TxnFsmLog

	// recover Tunnel

	// recover Tunnel.DataPathPool

	// recover Tunnel.DataPathPool[*] - DataPath

	// recover Tunnel.DataPathPool[*].DataPathNode

	// recover Tunnel.DataPathPool[*].DataPathNode.UpLinkTunnel/DownLinkTunnel

	// recover Tunnel.DataPathPool[*].DataPathNode.UpLinkTunnel/DownLinkTunnel.SrcEndPoint/DestEndPoint

	return nil
}



func ToBsonM(data *SMContext) (ret bson.M) {

	// Marshal data into json format
	tmp, err := json.Marshal(data)
	if err != nil {
		logger.CtxLog.Errorf("SMContext marshall error: %v", err)
	}

	// unmarshal data into bson format
	err = json.Unmarshal(tmp, &ret)
	if err != nil {
		logger.CtxLog.Errorf("SMContext unmarshall error: %v", err)
	}

	return
}
func ToBsonMSeidRef(data SeidSmContextRef) (ret bson.M) {

	// Marshal data into json format
	tmp, err := json.Marshal(data)
	if err != nil {
		logger.CtxLog.Errorf("SMContext marshall error: %v", err)
	}

	// unmarshal data into bson format
	err = json.Unmarshal(tmp, &ret)
	if err != nil {
		logger.CtxLog.Errorf("SMContext unmarshall error: %v", err)
	}

	return
}

func StoreContextInDB(smContext *SMContext) {
	fmt.Println("db - Store SMContext In DB w ref!!")
	smContextBsonA := ToBsonM(smContext)
	filter := bson.M{"ref": smContext.Ref}
	logger.CtxLog.Infof("filter : ", filter)

	MongoDBLibrary.RestfulAPIPost(SmContextDataColl, filter, smContextBsonA)
	fmt.Println("db - finished Store SMContext In DB!!")
}

type SeidSmContextRef struct {
	Ref string `json:"ref" yaml:"ref" bson:"ref"`
	Seid uint64 `json:"seid" yaml:"seid" bson:"seid"`
}

func StoreSeidContextInDB(seid uint64, smContext *SMContext) {
	fmt.Println("db - StoreSeidContextInDB In DB!!")
	item := SeidSmContextRef{
		Ref: smContext.Ref,
		Seid: seid,
	}
	itemBsonA := ToBsonMSeidRef(item)
	filter := bson.M{"seid": seid}
	logger.CtxLog.Infof("filter : ", filter)

	MongoDBLibrary.RestfulAPIPost(SeidSmContextCol, filter, itemBsonA)
	fmt.Println("db - finished StoreSeidContextInDB In DB!!")
}

func GetSMContextByRefInDB(ref string) {
	filter := bson.M{}
	filter["ref"] = ref

	result := MongoDBLibrary.RestfulAPIGetOne(SmContextDataColl, filter)

	fmt.Println("GetSMContextByRefInDB, smf state json : ", result)

	// TBA Return SM context, reconstruct SM context from DB json
}

func GetSMContextBySEIDInDB(seid uint64){
	filter := bson.M{}
	filter["seid"] = seid
	
	result := MongoDBLibrary.RestfulAPIGetOne(SeidSmContextCol, filter)
	seidStr := strconv.FormatInt(int64(seid), 10)
	ref := result[seidStr].(string)
	fmt.Println("GetSMContextBySEIDInDB, ref json : ", ref)

	GetSMContextByRefInDB(ref)
	// TBA Return SM context, reconstruct SM context from DB json

}

func DeleteContextInDB(smContext *SMContext) {
	fmt.Println("db - delete SMContext In DB w pduSessionID!!")
	// smContextBsonA := ToBsonM(smContext)
	filter := bson.M{"ref": smContext.Ref}
	logger.CtxLog.Infof("filter : ", filter)

	MongoDBLibrary.RestfulAPIDeleteOne(SmContextDataColl, filter)
	fmt.Println("db - finished delete SMContext In DB!!")
}

func mapToByte(data map[string]interface{}) (ret []byte) {
	ret, _ = json.Marshal(data)
	return
}



