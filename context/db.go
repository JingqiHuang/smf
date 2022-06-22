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

	"github.com/omec-project/pfcp/pfcpType"
	"github.com/omec-project/smf/logger"

	// "github.com/free5gc/openapi/models"

	"net"
	"reflect"
	"strconv"
	"time"

	"github.com/omec-project/idgenerator"
)

const (
	SmContextDataColl  = "smf.data.smContext"
	TransactionDataCol = "smf.data.transaction"
	SeidSmContextCol   = "smf.data.seidSmContext"
	NodeInDBCol        = "smf.data.nodeInDB"
	SMPolicyClientCol  = "smf.data.smPolicyClient"
)

type Change struct {
	Type string      // The type of change detected; can be one of create, update or delete
	Path []string    // The path of the detected change; will contain any field name or array index that was part of the traversal
	From interface{} // The original value that was present in the "from" structure
	To   interface{} // The new value that was detected as a change in the "to" structure
}

// Transaction
type TransactionInDB struct {
	startTime time.Time
	endTime   time.Time
	TxnId     uint32
	Priority  uint32
	Req       interface{}
	Rsp       interface{}
	CtxtRef   string
	// MsgType            svcmsgtypes.SmfMsgType
	MsgType string
	CtxtKey string
	Err     error
	// Status             chan bool
	NextTxnId uint32
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
	FirstDPNode *DataPathNodeInDB
	// FirstDPNode *DataPathNodeInDB
}

type DataPathNodeInDB struct {
	// UPF *UPF
	DataPathNodeUPFNodeID NodeIDInDB

	ULTunnelInfo *TunnelInfo
	DLTunnelInfo *TunnelInfo

	IsBranchingPoint bool
}
type TunnelInfo struct {
	DataPathNodeUPFNodeID NodeIDInDB
	TEID                  uint32
	PDR                   map[string]*PDR
}

type NodeIDInDB struct {
	NodeIdType  uint8 // 0x00001111
	NodeIdValue []byte
}

type PFCPSessionContextInDB struct {
	PDRs       map[uint16]*PDR
	NodeID     pfcpType.NodeID
	LocalSEID  string
	RemoteSEID string
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
	// testDB()

}

type PFCPContextInDB map[string]PFCPSessionContextInDB

func (smContext *SMContext) MarshalJSON() ([]byte, error) {
	type Alias SMContext
	// UPTunnel: GTPTunnel -> TEID
	// fmt.Println("db - in customized MarshalJSON ")

	/*
		var msgTypeVal string

		var activeTxnVal TransactionInDB
		if smContext.ActiveTxn != nil {
			msgTypeVal = svcmsgtypes.SmfMsgTypeString(smContext.ActiveTxn.MsgType)
			fmt.Println("db - in customized MarshalJSON ", msgTypeVal)
			activeTxnVal.startTime = smContext.ActiveTxn.StartTime()
			activeTxnVal.endTime = smContext.ActiveTxn.EndTime()
			activeTxnVal.TxnId = smContext.ActiveTxn.TxnId
			activeTxnVal.Priority = smContext.ActiveTxn.Priority
			activeTxnVal.Req = smContext.ActiveTxn.Req
			activeTxnVal.Rsp = smContext.ActiveTxn.Rsp
			activeTxnVal.MsgType = msgTypeVal
			activeTxnVal.CtxtKey = smContext.ActiveTxn.CtxtKey
			activeTxnVal.Err = smContext.ActiveTxn.Err
			if smContext.ActiveTxn.NextTxn != nil {
				activeTxnVal.NextTxnId = smContext.ActiveTxn.NextTxn.TxnId
			}
			if smContext.ActiveTxn.Ctxt != nil {
				activeTxnVal.CtxtRef = smContext.ActiveTxn.Ctxt.(*SMContext).Ref
			}
			StoreTxnInDB(&activeTxnVal)
		}*/

	// fmt.Println("db - in MarshalJSON after activeTxnVal")

	dataPathPoolInDBVal := make(map[int64]*DataPathInDB)

	var dataPathInDBIf interface{}
	var FirstDPNodeIf interface{}

	var upTunnelVal UPTunnelInDB
	if smContext.Tunnel != nil {
		upTunnelVal.ANInformation = smContext.Tunnel.ANInformation

		if smContext.Tunnel.DataPathPool != nil {
			// fmt.Println("db - smContext.Tunnel.DataPathPool len ",len(smContext.Tunnel.DataPathPool))
			for key, val := range smContext.Tunnel.DataPathPool {

				dataPathInDBIf = val
				dataPath := dataPathInDBIf.(*DataPath)

				firstDPNode := dataPath.FirstDPNode
				FirstDPNodeIf = firstDPNode

				dataPathNode := FirstDPNodeIf.(*DataPathNode)
				// fmt.Println("db - in MarshalJSON dataPathNode from dataPath.FirstDPNode %v", dataPathNode)
				// fmt.Println("db - in MarshalJSON dataPath %v", dataPath)
				// convert dataPathNode to DataPathNodeInDB
				dataPathNodeInDBVal := StoreDataPathNode(dataPathNode)
				// dataPathNodeInDBVal := &DataPathNodeInDB{}

				// fmt.Println("db - in MarshalJSON dataPathNodeInDBVal %v", dataPathNodeInDBVal)

				newDataPathInDB := &DataPathInDB{
					Activated:         dataPath.Activated,
					IsDefaultPath:     dataPath.IsDefaultPath,
					Destination:       dataPath.Destination,
					HasBranchingPoint: dataPath.HasBranchingPoint,
					FirstDPNode:       dataPathNodeInDBVal,
				}

				dataPathPoolInDBVal[key] = newDataPathInDB
			}
			upTunnelVal.DataPathPool = dataPathPoolInDBVal
		}
	}

	var pfcpSessionContextInDB PFCPSessionContextInDB
	PFCPContextVal := make(PFCPContextInDB)
	// store localseid and remoteseid
	for key, pfcpCtx := range smContext.PFCPContext {
		pfcpSessionContextInDB.NodeID = pfcpCtx.NodeID
		pfcpSessionContextInDB.PDRs = pfcpCtx.PDRs
		fmt.Println("in customized marshalling pfcpCtx.PDRs ", pfcpCtx.PDRs)
		fmt.Println("in customized marshalling pfcpSessionContextInDB.PDRs ", pfcpSessionContextInDB.PDRs)
		pfcpSessionContextInDB.LocalSEID = strconv.FormatUint(pfcpCtx.LocalSEID, 10)
		pfcpSessionContextInDB.RemoteSEID = strconv.FormatUint(pfcpCtx.RemoteSEID, 10)
		PFCPContextVal[key] = pfcpSessionContextInDB
	}
	fmt.Println("in customized marshalling PFCPContextVal ", PFCPContextVal)

	// fmt.Println("db - in MarshalJSON before return")
	// fmt.Println("db - in MarshalJSON smContext ", smContext)
	return json.Marshal(&struct {
		// ActiveTxn   TransactionInDB `json:"activeTxn"`
		Tunnel      UPTunnelInDB    `json:"tunnel"`
		PFCPContext PFCPContextInDB `json:"pfcpContext"`
		*Alias
	}{
		// ActiveTxn:   activeTxnVal,
		Tunnel:      upTunnelVal,
		PFCPContext: PFCPContextVal,
		Alias:       (*Alias)(smContext),
	})
}

func (smContext *SMContext) UnmarshalJSON(data []byte) error {
	fmt.Println("db - in UnmarshalJSON")
	type Alias SMContext
	aux := &struct {
		// ActiveTxn      TransactionInDB `json:"activeTxn"`
		Tunnel         UPTunnelInDB    `json:"tunnel"`
		PFCPContextVal PFCPContextInDB `json:"pfcpContext"`
		*Alias
	}{
		Alias: (*Alias)(smContext),
	}
	// fmt.Println("db - before Unmarshal in customized unMarshall")
	if err := json.Unmarshal(data, &aux); err != nil {
		fmt.Println("Err in customized unMarshall!!")
		return err
	}

	fmt.Println("db - in UnmarshalJSON recovering smContext.UpCnxState = ", smContext.UpCnxState)

	// recover smContext.PFCPContext
	smContext.PFCPContext = make(map[string]*PFCPSessionContext)
	fmt.Println("in customized unmarshalling aux.PFCPContextVal ", aux.PFCPContextVal)
	for key, pfcpCtxInDB := range aux.PFCPContextVal {
		smContext.PFCPContext[key] = &PFCPSessionContext{}
		smContext.PFCPContext[key].NodeID = pfcpCtxInDB.NodeID
		smContext.PFCPContext[key].PDRs = pfcpCtxInDB.PDRs
		fmt.Println("in customized unmarshalling key ", key)
		fmt.Println("in customized unmarshalling pfcpCtxInDB.PDRs ", pfcpCtxInDB.PDRs)
		fmt.Println("in customized unmarshalling smContext.PFCPContext[key].PDRs ", smContext.PFCPContext[key].PDRs)
		smContext.PFCPContext[key].LocalSEID, _ = strconv.ParseUint(string(pfcpCtxInDB.LocalSEID), 10, 64)
		smContext.PFCPContext[key].RemoteSEID, _ = strconv.ParseUint(string(pfcpCtxInDB.RemoteSEID), 10, 64)
	}
	/*
		smContext.ActiveTxn = &transaction.Transaction{}
		// recover ActiveTxn
		smContext.ActiveTxn.SetStartTime(aux.ActiveTxn.StartTime())
		smContext.ActiveTxn.SetEndTime(aux.ActiveTxn.EndTime())
		smContext.ActiveTxn.TxnId = aux.ActiveTxn.TxnId
		smContext.ActiveTxn.Priority = aux.ActiveTxn.Priority
		smContext.ActiveTxn.Req = aux.ActiveTxn.Req
		smContext.ActiveTxn.Rsp = aux.ActiveTxn.Rsp
		smContext.ActiveTxn.MsgType = svcmsgtypes.SmfMsgTypeType(aux.ActiveTxn.MsgType)
		smContext.ActiveTxn.CtxtKey = aux.ActiveTxn.CtxtKey
		smContext.ActiveTxn.Err = aux.ActiveTxn.Err
		// fmt.Println("db - after Unmarshal in customized unMarshall finished recovered ActiveTxn %v", smContext.ActiveTxn)
		subField := logrus.Fields{"txnid": smContext.ActiveTxn.TxnId, "txntype": string(smContext.ActiveTxn.MsgType), "ctxtkey": smContext.ActiveTxn.CtxtKey}
		smContext.ActiveTxn.TxnFsmLog = logger.TxnFsmLog.WithFields(subField)
		smContext.ActiveTxn.Status = make(chan bool)
		// TODO: need to retrieve next txn in db
		// smContext.ActiveTxn.NextTxnId = &aux.ActiveTxn.NextTxnId
		smContext.ActiveTxn.NextTxn = nil
	*/
	// fmt.Println("db - after Unmarshal in customized unMarshall start recover Tunnel %v", aux.Tunnel)
	var dataPathInDBIf interface{}
	var FirstDPNodeIf interface{}
	var nilVal *UPTunnelInDB = nil
	smContext.Tunnel = &UPTunnel{}
	if &aux.Tunnel != nilVal {
		smContext.Tunnel.ANInformation = aux.Tunnel.ANInformation
		smContext.Tunnel.PathIDGenerator = idgenerator.NewGenerator(1, 2147483647)
		smContext.Tunnel.DataPathPool = NewDataPathPool()
		fmt.Println("aux.Tunnel.DataPathPool %v", aux.Tunnel.DataPathPool)
		for key, val := range aux.Tunnel.DataPathPool {
			dataPathInDBIf = val
			dataPathInDB := dataPathInDBIf.(*DataPathInDB)
			// fmt.Println("dataPathInDB - dataPathInDBIf %v", *dataPathInDB)
			// fmt.Println("dataPathInDB - dataPathInDBIf dataPathInDB.FirstDPNode %v", *dataPathInDB.FirstDPNode)
			// fmt.Println("dataPathInDB - dataPathInDBIf dataPathInDB.FirstDPNode.DataPathNodeUPFNodeID %v", dataPathInDB.FirstDPNode.DataPathNodeUPFNodeID)
			firstDPNode := dataPathInDB.FirstDPNode
			FirstDPNodeIf = firstDPNode
			dataPathNodeInDBVal := FirstDPNodeIf.(*DataPathNodeInDB)
			dataPathNodeVal := RecoverDataPathNode(dataPathNodeInDBVal)
			// fmt.Println("dataPathInDB - dataPathNodeVal after RecoverDataPathNode %v", dataPathNodeVal)

			newDataPath := NewDataPath()

			newDataPath.Activated = dataPathInDB.Activated
			newDataPath.IsDefaultPath = dataPathInDB.IsDefaultPath
			newDataPath.Destination = dataPathInDB.Destination
			newDataPath.HasBranchingPoint = dataPathInDB.HasBranchingPoint

			newDataPath.FirstDPNode = dataPathNodeVal

			smContext.Tunnel.DataPathPool[key] = newDataPath

			// fmt.Println("dataPathInDB - newDataPath after smContext.Tunnel.DataPathPool[key] %v", newDataPath)
		}

		// fmt.Println("smContext.Tunnel.DataPathPool %v", smContext.Tunnel.DataPathPool)
	}
	// recover logs
	smContext.initLogTags()
	// recover SBIPFCPCommunicationChan
	smContext.SBIPFCPCommunicationChan = make(chan PFCPSessionResponseStatus, 1)
	// smContext.ActiveTxn.Ctxt = smContext
	// smContext.ActiveTxn.CtxtKey = smContext.Ref

	fmt.Println("db - after Unmarshal in customized unMarshall after recover smContext ", smContext)

	return nil
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

func ToBsonM(data *SMContext) (ret bson.M) {

	// Marshal data into json format
	fmt.Println("db - in ToBsonM before marshal")
	tmp, err := json.Marshal(data)
	if err != nil {
		logger.CtxLog.Errorf("SMContext marshall error: %v", err)
	}
	// unmarshal data into bson format

	// fmt.Println("db - in ToBsonM after json marshal = ", tmp)
	err = json.Unmarshal(tmp, &ret)
	// fmt.Println("db - in ToBsonM after json unmarshal = ", &ret)
	if err != nil {
		logger.CtxLog.Errorf("SMContext unmarshall error: %v", err)
	}

	return
}

func StoreSmContextInDB(smContext *SMContext) {
	fmt.Println("db - Store SMContext In DB w ref!!")
	fmt.Println("db - in StoreSmContextInDB before ToBsonM smContext = ", smContext)
	fmt.Println("db - in StoreSmContextInDB before ToBsonM smContext.UpCnxState = ", smContext.UpCnxState)
	smContextBsonA := ToBsonM(smContext)
	fmt.Println("db - in StoreSmContextInDB after ToBsonM smContextBsonA = ", smContextBsonA)
	filter := bson.M{"ref": smContext.Ref}
	logger.CtxLog.Infof("filter: ", filter)

	MongoDBLibrary.RestfulAPIPost(SmContextDataColl, filter, smContextBsonA)

}

type SeidSmContextRef struct {
	Ref  string `json:"ref" yaml:"ref" bson:"ref"`
	Seid uint64 `json:"seid" yaml:"seid" bson:"seid"`
}

func StoreSeidContextInDB(seid uint64, smContext *SMContext) {

	item := SeidSmContextRef{
		Ref:  smContext.Ref,
		Seid: seid,
	}
	itemBsonA := ToBsonMSeidRef(item)
	filter := bson.M{"seid": seid}
	logger.CtxLog.Infof("filter : ", filter)

	MongoDBLibrary.RestfulAPIPost(SeidSmContextCol, filter, itemBsonA)
	fmt.Println("db - finished StoreSeidContextInDB In DB!!")
}

func GetSMContextByRefInDB(ref string) (smContext *SMContext) {
	fmt.Println("db - in GetSMContextByRefInDB")
	smContext = &SMContext{}
	filter := bson.M{}
	filter["ref"] = ref

	result := MongoDBLibrary.RestfulAPIGetOne(SmContextDataColl, filter)

	// fmt.Println("GetSMContextByRefInDB, smf state json : %v", result)
	// TBA Return SM context, reconstruct SM context from DB json
	err := json.Unmarshal(mapToByte(result), smContext)
	smContext.SMLock.Lock()
	defer smContext.SMLock.Unlock()

	if err != nil {
		logger.CtxLog.Errorf("smContext unmarshall error: %v", err)
		return nil
	}
	// fmt.Println("GetSMContextByRefInDB, before return : %v", smContext)
	return smContext
}

func GetSMContextBySEIDInDB(seid uint64) (smContext *SMContext) {
	filter := bson.M{}
	filter["seid"] = seid

	result := MongoDBLibrary.RestfulAPIGetOne(SeidSmContextCol, filter)
	ref := result["ref"].(string)
	fmt.Println("StoreSeidContextInDB, result string : ", ref)

	return GetSMContextByRefInDB(ref)

}

func DeleteContextInDBBySEID(seid uint64) {
	fmt.Println("db - delete SMContext In DB by seid!!")
	// smContextBsonA := ToBsonM(smContext)
	filter := bson.M{"seid": seid}
	logger.CtxLog.Infof("filter : ", filter)

	result := MongoDBLibrary.RestfulAPIGetOne(SeidSmContextCol, filter)
	ref := result["ref"].(string)
	fmt.Println("GetSMContextBySEIDInDB, ref string : ", ref)

	MongoDBLibrary.RestfulAPIDeleteOne(SeidSmContextCol, filter)
	DeleteContextInDBByRef(ref)
	fmt.Println("db - finished delete SMContext by seid In DB!!")

}

func DeleteContextInDBByRef(ref string) {
	fmt.Println("db - delete SMContext In DB w ref!!")
	// smContextBsonA := ToBsonM(smContext)
	filter := bson.M{"ref": ref}
	logger.CtxLog.Infof("filter : ", filter)
	MongoDBLibrary.RestfulAPIDeleteOne(SmContextDataColl, filter)
	fmt.Println("db - finished delete SMContext In DB!!")
}

func mapToByte(data map[string]interface{}) (ret []byte) {
	// fmt.Println("db - in mapToByte!! data: %v", data)
	ret, _ = json.Marshal(data)
	// fmt.Println("db - in mapToByte!! ret: %v", ret)
	return
}

// func testDB() {
// 	filter := bson.M{}
// 	filter["txnID"] = -1

// 	result := MongoDBLibrary.RestfulAPIGetOne(TransactionDataCol, filter)

// 	print("in db - testdb() result ", result)
// 	print("in db - testdb() result type ", reflect.TypeOf(result))
// }

func CompareSMContext(sourceSMCtxt *SMContext, destSMCtxt *SMContext) {

	res1 := reflect.DeepEqual(sourceSMCtxt.Ref, destSMCtxt.Ref)
	if res1 {
		fmt.Println("two Ref fields are equal")
	} else {
		fmt.Println("two Ref fields are not equal")
		fmt.Println("sourceSMCtxt values", sourceSMCtxt.Ref)
		fmt.Println("destSMCtxt values", destSMCtxt.Ref)
	}
	res1 = reflect.DeepEqual(sourceSMCtxt.PDUAddress, destSMCtxt.PDUAddress)
	if res1 {
		fmt.Println("two PDUAddress fields are equal")
	} else {
		fmt.Println("two PDUAddress fields are not equal")
		fmt.Println("sourceSMCtxt values", sourceSMCtxt.PDUAddress)
		fmt.Println("destSMCtxt values", destSMCtxt.PDUAddress)
		fmt.Println("sourceSMCtxt byte values %v", []byte(sourceSMCtxt.PDUAddress))
		fmt.Println("destSMCtxt byte values %v", []byte(destSMCtxt.PDUAddress))
	}
	res1 = reflect.DeepEqual(sourceSMCtxt.SMPolicyClient, destSMCtxt.SMPolicyClient)
	if res1 {
		fmt.Println("two SMPolicyClient fields are equal")
	} else {
		fmt.Println("two SMPolicyClient fields are not equal")
		fmt.Println("sourceSMCtxt values", sourceSMCtxt.SMPolicyClient)
		fmt.Println("destSMCtxt values", destSMCtxt.SMPolicyClient)
	}
	res1 = reflect.DeepEqual(sourceSMCtxt.CommunicationClient, destSMCtxt.CommunicationClient)
	if res1 {
		fmt.Println("two  CommunicationClient fields are equal")
	} else {
		fmt.Println("two CommunicationClient fields are not equal")
		fmt.Println("sourceSMCtxt values", sourceSMCtxt.CommunicationClient)
		fmt.Println("destSMCtxt values", destSMCtxt.CommunicationClient)
	}
	res1 = reflect.DeepEqual(sourceSMCtxt.Tunnel.DataPathPool, destSMCtxt.Tunnel.DataPathPool)
	if res1 {
		fmt.Println("two Tunnel.DataPathPool fields are equal")
	} else {
		fmt.Println("two Tunnel.DataPathPool fields are not equal")
		fmt.Println("sourceSMCtxt values ", sourceSMCtxt.Tunnel.DataPathPool)
		fmt.Println("destSMCtxt values ", destSMCtxt.Tunnel.DataPathPool)
		// fmt.Println("sourceSMCtxt values [1]", sourceSMCtxt.Tunnel.DataPathPool[1])
		// fmt.Println("destSMCtxt values [1]", destSMCtxt.Tunnel.DataPathPool[1])
		// fmt.Println("sourceSMCtxt ipv4 len", len(sourceSMCtxt.Tunnel.DataPathPool[1].FirstDPNode.UpLinkTunnel.PDR["ALLOW-ALL"].PDI.UEIPAddress.Ipv4Address))
		// fmt.Println("destSMCtxt ipv4 len", len(sourceSMCtxt.Tunnel.DataPathPool[1].FirstDPNode.UpLinkTunnel.PDR["ALLOW-ALL"].PDI.UEIPAddress.Ipv4Address))
		// fmt.Println("sourceSMCtxt test", sourceSMCtxt.Tunnel.DataPathPool[1].FirstDPNode.UpLinkTunnel.PDR["ALLOW-ALL"].PDI.UEIPAddress.Ipv4Address[0])
		// fmt.Println("destSMCtxt test", destSMCtxt.Tunnel.DataPathPool[1].FirstDPNode.UpLinkTunnel.PDR["ALLOW-ALL"].PDI.UEIPAddress.Ipv4Address[0])
		// // fmt.Println("sourceSMCtxt test", sourceSMCtxt.Tunnel.DataPathPool[1].FirstDPNode.UpLinkTunnel.PDR["ALLOW-ALL"].PDI.UEIPAddress.Ipv4Address[10])
		// // fmt.Println("destSMCtxt test", destSMCtxt.Tunnel.DataPathPool[1].FirstDPNode.UpLinkTunnel.PDR["ALLOW-ALL"].PDI.UEIPAddress.Ipv4Address[10])
		// fmt.Printf("sourceSMCtxt test byte mystr:\t %v \n", []byte(sourceSMCtxt.Tunnel.DataPathPool[1].FirstDPNode.UpLinkTunnel.PDR["ALLOW-ALL"].PDI.UEIPAddress.Ipv4Address))
		// fmt.Printf("destSMCtxt test byte mystr:\t %v \n", []byte(destSMCtxt.Tunnel.DataPathPool[1].FirstDPNode.UpLinkTunnel.PDR["ALLOW-ALL"].PDI.UEIPAddress.Ipv4Address))
	}
	res1 = reflect.DeepEqual(sourceSMCtxt.PFCPContext, destSMCtxt.PFCPContext)
	if res1 {
		fmt.Println("two PFCPContext fields are equal")
	} else {
		fmt.Println("two PFCPContext fields are not equal")
		fmt.Println("sourceSMCtxt values", sourceSMCtxt.PFCPContext)
		fmt.Println("destSMCtxt values", destSMCtxt.PFCPContext)
	}
	res1 = reflect.DeepEqual(sourceSMCtxt.UeLocation, destSMCtxt.UeLocation)
	if res1 {
		fmt.Println("two UeLocation fields are equal")
	} else {
		fmt.Println("two UeLocation fields are not equal")
		fmt.Println("sourceSMCtxt values", sourceSMCtxt.UeLocation)
		fmt.Println("destSMCtxt values", destSMCtxt.UeLocation)
	}
	res1 = reflect.DeepEqual(sourceSMCtxt.BPManager, destSMCtxt.BPManager)
	if res1 {
		fmt.Println("two BPManager fields are equal")
	} else {
		fmt.Println("two BPManager fields are not equal")
		fmt.Println("sourceSMCtxt values", sourceSMCtxt.BPManager)
		fmt.Println("destSMCtxt values", destSMCtxt.BPManager)
	}
	res1 = reflect.DeepEqual(sourceSMCtxt.ActiveTxn, destSMCtxt.ActiveTxn)
	if res1 {
		fmt.Println("two ActiveTxn fields are equal")
	} else {
		fmt.Println("two ActiveTxn fields are not equal")
		fmt.Println("sourceSMCtxt values", sourceSMCtxt.ActiveTxn)
		fmt.Println("destSMCtxt values", destSMCtxt.ActiveTxn)
	}
	// res1 = reflect.DeepEqual(sourceSMCtxt.Tunnel.DataPathPool[1].FirstDPNode.DownLinkTunnel.DestEndPoint, destSMCtxt.Tunnel.DataPathPool[1].FirstDPNode.DownLinkTunnel.DestEndPoint)
	// if res1 {
	// 	fmt.Println("two DestEndPoint fields are equal")
	// } else {
	// 	fmt.Println("two FirstDPNode.DownLinkTunnel.DestEndPoint fields are not equal")
	// 	fmt.Println("sourceSMCtxt values", sourceSMCtxt.Tunnel.DataPathPool[1].FirstDPNode.DownLinkTunnel.DestEndPoint)
	// 	fmt.Println("destSMCtxt values", destSMCtxt.Tunnel.DataPathPool[1].FirstDPNode.DownLinkTunnel.DestEndPoint)
	// 	fmt.Println("sourceSMCtxt values str %v", sourceSMCtxt.Tunnel.DataPathPool[1].FirstDPNode.DownLinkTunnel.DestEndPoint)
	// 	fmt.Println("destSMCtxt values str %v", destSMCtxt.Tunnel.DataPathPool[1].FirstDPNode.DownLinkTunnel.DestEndPoint)
	// 	fmt.Println("sourceSMCtxt values ul %v", sourceSMCtxt.Tunnel.DataPathPool[1].FirstDPNode.DownLinkTunnel.DestEndPoint.UpLinkTunnel)
	// 	fmt.Println("destSMCtxt values ul %v", destSMCtxt.Tunnel.DataPathPool[1].FirstDPNode.DownLinkTunnel.DestEndPoint.UpLinkTunnel)
	// 	fmt.Println("sourceSMCtxt values dl %v", sourceSMCtxt.Tunnel.DataPathPool[1].FirstDPNode.DownLinkTunnel.DestEndPoint.DownLinkTunnel)
	// 	fmt.Println("destSMCtxt values dl %v", destSMCtxt.Tunnel.DataPathPool[1].FirstDPNode.DownLinkTunnel.DestEndPoint.DownLinkTunnel)
	// }

}
