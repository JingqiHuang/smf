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

	"github.com/omec-project/smf/logger"

	// "github.com/free5gc/openapi/models"

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

func SetupSmfCollection() {
	fmt.Println("db - SetupSmfCollection")
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

}

// print out sm context
func (smContext *SMContext) String() string {
	return fmt.Sprintf("smContext content: Ref:[%v],\nSupi: [%v],\nPei:[%v],\nGpsi:[%v],\nPDUSessionID:[%v],\nDnn:[%v],Snssai: [%v],\nHplmnSnssai: [%v],\nServingNetwork: [%v],\nServingNfId: [%v],\nUpCnxState: [%v],\nAnType: [%v],\nRatType: [%v],\nPDUAddress: [%v],\nSelectedPDUSessionType: [%v],\nSmStatusNotifyUri: [%v],\nSelectedPCFProfile: [%v],\nSMContextState: [%v],\nTunnel: [%v],\nPFCPContext: [%v],\nIdentifier: [%v],\nDNNInfo: [%v],\nSmPolicyData: [%v],\nEstAcceptCause5gSMValue: [%v]\n", smContext.Ref, smContext.Supi, smContext.Pei, smContext.Gpsi, smContext.PDUSessionID, smContext.Dnn, smContext.Snssai, smContext.HplmnSnssai, smContext.ServingNetwork, smContext.ServingNfId, smContext.UpCnxState, smContext.AnType, smContext.RatType, smContext.PDUAddress, smContext.SelectedPDUSessionType, smContext.SmStatusNotifyUri, smContext.SelectedPCFProfile, smContext.SMContextState, smContext.Tunnel, smContext.PFCPContext, smContext.Identifier, smContext.DNNInfo, smContext.SmPolicyData, smContext.EstAcceptCause5gSMValue)
}

// customized mashaller for sm context
func (smContext *SMContext) MarshalJSON() ([]byte, error) {
	type Alias SMContext

	dataPathPoolInDBVal := make(map[int64]*DataPathInDB)

	var dataPathInDBIf interface{}
	var FirstDPNodeIf interface{}

	var upTunnelVal UPTunnelInDB
	if smContext.Tunnel != nil {
		upTunnelVal.ANInformation = smContext.Tunnel.ANInformation

		if smContext.Tunnel.DataPathPool != nil {
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
		pfcpSessionContextInDB.LocalSEID = strconv.FormatUint(pfcpCtx.LocalSEID, 10)
		pfcpSessionContextInDB.RemoteSEID = strconv.FormatUint(pfcpCtx.RemoteSEID, 10)
		PFCPContextVal[key] = pfcpSessionContextInDB
	}

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

// customized unmashaller for sm context
func (smContext *SMContext) UnmarshalJSON(data []byte) error {
	fmt.Println("db - in UnmarshalJSON")
	type Alias SMContext
	aux := &struct {
		Tunnel         UPTunnelInDB    `json:"tunnel"`
		PFCPContextVal PFCPContextInDB `json:"pfcpContext"`
		*Alias
	}{
		Alias: (*Alias)(smContext),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		fmt.Println("Err in customized unMarshall!!")
		return err
	}

	// recover smContext.PFCPContext
	smContext.PFCPContext = make(map[string]*PFCPSessionContext)
	for key, pfcpCtxInDB := range aux.PFCPContextVal {
		smContext.PFCPContext[key] = &PFCPSessionContext{}
		smContext.PFCPContext[key].NodeID = pfcpCtxInDB.NodeID
		smContext.PFCPContext[key].PDRs = pfcpCtxInDB.PDRs
		smContext.PFCPContext[key].LocalSEID, _ = strconv.ParseUint(string(pfcpCtxInDB.LocalSEID), 10, 64)
		smContext.PFCPContext[key].RemoteSEID, _ = strconv.ParseUint(string(pfcpCtxInDB.RemoteSEID), 10, 64)
	}

	var dataPathInDBIf interface{}
	var FirstDPNodeIf interface{}
	var nilVal *UPTunnelInDB = nil
	smContext.Tunnel = &UPTunnel{}
	if &aux.Tunnel != nilVal {
		smContext.Tunnel.ANInformation = aux.Tunnel.ANInformation
		smContext.Tunnel.PathIDGenerator = idgenerator.NewGenerator(1, 2147483647)
		smContext.Tunnel.DataPathPool = NewDataPathPool()
		for key, val := range aux.Tunnel.DataPathPool {
			dataPathInDBIf = val
			dataPathInDB := dataPathInDBIf.(*DataPathInDB)

			firstDPNode := dataPathInDB.FirstDPNode
			FirstDPNodeIf = firstDPNode
			dataPathNodeInDBVal := FirstDPNodeIf.(*DataPathNodeInDB)
			dataPathNodeVal := RecoverDataPathNode(dataPathNodeInDBVal)

			newDataPath := NewDataPath()

			newDataPath.Activated = dataPathInDB.Activated
			newDataPath.IsDefaultPath = dataPathInDB.IsDefaultPath
			newDataPath.Destination = dataPathInDB.Destination
			newDataPath.HasBranchingPoint = dataPathInDB.HasBranchingPoint

			newDataPath.FirstDPNode = dataPathNodeVal

			smContext.Tunnel.DataPathPool[key] = newDataPath

		}

	}
	// recover logs
	smContext.initLogTags()
	// recover SBIPFCPCommunicationChan
	smContext.SBIPFCPCommunicationChan = make(chan PFCPSessionResponseStatus, 1)

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

// Store SmContext In DB
func StoreSmContextInDB(smContext *SMContext) {
	fmt.Println("db - Store SMContext In DB w ref")
	smContext.SMLock.Lock()
	defer smContext.SMLock.Unlock()
	smContextBsonA := ToBsonM(smContext)
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

// Store Context by SEID In DB
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
}

// store mapping ref to seid in db
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
}

func GetSeidByRefInDB(ref string) (seid uint64) {
	filter := bson.M{}
	filter["ref"] = ref

	result := MongoDBLibrary.RestfulAPIGetOne(RefSeidCol, filter)
	seidStr := result["seid"].(string)
	seid, _ = strconv.ParseUint(seidStr, 10, 64)
	return
}

// GetSMContext By Ref from DB
func GetSMContextByRefInDB(ref string) (smContext *SMContext) {
	fmt.Println("db - in GetSMContextByRefInDB")
	smContext = &SMContext{}
	filter := bson.M{}
	filter["ref"] = ref

	result := MongoDBLibrary.RestfulAPIGetOne(SmContextDataColl, filter)

	err := json.Unmarshal(mapToByte(result), smContext)
	smContext.SMLock.Lock()
	defer smContext.SMLock.Unlock()

	if err != nil {
		logger.CtxLog.Errorf("smContext unmarshall error: %v", err)
		return nil
	}

	return smContext
}

// GetSMContext By SEID from DB
func GetSMContextBySEIDInDB(seidUint uint64) (smContext *SMContext) {
	seid := SeidConv(seidUint)
	filter := bson.M{}
	filter["seid"] = seid

	result := MongoDBLibrary.RestfulAPIGetOne(SeidSmContextCol, filter)
	ref := result["ref"].(string)
	fmt.Println("StoreSeidContextInDB, result string : ", ref)

	return GetSMContext(ref)
}

// Delete SMContext By SEID from DB
func DeleteContextInDBBySEID(seidUint uint64) {

	seid := SeidConv(seidUint)
	fmt.Println("db - delete SMContext In DB by seid")
	// smContextBsonA := ToBsonM(smContext)
	filter := bson.M{"seid": seid}
	logger.CtxLog.Infof("filter : ", filter)

	result := MongoDBLibrary.RestfulAPIGetOne(SeidSmContextCol, filter)
	ref := result["ref"].(string)

	MongoDBLibrary.RestfulAPIDeleteOne(SeidSmContextCol, filter)
	DeleteContextInDBByRef(ref)

}

// Delete SMContext By ref from DB
func DeleteContextInDBByRef(ref string) {
	fmt.Println("db - delete SMContext In DB w ref")
	// smContextBsonA := ToBsonM(smContext)
	filter := bson.M{"ref": ref}
	logger.CtxLog.Infof("filter : ", filter)
	MongoDBLibrary.RestfulAPIDeleteOne(SmContextDataColl, filter)
}

// Delete SMContext in smContextPool and seidSMContextMap, for test
func ClearSMContextInMem(ref string) {
	smContextPool.Delete(ref)
	seid := GetSeidByRefInDB(ref)
	seidSMContextMap.Delete(seid)
}

func mapToByte(data map[string]interface{}) (ret []byte) {
	ret, _ = json.Marshal(data)
	return
}

func ShowSmContextPool() {
	smContextPool.Range(func(k, v interface{}) bool {
		fmt.Println("db - iterate:", k, v)
		return true
	})

}

func GetSmContextPool() sync.Map {
	return smContextPool
}

func StoreSmContextPool(smContext *SMContext) {
	smContextPool.Store(smContext.Ref, smContext)
}
