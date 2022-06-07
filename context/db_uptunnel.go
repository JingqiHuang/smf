package context

import (
	"encoding/json"
	"fmt"
	"github.com/badhrinathpa/MongoDBLibrary"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/omec-project/smf/logger"

	// "github.com/omec-project/idgenerator"
	// "net"
	// "time"
	// "github.com/omec-project/smf/msgtypes/svcmsgtypes"
	// "strconv"
	// "github.com/sirupsen/logrus"
	"github.com/omec-project/pfcp/pfcpType"
)

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


func GetNodeIDInDB(nodeID pfcpType.NodeID) NodeIDInDB {
	nodeIDInDB := NodeIDInDB{
		NodeIdType: nodeID.NodeIdType,
		NodeIdValue: nodeID.NodeIdValue,
	}
	return nodeIDInDB
}

func GetNodeID(nodeIDInDB NodeIDInDB) pfcpType.NodeID{
	nodeID := pfcpType.NodeID{
		NodeIdType: nodeIDInDB.NodeIdType,
		NodeIdValue: nodeIDInDB.NodeIdValue,
	}
	return nodeID
}

func RecoverUpLinkTunnel(tunnelInfo *TunnelInfo) *GTPTunnel {
	tunnel := &GTPTunnel{
		TEID: tunnelInfo.TEID,
		PDR: tunnelInfo.PDR,
	}
	var nilVal *NodeIDInDB = nil
	if &tunnelInfo.DataPathNodeUPFNodeID != nilVal {
		dEP := RecoverDataPathNode(tunnelInfo.DataPathNodeUPFNodeID)
		tunnel.DestEndPoint = dEP
	}
	return tunnel		

}

func RecoverDownLinkTunnel(tunnelInfo *TunnelInfo) *GTPTunnel {

	sEP := RecoverDataPathNode(tunnelInfo.DataPathNodeUPFNodeID)
	tunnel := &GTPTunnel{
		SrcEndPoint: sEP,
		TEID: tunnelInfo.TEID,
		PDR: tunnelInfo.PDR,
	}
	return tunnel
}

func ToBsonMNodeInDB(data DataPathNodeInDB) (ret bson.M) {

	// Marshal data into json format
	tmp, err := json.Marshal(data)
	if err != nil {
		logger.CtxLog.Errorf("ToBsonMNodeInDB marshall error: %v", err)
	}

	// unmarshal data into bson format
	err = json.Unmarshal(tmp, &ret)
	if err != nil {
		logger.CtxLog.Errorf("ToBsonMNodeInDB unmarshall error: %v", err)
	}

	return
}

func StoreNodeInDB(nodeInDB DataPathNodeInDB) {
	fmt.Println("db - storeNodeInDB In DB!!")
	itemBsonA := ToBsonMNodeInDB(nodeInDB)
	filter := bson.M{"nodeIDInDB": nodeInDB.DataPathNodeUPFNodeID}
	logger.CtxLog.Infof("filter : ", filter)

	MongoDBLibrary.RestfulAPIPost(NodeInDBCol, filter, itemBsonA)
	fmt.Println("db - finished storeNodeInDB In DB!!")
}

func GetNodeInDBFromDB(nodeIDInDB NodeIDInDB) *DataPathNodeInDB{
	filter := bson.M{}
	filter["nodeIDInDB"] = nodeIDInDB

	result := MongoDBLibrary.RestfulAPIGetOne(NodeInDBCol, filter)

	dataPathNodeInDB := &DataPathNodeInDB{}
	fmt.Println("GetNodeInDB, smf state json : ", result)

	err := json.Unmarshal(mapToByte(result), dataPathNodeInDB)
	if err != nil {
		logger.CtxLog.Errorf("GetNodeInDBFromDB unmarshall error: %v", err)
		return nil
	}
	return dataPathNodeInDB
}

func RecoverDataPathNode(nodeIDInDB NodeIDInDB) *DataPathNode {

	nodeInDB := GetNodeInDBFromDB(nodeIDInDB)
	dataPathNode := &DataPathNode{
		IsBranchingPoint: nodeInDB.IsBranchingPoint,
		UPF: RetrieveUPFNodeByNodeID(GetNodeID(nodeInDB.DataPathNodeUPFNodeID)),
	}
	var nilVal *TunnelInfo = nil
	if nodeInDB.ULTunnelInfo != nilVal {
		dataPathNode.UpLinkTunnel = RecoverUpLinkTunnel(nodeInDB.ULTunnelInfo)
	}
	if nodeInDB.DLTunnelInfo != nilVal {
		dataPathNode.DownLinkTunnel = RecoverDownLinkTunnel(nodeInDB.DLTunnelInfo)
	}
	return dataPathNode


	return nil
}


func DFSRecoverDataPathNode(dataPathNodeInDB *DataPathNodeInDB)  *DataPathNode{
	var nilValDpn *DataPathNodeInDB = nil
	var nilVarTunnelInfo *TunnelInfo = nil
	if dataPathNodeInDB != nilValDpn {
		dataPathNode :=  &DataPathNode{
			UPF: RetrieveUPFNodeByNodeID(GetNodeID(dataPathNodeInDB.DataPathNodeUPFNodeID)),
			IsBranchingPoint: dataPathNodeInDB.IsBranchingPoint,
		}

		upLinkTunnel := &GTPTunnel{}
		downLinkTunnel := &GTPTunnel{}

		uLTunnelInfo := dataPathNodeInDB.ULTunnelInfo
		dLTunnelInfo := dataPathNodeInDB.DLTunnelInfo
		
		if  uLTunnelInfo != nilVarTunnelInfo {
			upLinkTunnel = RecoverUpLinkTunnel(dataPathNodeInDB.ULTunnelInfo)
		}

		if dLTunnelInfo != nilVarTunnelInfo {
			downLinkTunnel = RecoverDownLinkTunnel(dataPathNodeInDB.DLTunnelInfo)
		}

		dataPathNode.UpLinkTunnel = upLinkTunnel
		dataPathNode.DownLinkTunnel = downLinkTunnel

		return dataPathNode
	}

	return nil
}

func DFSStoreDataPathNode(dataPathNode *DataPathNode) {
	var nilValDpn *DataPathNode = nil
	var nilValTunnel *GTPTunnel = nil
	if dataPathNode != nilValDpn {

		dataPathNodeInDB :=  DataPathNodeInDB{
			DataPathNodeUPFNodeID: GetNodeIDInDB(dataPathNode.UPF.NodeID),
			IsBranchingPoint: dataPathNode.IsBranchingPoint,
		}

		uLTunnelInfo := &TunnelInfo{}
		dLTunnelInfo := &TunnelInfo{}

		upLinkTunnel := dataPathNode.UpLinkTunnel
		downLinkTunnel := dataPathNode.DownLinkTunnel

		if upLinkTunnel != nilValTunnel {
			// store uLTunnelInfo
			uLTunnelInfo.TEID = upLinkTunnel.TEID
			uLTunnelInfo.PDR = upLinkTunnel.PDR

			upLinkTunnelDEP := upLinkTunnel.DestEndPoint

			if upLinkTunnelDEP != nilValDpn {
				DFSStoreDataPathNode(upLinkTunnel.DestEndPoint)
			}

			dataPathNodeInDB.ULTunnelInfo = uLTunnelInfo

		}

		if downLinkTunnel != nilValTunnel {
			// store dLTunnelInfo
			dLTunnelInfo.TEID = downLinkTunnel.TEID
			dLTunnelInfo.PDR = downLinkTunnel.PDR

			dlLinkTunnelSEP := downLinkTunnel.SrcEndPoint

			if dlLinkTunnelSEP != nilValDpn {
				DFSStoreDataPathNode(downLinkTunnel.SrcEndPoint)
			}
			dataPathNodeInDB.DLTunnelInfo = dLTunnelInfo
		}

		// store dataPathNodeInDB into DB
	}

}

