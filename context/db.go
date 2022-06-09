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

	var msgTypeVal string
	msgTypeVal = svcmsgtypes.SmfMsgTypeString(smContext.ActiveTxn.MsgType)
	fmt.Println("db - in MarshalJSON ", msgTypeVal)
	var activeTxnVal TransactionInDB
	if smContext.ActiveTxn != nil {
		activeTxnVal.startTime =  smContext.ActiveTxn.StartTime()
		activeTxnVal.endTime = smContext.ActiveTxn.EndTime()
		activeTxnVal.TxnId = smContext.ActiveTxn.TxnId
		activeTxnVal.Priority= smContext.ActiveTxn.Priority
			// Req: smContext.ActiveTxn.Req,
			// Rsp: smContext.ActiveTxn.Rsp,
			// Ctxt: smContext.ActiveTxn.Ctxt,
		activeTxnVal.MsgType= msgTypeVal
		activeTxnVal.CtxtKey= smContext.ActiveTxn.CtxtKey
		activeTxnVal.Err= smContext.ActiveTxn.Err
			// Status: smContext.ActiveTxn.Status,
			// NextTxnId: smContext.ActiveTxn.NextTxn.TxnId,
			// TxnFsmLog: smContext.ActiveTxn.TxnFsmLog,
	} 

	fmt.Println("db - in MarshalJSON after activeTxnVal")
	// if smContext.ActiveTxn.Req != nil {
	// 	activeTxnVal.Req = smContext.ActiveTxn.Req
	// }
	// if smContext.ActiveTxn.Rsp != nil {
	// 	activeTxnVal.Rsp = smContext.ActiveTxn.Rsp
	// }
	// if smContext.ActiveTxn.Ctxt != nil {
	// 	activeTxnVal.Ctxt = smContext.ActiveTxn.Ctxt
	// }
	
	
	dataPathPoolInDBVal:= make(map[int64]*DataPathInDB)

	var val1 interface{}
	var tmp2 interface{}

	var upTunnelVal UPTunnelInDB
	if smContext.Tunnel != nil {
		upTunnelVal.ANInformation = smContext.Tunnel.ANInformation

		if smContext.Tunnel.DataPathPool != nil {
			fmt.Println("db - smContext.Tunnel.DataPathPool len ",len(smContext.Tunnel.DataPathPool)) 
			for key, val := range smContext.Tunnel.DataPathPool {

				val1 = val
				dataPath := val1.(*DataPath)
				
				tmp := dataPath.FirstDPNode
				tmp2 = tmp

				dataPathNode := tmp2.(*DataPathNode)

				// convert dataPathNode to DataPathNodeInDB 
				dataPathNodeInDBVal := DFSStoreDataPathNode(dataPathNode)
				// dataPathNodeInDBVal := &DataPathNodeInDB{}

				fmt.Println("db - in MarshalJSON dataPathNodeInDBVal = %v", dataPathNodeInDBVal)
				
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
	
	/*
	var val1 interface{}
	// var tmp2 interface{}
	
	if &aux.Tunnel != nil {

		smContext.Tunnel.ANInformation = aux.Tunnel.ANInformation
		smContext.Tunnel.PathIDGenerator = idgenerator.NewGenerator(1, 2147483647)
		smContext.Tunnel.DataPathPool = make(DataPathPool)

		for key, val := range aux.Tunnel.DataPathPool {
			val1 = val
			dataPathInDB := val1.(*DataPathInDB)

			// tmp := dataPathInDB.FirstDPNode
			// tmp2 = tmp
			// dataPathNodeInDBVal := tmp2.(*DataPathNodeInDB)
			// dataPathNodeVal := DFSRecoverDataPathNode(dataPathNodeInDBVal)
			dataPathNodeVal := &DataPathNode{}

			newDataPath := &DataPath{
				Activated: dataPathInDB.Activated,
				IsDefaultPath: dataPathInDB.IsDefaultPath,
				Destination: dataPathInDB.Destination,
				HasBranchingPoint: dataPathInDB.HasBranchingPoint,
				FirstDPNode: dataPathNodeVal,				
			}
			smContext.Tunnel.DataPathPool[key] = newDataPath
		}
		
		fmt.Println("smContext.Tunnel.DataPathPool %v", smContext.Tunnel.DataPathPool)
	}
	*/
	
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
	// fmt.Println("db - finished Store SMContext In DB!!---checking")

	// filter = bson.M{}
	// filter["ref"] = smContext.Ref

	// result := MongoDBLibrary.RestfulAPIGetOne(SmContextDataColl, filter)

	// fmt.Println("StoreSmContextInDB, smf state json : ", result)
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

func GetSMContextByRefInDB(ref string) (smContext *SMContext){
	filter := bson.M{}
	filter["ref"] = ref

	result := MongoDBLibrary.RestfulAPIGetOne(SmContextDataColl, filter)

	fmt.Println("GetSMContextByRefInDB, smf state json : ", result)
	// TBA Return SM context, reconstruct SM context from DB json
	err := json.Unmarshal(mapToByte(result), smContext)
	
	smContext.SMLock.Lock()
	defer smContext.SMLock.Unlock()

	if err != nil {
		logger.CtxLog.Errorf("smContext unmarshall error: %v", err)
		return nil
	}
	smContext.initLogTags()
	return smContext
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