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

	"github.com/omec-project/smf/logger"
	"github.com/sirupsen/logrus"
	// "github.com/free5gc/openapi/models"

	"github.com/omec-project/idgenerator"
	"net"
	"time"
	"github.com/omec-project/smf/msgtypes/svcmsgtypes"
	"strconv"
	"github.com/omec-project/smf/transaction"
	"reflect"

)

const (
	SmContextDataColl = "smf.data.smContext"
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
	CtxtRef             string
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
	FirstDPNode *DataPathNodeInDB
	// FirstDPNode *DataPathNodeInDB
}

type DataPathNodeInDB  struct {
    // UPF *UPF
	DataPathNodeUPFNodeID NodeIDInDB 

	ULTunnelInfo	*TunnelInfo
	DLTunnelInfo	*TunnelInfo

    IsBranchingPoint bool
}
type TunnelInfo struct{
	DataPathNodeUPFNodeID NodeIDInDB
	TEID uint32
	PDR  map[string]*PDR
}

type NodeIDInDB struct {
	NodeIdType  uint8 // 0x00001111
	NodeIdValue []byte
}

func SetupSmfCollection() {
	fmt.Println("db - SetupSmfCollection!!")
	MongoDBLibrary.SetMongoDB("sdcore", "mongodb://mongodb")
	_, err := MongoDBLibrary.CreateIndex(SmContextDataColl, "supi")
	if err != nil {
		logger.CtxLog.Errorf("Create index failed on Supi field.")
	}
	MongoDBLibrary.SetMongoDB("sdcore", "mongodb://mongodb")
	_, err = MongoDBLibrary.CreateIndex(SmContextDataColl, "ref")
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
	testDB()
	
}


func (smContext *SMContext) MarshalJSON() ([]byte, error) {
	type Alias SMContext
	// UPTunnel: GTPTunnel -> TEID
	fmt.Println("db - in customized MarshalJSON ")
	var msgTypeVal string
	
	var activeTxnVal TransactionInDB
	if smContext.ActiveTxn != nil {
		msgTypeVal = svcmsgtypes.SmfMsgTypeString(smContext.ActiveTxn.MsgType)
		fmt.Println("db - in customized MarshalJSON ", msgTypeVal)
		activeTxnVal.startTime =  smContext.ActiveTxn.StartTime()
		activeTxnVal.endTime = smContext.ActiveTxn.EndTime()
		activeTxnVal.TxnId = smContext.ActiveTxn.TxnId
		activeTxnVal.Priority= smContext.ActiveTxn.Priority
		activeTxnVal.Req = smContext.ActiveTxn.Req
		activeTxnVal.Rsp = smContext.ActiveTxn.Rsp
		activeTxnVal.MsgType= msgTypeVal
		activeTxnVal.CtxtKey= smContext.ActiveTxn.CtxtKey
		activeTxnVal.Err= smContext.ActiveTxn.Err
		//  ---- why bug
		fmt.Println("db - in MarshalJSON check req %v %v ", smContext.ActiveTxn.Req, (smContext.ActiveTxn.Req != nil))
		fmt.Println("db - in MarshalJSON check rsp %v %v ", smContext.ActiveTxn.Rsp, (smContext.ActiveTxn.Rsp != nil))
		fmt.Println("db - in MarshalJSON check CtxtRef ", (smContext.ActiveTxn.Ctxt.(*SMContext).Ref))
		if smContext.ActiveTxn.NextTxn != nil {
			activeTxnVal.NextTxnId = smContext.ActiveTxn.NextTxn.TxnId
		}
		if smContext.ActiveTxn.Ctxt != nil {
			activeTxnVal.CtxtRef = smContext.ActiveTxn.Ctxt.(*SMContext).Ref
		}
		StoreTxnInDB(&activeTxnVal)
	} 

	fmt.Println("db - in MarshalJSON after activeTxnVal")
	
	
	
	dataPathPoolInDBVal:= make(map[int64]*DataPathInDB)

	var dataPathInDBIf interface{}
	var FirstDPNodeIf interface{}

	var upTunnelVal UPTunnelInDB
	if smContext.Tunnel != nil {
		upTunnelVal.ANInformation = smContext.Tunnel.ANInformation

		if smContext.Tunnel.DataPathPool != nil {
			fmt.Println("db - smContext.Tunnel.DataPathPool len ",len(smContext.Tunnel.DataPathPool)) 
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

				fmt.Println("db - in MarshalJSON dataPathNodeInDBVal %v", dataPathNodeInDBVal)
				
				newDataPathInDB := &DataPathInDB{
					Activated: dataPath.Activated,
					IsDefaultPath: dataPath.IsDefaultPath,
					Destination: dataPath.Destination,
					HasBranchingPoint: dataPath.HasBranchingPoint,
					FirstDPNode: dataPathNodeInDBVal,
				}
				
				dataPathPoolInDBVal[key] = newDataPathInDB
			}
			upTunnelVal.DataPathPool = dataPathPoolInDBVal
		}
	}
	
	fmt.Println("db - in MarshalJSON before return")
	
	return json.Marshal(&struct {
		ActiveTxn	TransactionInDB	`json:"activeTxn"`
		Tunnel	UPTunnelInDB	`json:"tunnel"`
		*Alias
	}{
		ActiveTxn: activeTxnVal,
		Tunnel: upTunnelVal,
		Alias:       (*Alias)(smContext),
	})
}

func (smContext *SMContext) UnmarshalJSON(data []byte) error {
	fmt.Println("db - in UnmarshalJSON")
	type Alias SMContext
	aux := &struct {
		ActiveTxn	TransactionInDB	`json:"activeTxn"`
		Tunnel	UPTunnelInDB	`json:"tunnel"`
		*Alias
	}{
		Alias: (*Alias)(smContext),
	}
	// fmt.Println("db - before Unmarshal in customized unMarshall")
	if err := json.Unmarshal(data, &aux); err != nil {
		fmt.Println("Err in customized unMarshall!!")
		return err
	}
	
	smContext.ActiveTxn = &transaction.Transaction{}
	// recover ActiveTxn
	smContext.ActiveTxn.SetStartTime(aux.ActiveTxn.StartTime())
	smContext.ActiveTxn.SetEndTime(aux.ActiveTxn.EndTime())
	smContext.ActiveTxn.TxnId = aux.ActiveTxn.TxnId
	smContext.ActiveTxn.Priority = aux.ActiveTxn.Priority
	smContext.ActiveTxn.Req = aux.ActiveTxn.Req
	smContext.ActiveTxn.Rsp = aux.ActiveTxn.Rsp
	smContext.ActiveTxn.Ctxt = GetSMContext(aux.ActiveTxn.CtxtKey)
	smContext.ActiveTxn.MsgType = svcmsgtypes.SmfMsgTypeType(aux.ActiveTxn.MsgType)
	smContext.ActiveTxn.CtxtKey = aux.ActiveTxn.CtxtKey
	smContext.ActiveTxn.Err = aux.ActiveTxn.Err
	fmt.Println("db - after Unmarshal in customized unMarshall finished recovered ActiveTxn %v", smContext.ActiveTxn)
	subField := logrus.Fields{"txnid": smContext.ActiveTxn.TxnId, "txntype": string(smContext.ActiveTxn.MsgType), "ctxtkey": smContext.ActiveTxn.CtxtKey}
	smContext.ActiveTxn.TxnFsmLog = logger.TxnFsmLog.WithFields(subField)

	smContext.ActiveTxn.Status = make(chan bool)
	// TODO: need to retrieve next txn in db
	// smContext.ActiveTxn.NextTxnId = &aux.ActiveTxn.NextTxnId
	
	fmt.Println("db - after Unmarshal in customized unMarshall start recover Tunnel %v", aux.Tunnel)
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
			fmt.Println("dataPathInDB - dataPathInDBIf %v", *dataPathInDB)
			fmt.Println("dataPathInDB - dataPathInDBIf dataPathInDB.FirstDPNode %v", *dataPathInDB.FirstDPNode)
			fmt.Println("dataPathInDB - dataPathInDBIf dataPathInDB.FirstDPNode.DataPathNodeUPFNodeID %v", dataPathInDB.FirstDPNode.DataPathNodeUPFNodeID)
			firstDPNode := dataPathInDB.FirstDPNode
			FirstDPNodeIf = firstDPNode
			dataPathNodeInDBVal := FirstDPNodeIf.(*DataPathNodeInDB)
			dataPathNodeVal := RecoverDataPathNode(dataPathNodeInDBVal)
			fmt.Println("dataPathInDB - dataPathNodeVal after RecoverDataPathNode %v", dataPathNodeVal)
			
			newDataPath := NewDataPath()
			
			newDataPath.Activated =  dataPathInDB.Activated
			newDataPath.IsDefaultPath = dataPathInDB.IsDefaultPath
			newDataPath.Destination = dataPathInDB.Destination
			newDataPath.HasBranchingPoint = dataPathInDB.HasBranchingPoint
			
			newDataPath.FirstDPNode = dataPathNodeVal

			smContext.Tunnel.DataPathPool[key] = newDataPath
			
			fmt.Println("dataPathInDB - newDataPath after smContext.Tunnel.DataPathPool[key] %v", newDataPath)
		}
		
		fmt.Println("smContext.Tunnel.DataPathPool %v", smContext.Tunnel.DataPathPool)
	}	
	fmt.Println("db - after Unmarshal in customized unMarshall after recover Tunnel %v", smContext.Tunnel)
	fmt.Println("db - in UnMarshalJSON before return")
	
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
	fmt.Println("db - in ToBsonM before Unmarshal")
	err = json.Unmarshal(tmp, &ret)
	if err != nil {
		logger.CtxLog.Errorf("SMContext unmarshall error: %v", err)
	}

	return
}

func StoreSmContextInDB(smContext *SMContext) {
	fmt.Println("db - Store SMContext In DB w ref!!")
	smContextBsonA := ToBsonM(smContext)
	fmt.Println("db - in StoreSmContextInDB after ToBsonM")
	filter := bson.M{"ref": smContext.Ref}
	logger.CtxLog.Infof("filter: ", filter)

	MongoDBLibrary.RestfulAPIPost(SmContextDataColl, filter, smContextBsonA)
}

type SeidSmContextRef struct {
	Ref string `json:"ref" yaml:"ref" bson:"ref"`
	Seid uint64 `json:"seid" yaml:"seid" bson:"seid"`
}

func StoreSeidContextInDB(seid uint64, smContext *SMContext) {
	fmt.Println("db - StoreSeidContextInDB In DB!! smcontext ", smContext)
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

func GetSMContextByRefInDB(ref string) (smContext *SMContext){
	smContext = &SMContext{}
	filter := bson.M{}
	filter["ref"] = ref

	result := MongoDBLibrary.RestfulAPIGetOne(SmContextDataColl, filter)

	fmt.Println("GetSMContextByRefInDB, smf state json : %v", result)
	// TBA Return SM context, reconstruct SM context from DB json
	err := json.Unmarshal(mapToByte(result), smContext)
	fmt.Println("GetSMContextByRefInDB, after Unmarshal : %v", smContext)
	smContext.SMLock.Lock()
	defer smContext.SMLock.Unlock()

	if err != nil {
		logger.CtxLog.Errorf("smContext unmarshall error: %v", err)
		return nil
	}
	
	return smContext
}

func GetSMContextBySEIDInDB(seid uint64) (smContext *SMContext){
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
	seidStr := strconv.FormatInt(int64(seid), 10)
	ref := result[seidStr].(string)
	fmt.Println("GetSMContextBySEIDInDB, ref string : ", ref)

	MongoDBLibrary.RestfulAPIDeleteOne(SmContextDataColl, filter)
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
	fmt.Println("db - in mapToByte!! data: %v", data)
	ret, _ = json.Marshal(data)
	// fmt.Println("db - in mapToByte!! ret: %v", ret)
	return 
}

func testDB(){
	filter := bson.M{}
	filter["txnID"] = -1

	result := MongoDBLibrary.RestfulAPIGetOne(TransactionDataCol, filter)

	print("in db - testdb() result ", result)
	print("in db - testdb() result type ", reflect.TypeOf(result))
}