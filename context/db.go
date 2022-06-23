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
	"sync"

	"github.com/badhrinathpa/MongoDBLibrary"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/omec-project/pfcp/pfcpType"
	"github.com/omec-project/smf/logger"

	// "github.com/free5gc/openapi/models"

	"net"
	"strconv"

	"github.com/omec-project/idgenerator"
)

const (
	SmContextDataColl = "smf.data.smContext"
	// TransactionDataCol = "smf.data.transaction"
	SeidSmContextCol = "smf.data.seidSmContext"
	NodeInDBCol      = "smf.data.nodeInDB"
	// SMPolicyClientCol  = "smf.data.smPolicyClient"
	RefSeidCol = "smf.data.refToSeid"
)

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
	MongoDBLibrary.SetMongoDB("sdcore", "mongodb://mongodb")
	_, err := MongoDBLibrary.CreateIndex(SmContextDataColl, "ref")
	if err != nil {
		logger.CtxLog.Errorf("Create index failed on ref field.")
	}
	MongoDBLibrary.SetMongoDB("sdcore", "mongodb://mongodb")
	_, err = MongoDBLibrary.CreateIndex(SeidSmContextCol, "seid")
	if err != nil {
		logger.CtxLog.Errorf("Create index failed on TxnId field.")
	}
	// testDB()

}

type PFCPContextInDB map[string]PFCPSessionContextInDB

func (smContext *SMContext) String() string {

	return fmt.Sprintf("smContext content: Ref:[%v],\nSupi: [%v],\nPei:[%v],\nGpsi:[%v],\nPDUSessionID:[%v],\nDnn:[%v],Snssai: [%v],\nHplmnSnssai: [%v],\nServingNetwork: [%v],\nServingNfId: [%v],\nUpCnxState: [%v],\nAnType: [%v],\nRatType: [%v],\nPDUAddress: [%v],\nSelectedPDUSessionType: [%v],\nSmStatusNotifyUri: [%v],\nSelectedPCFProfile: [%v],\nSMContextState: [%v],\nTunnel: [%v],\nPFCPContext: [%v],\nIdentifier: [%v],\nDNNInfo: [%v],\nSmPolicyData: [%v],\nEstAcceptCause5gSMValue: [%v]\n", smContext.Ref, smContext.Supi, smContext.Pei, smContext.Gpsi, smContext.PDUSessionID, smContext.Dnn, smContext.Snssai, smContext.HplmnSnssai, smContext.ServingNetwork, smContext.ServingNfId, smContext.UpCnxState, smContext.AnType, smContext.RatType, smContext.PDUAddress, smContext.SelectedPDUSessionType, smContext.SmStatusNotifyUri, smContext.SelectedPCFProfile, smContext.SMContextState, smContext.Tunnel, smContext.PFCPContext, smContext.Identifier, smContext.DNNInfo, smContext.SmPolicyData, smContext.EstAcceptCause5gSMValue)
}

func (smContext *SMContext) MarshalJSON() ([]byte, error) {
	type Alias SMContext

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

				dataPathNodeInDBVal := StoreDataPathNode(dataPathNode)
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
		// fmt.Println("in customized marshalling pfcpCtx.PDRs ", pfcpCtx.PDRs)
		// fmt.Println("in customized marshalling pfcpSessionContextInDB.PDRs ", pfcpSessionContextInDB.PDRs)
		pfcpSessionContextInDB.LocalSEID = strconv.FormatUint(pfcpCtx.LocalSEID, 10)
		pfcpSessionContextInDB.RemoteSEID = strconv.FormatUint(pfcpCtx.RemoteSEID, 10)
		PFCPContextVal[key] = pfcpSessionContextInDB
	}
	// fmt.Println("db - in customized marshalling PFCPContextVal ", PFCPContextVal)

	return json.Marshal(&struct {
		Tunnel      UPTunnelInDB    `json:"tunnel"`
		PFCPContext PFCPContextInDB `json:"pfcpContext"`
		*Alias
	}{
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
	// fmt.Println("in customized unmarshalling aux.PFCPContextVal ", aux.PFCPContextVal)
	for key, pfcpCtxInDB := range aux.PFCPContextVal {
		smContext.PFCPContext[key] = &PFCPSessionContext{}
		smContext.PFCPContext[key].NodeID = pfcpCtxInDB.NodeID
		smContext.PFCPContext[key].PDRs = pfcpCtxInDB.PDRs
		// fmt.Println("in customized unmarshalling key ", key)
		// fmt.Println("in customized unmarshalling pfcpCtxInDB.PDRs ", pfcpCtxInDB.PDRs)
		// fmt.Println("in customized unmarshalling smContext.PFCPContext[key].PDRs ", smContext.PFCPContext[key].PDRs)
		smContext.PFCPContext[key].LocalSEID, _ = strconv.ParseUint(string(pfcpCtxInDB.LocalSEID), 10, 64)
		smContext.PFCPContext[key].RemoteSEID, _ = strconv.ParseUint(string(pfcpCtxInDB.RemoteSEID), 10, 64)
	}

	// fmt.Println("db - after Unmarshal in customized unMarshall start recover Tunnel %v", aux.Tunnel)
	var dataPathInDBIf interface{}
	var FirstDPNodeIf interface{}
	var nilVal *UPTunnelInDB = nil
	smContext.Tunnel = &UPTunnel{}
	if &aux.Tunnel != nilVal {
		smContext.Tunnel.ANInformation = aux.Tunnel.ANInformation
		smContext.Tunnel.PathIDGenerator = idgenerator.NewGenerator(1, 2147483647)
		smContext.Tunnel.DataPathPool = NewDataPathPool()
		// fmt.Println("aux.Tunnel.DataPathPool %v", aux.Tunnel.DataPathPool)
		for key, val := range aux.Tunnel.DataPathPool {
			dataPathInDBIf = val
			dataPathInDB := dataPathInDBIf.(*DataPathInDB)

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

		}

		// fmt.Println("smContext.Tunnel.DataPathPool %v", smContext.Tunnel.DataPathPool)
	}
	// recover logs
	smContext.initLogTags()
	// recover SBIPFCPCommunicationChan
	smContext.SBIPFCPCommunicationChan = make(chan PFCPSessionResponseStatus, 1)

	fmt.Printf("db - after Unmarshal in customized unMarshall after recover smContext %s\n", smContext)

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
	err = json.Unmarshal(tmp, &ret)
	if err != nil {
		logger.CtxLog.Errorf("SMContext unmarshall error: %v", err)
	}

	return
}

func StoreSmContextInDB(smContext *SMContext) {
	fmt.Println("db - Store SMContext In DB w ref!!")
	smContext.SMLock.Lock()
	defer smContext.SMLock.Unlock()
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
	Seid string `json:"seid" yaml:"seid" bson:"seid"`
}

func SeidConv(seid uint64) (seidStr string) {
	seidStr = strconv.FormatUint(seid, 10)
	return seidStr
}

func StoreSeidContextInDB(seidUint uint64, smContext *SMContext) {
	seid := SeidConv(seidUint)
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

func StoreRefToSeidInDB(seidUint uint64, smContext *SMContext) {
	seid := SeidConv(seidUint)
	item := SeidSmContextRef{
		Ref:  smContext.Ref,
		Seid: seid,
	}
	itemBsonA := ToBsonMSeidRef(item)
	filter := bson.M{"ref": smContext.Ref}
	logger.CtxLog.Infof("filter : ", filter)

	MongoDBLibrary.RestfulAPIPost(RefSeidCol, filter, itemBsonA)
	fmt.Println("db - finished StoreSeidContextInDB In DB!!")
}

func GetSeidByRefInDB(ref string) (seid uint64) {
	filter := bson.M{}
	filter["ref"] = ref

	result := MongoDBLibrary.RestfulAPIGetOne(RefSeidCol, filter)
	seidStr := result["seid"].(string)
	fmt.Println("GetSeidByRefInDB, result string : ", seidStr)
	seid, _ = strconv.ParseUint(seidStr, 10, 64)
	return
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

func GetSMContextBySEIDInDB(seidUint uint64) (smContext *SMContext) {
	seid := SeidConv(seidUint)
	filter := bson.M{}
	filter["seid"] = seid

	result := MongoDBLibrary.RestfulAPIGetOne(SeidSmContextCol, filter)
	ref := result["ref"].(string)
	fmt.Println("StoreSeidContextInDB, result string : ", ref)

	return GetSMContext(ref)
}

func DeleteContextInDBBySEID(seidUint uint64) {

	seid := SeidConv(seidUint)
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
	fmt.Println("db - finished  DeleteContextInDBByRef In DB!!")
}

func ClearSMContextInMem(ref string) {
	fmt.Println("db - in ClearSMContextInMem")
	fmt.Println("db - in ClearSMContextInMem Ref ", ref)
	smContextPool.Delete(ref)
	seid := GetSeidByRefInDB(ref)
	fmt.Println("db - in ClearSMContextInMem seid ", seid)
	seidSMContextMap.Delete(seid)
}

func mapToByte(data map[string]interface{}) (ret []byte) {
	// fmt.Println("db - in mapToByte!! data: %v", data)
	ret, _ = json.Marshal(data)
	// fmt.Println("db - in mapToByte!! ret: %v", ret)
	return
}

func ShowSmContextPool() {
	fmt.Println("db - in ShowSmContextPool()")
	smContextPool.Range(func(k, v interface{}) bool {
		fmt.Println("db - iterate:", k, v)
		return true
	})

}

func GetSmContextPool() sync.Map {
	return smContextPool
}

func StoreSmContextPool(smContext *SMContext) {
	fmt.Println("db - in StoreSmContextPool()")
	smContextPool.Store(smContext.Ref, smContext)
}
