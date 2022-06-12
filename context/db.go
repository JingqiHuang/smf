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
	// "github.com/sirupsen/logrus"
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
	//  ---- why bug
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
				fmt.Println("db - in MarshalJSON dataPathNode from dataPath.FirstDPNode %v", dataPathNode)
				fmt.Println("db - in MarshalJSON dataPath %v", dataPath)
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
	// fmt.Println("db - after Unmarshal in customized unMarshall")
	// fmt.Println("db - after Unmarshal in customized unMarshall smContext = %v", smContext)
	// fmt.Println("db - after Unmarshal in customized unMarshall aux.ActiveTxn = %v",aux.ActiveTxn)

	// fmt.Println("db - after Unmarshal in customized unMarshall start recover ActiveTxn")
	smContext.ActiveTxn = &transaction.Transaction{}
	// recover ActiveTxn
	smContext.ActiveTxn.SetStartTime(aux.ActiveTxn.StartTime())
	smContext.ActiveTxn.SetEndTime(aux.ActiveTxn.EndTime())
	smContext.ActiveTxn.TxnId = aux.ActiveTxn.TxnId
	smContext.ActiveTxn.Priority = aux.ActiveTxn.Priority
	// smContext.ActiveTxn.Req = aux.ActiveTxn.Req
	// smContext.ActiveTxn.Rsp = aux.ActiveTxn.Rsp
	// smContext.ActiveTxn.Ctxt = aux.ActiveTxn.Ctxt
	smContext.ActiveTxn.MsgType = svcmsgtypes.SmfMsgTypeType(aux.ActiveTxn.MsgType)
	smContext.ActiveTxn.CtxtKey = aux.ActiveTxn.CtxtKey
	smContext.ActiveTxn.Err = aux.ActiveTxn.Err
	fmt.Println("db - after Unmarshal in customized unMarshall finished recovered ActiveTxn %v", smContext.ActiveTxn)
	// subField := logrus.Fields{"txnid": smContext.ActiveTxn.TxnId, "txntype": string(smContext.ActiveTxn.MsgType), "ctxtkey": smContext.ActiveTxn.CtxtKey}
	// smContext.ActiveTxn.TxnFsmLog = logger.TxnFsmLog.WithFields(subField)

	// smContext.ActiveTxn.Status = aux.ActiveTxn.Status
	// // need to retrieve next txn in db
	// // smContext.ActiveTxn.NextTxnId = &aux.ActiveTxn.NextTxnId
	// smContext.ActiveTxn.TxnFsmLog = aux.ActiveTxn.TxnFsmLog
	
	fmt.Println("db - after Unmarshal in customized unMarshall start recover Tunnel %v", aux.Tunnel)
	var val1 interface{}
	var tmp2 interface{}
	var nilVal *UPTunnelInDB = nil
	smContext.Tunnel = &UPTunnel{}
	if &aux.Tunnel != nilVal {
		smContext.Tunnel.ANInformation = aux.Tunnel.ANInformation
		smContext.Tunnel.PathIDGenerator = idgenerator.NewGenerator(1, 2147483647)
		smContext.Tunnel.DataPathPool = NewDataPathPool()
		fmt.Println("aux.Tunnel.DataPathPool %v", aux.Tunnel.DataPathPool)
		for key, val := range aux.Tunnel.DataPathPool {
			val1 = val
			dataPathInDB := val1.(*DataPathInDB)
			fmt.Println("dataPathInDB - val1 %v", *dataPathInDB)
			fmt.Println("dataPathInDB - val1 dataPathInDB.FirstDPNode %v", *dataPathInDB.FirstDPNode)
			fmt.Println("dataPathInDB - val1 dataPathInDB.FirstDPNode.DataPathNodeUPFNodeID %v", dataPathInDB.FirstDPNode.DataPathNodeUPFNodeID)
			tmp := dataPathInDB.FirstDPNode
			tmp2 = tmp
			dataPathNodeInDBVal := tmp2.(*DataPathNodeInDB)
			dataPathNodeVal := RecoverDataPathNode(dataPathNodeInDBVal)
			fmt.Println("dataPathInDB - dataPathNodeVal after RecoverDataPathNode %v", dataPathNodeVal)
			fmt.Println("dataPathInDB - dataPathNodeVal after RecoverDataPathNode == nil? %v", (dataPathNodeVal==nil))
			fmt.Println("dataPathInDB - dataPathNodeVal after RecoverDataPathNode type", reflect.TypeOf(dataPathNodeVal))
			fmt.Println("dataPathInDB - *dataPathNodeVal after RecoverDataPathNode %v", *dataPathNodeVal)
			fmt.Println("dataPathInDB - *dataPathNodeVal after RecoverDataPathNode type", reflect.TypeOf(*dataPathNodeVal))
			
			newDataPath := NewDataPath()
			// dataPathNodeVal := NewDataPathNode()
			newDataPath.Activated =  dataPathInDB.Activated
			newDataPath.IsDefaultPath = dataPathInDB.IsDefaultPath
			newDataPath.Destination = dataPathInDB.Destination
			newDataPath.HasBranchingPoint = dataPathInDB.HasBranchingPoint
			// fmt.Println("dataPathInDB - newDataPath before assigning FirstDPNode %v", newDataPath)
			newDataPath.FirstDPNode = dataPathNodeVal

			/*
			fmt.Println("\ndataPathInDB - dataPathNodeVal 1 %v", dataPathNodeVal)
			fmt.Println("\ndataPathInDB - dataPathNodeVal type %v", reflect.TypeOf(dataPathNodeVal))
			fmt.Println("\ndataPathInDB - newDataPath.Activated %v", newDataPath.Activated)
			fmt.Println("\ndataPathInDB - newDataPath.IsDefaultPath %v", newDataPath.IsDefaultPath)
			fmt.Println("\ndataPathInDB - newDataPath.Destination %v", newDataPath.Destination)
			fmt.Println("\ndataPathInDB - newDataPath.HasBranchingPoint %v", newDataPath.HasBranchingPoint)
			fmt.Println("\ndataPathInDB - dataPathNodeVal.UPF %v", dataPathNodeVal.UPF)
			fmt.Println("\ndataPathInDB - dataPathNodeVal.UpLinkTunnel %v", dataPathNodeVal.UpLinkTunnel)
			fmt.Println("\ndataPathInDB - dataPathNodeVal.DownLinkTunnel %v", dataPathNodeVal.DownLinkTunnel)
			fmt.Println("\ndataPathInDB - dataPathNodeVal.IsBranchingPoint %v", dataPathNodeVal.IsBranchingPoint)
			fmt.Println("dataPathInDB - newDataPath after assigning FirstDPNode ", newDataPath)
			fmt.Println("dataPathInDB -  newDataPath == nil? %v", (newDataPath==nil))

			tmp_dp := NewDataPath()
			fmt.Println("dataPathInDB - before assigning node tmp_dp ", tmp_dp)
			fmt.Println("dataPathInDB - tmp_dp.FirstDPNode ", tmp_dp.FirstDPNode)

			node := NewDataPathNode()
			
			tmp_dp.FirstDPNode = node
			fmt.Println("dataPathInDB - after assigning node tmp_dp ", tmp_dp)
			fmt.Println("dataPathInDB - tmp_dp.FirstDPNode ", tmp_dp.FirstDPNode)
			
			
			tmp_dp1 := NewDataPath()
			// tmp_dp1.Activated =  dataPathInDB.Activated
			fmt.Println("dataPathInDB - before assigning node tmp_dp1 ", tmp_dp1)
			fmt.Println("dataPathInDB - tmp_dp1.FirstDPNode ", tmp_dp1.FirstDPNode)
			node1 := &DataPathNode{
				UPF: dataPathNodeVal.UPF,
				UpLinkTunnel:   dataPathNodeVal.UpLinkTunnel,
				DownLinkTunnel: dataPathNodeVal.DownLinkTunnel,
				IsBranchingPoint: dataPathNodeVal.IsBranchingPoint,
			}
			
			tmp_dp1.FirstDPNode = node1
			
			fmt.Println("dataPathInDB - newDataPath after assigning node1 ", newDataPath)
			fmt.Println("dataPathInDB - after assigning node tmp_dp1 ", tmp_dp1)
			fmt.Println("dataPathInDB - tmp_dp.FirstDPNode ", tmp_dp1.FirstDPNode)

			tmp_dp2 := new(DataPath)
			fmt.Println("dataPathInDB - before assigning node tmp_dp2 ", tmp_dp2)
			fmt.Println("dataPathInDB - tmp_dp2.FirstDPNode ", tmp_dp2.FirstDPNode)

			node2 := new(DataPathNode)
			
			tmp_dp2.FirstDPNode = node2
			fmt.Println("dataPathInDB - after assigning node tmp_dp2 ", tmp_dp2)
			fmt.Println("dataPathInDB - tmp_dp2.FirstDPNode ", tmp_dp2.FirstDPNode)
			*/
			newDataPath.FirstDPNode = node1
			smContext.Tunnel.DataPathPool[key] = newDataPath
			fmt.Println("dataPathInDB - key %v", key)
			//TB debug!!!
			fmt.Println("dataPathInDB - newDataPath after smContext.Tunnel.DataPathPool[key] %v", newDataPath)
			fmt.Printf("dataPathInDB - newDataPath type: %T\n", newDataPath)
			// fmt.Println("dataPathInDB - &newDataPath %v", &newDataPath)
			// fmt.Printf("dataPathInDB - &newDataPath type: %T\n", &newDataPath)
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
	smContext = &SMContext{}
	filter := bson.M{}
	filter["ref"] = ref

	result := MongoDBLibrary.RestfulAPIGetOne(SmContextDataColl, filter)

	fmt.Println("GetSMContextByRefInDB, smf state json : %v", result)
	// TBA Return SM context, reconstruct SM context from DB json
	err := json.Unmarshal(mapToByte(result), smContext)
	fmt.Println("GetSMContextByRefInDB, after Unmarshal : %v", smContext)
	// smContext.SMLock.Lock()
	// defer smContext.SMLock.Unlock()

	if err != nil {
		logger.CtxLog.Errorf("smContext unmarshall error: %v", err)
		return nil
	}
	// smContext.initLogTags()
	// fmt.Println("GetSMContextByRefInDB, smf state smContext : %v", smContext)
	return smContext
}

func GetSMContextBySEIDInDB(seid uint64){
	filter := bson.M{}
	filter["seid"] = seid
	
	result := MongoDBLibrary.RestfulAPIGetOne(SeidSmContextCol, filter)
	seidStr := strconv.FormatInt(int64(seid), 10)
	ref := result[seidStr].(string)
	fmt.Println("GetSMContextBySEIDInDB, ref string : ", ref)

	GetSMContextByRefInDB(ref)
	// TBA Return SM context, reconstruct SM context from DB json

}

func DeleteContextInDB(smContext *SMContext) {
	fmt.Println("db - delete SMContext In DB w ref!!")
	// smContextBsonA := ToBsonM(smContext)
	filter := bson.M{"ref": smContext.Ref}
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