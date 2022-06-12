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


func GetNodeIDInDB(nodeID pfcpType.NodeID) (nodeIDInDB NodeIDInDB) {
	fmt.Println("db - In GetNodeIDInDB")
	nodeIDInDB = NodeIDInDB{
		NodeIdType: nodeID.NodeIdType,
		NodeIdValue: nodeID.NodeIdValue,
	}
	return nodeIDInDB
}

func GetNodeID(nodeIDInDB NodeIDInDB) (nodeID pfcpType.NodeID){
	fmt.Println("db - In GetNodeID")
	nodeID = pfcpType.NodeID{
		NodeIdType: nodeIDInDB.NodeIdType,
		NodeIdValue: nodeIDInDB.NodeIdValue,
	}
	return nodeID
}

func RecoverTunnel(tunnelInfo *TunnelInfo) (tunnel *GTPTunnel) {
	fmt.Println("db - In RecoverTunnel")
	tunnel = &GTPTunnel{
		TEID: tunnelInfo.TEID,
		PDR: tunnelInfo.PDR,
	}
	var nilVal *NodeIDInDB = nil
	fmt.Println("In RecoverTunnel &tunnelInfo.DataPathNodeUPFNodeID = %v", tunnelInfo.DataPathNodeUPFNodeID)
	if &tunnelInfo.DataPathNodeUPFNodeID != nilVal {
		endPoint := RecoverFirstDPNode(tunnelInfo.DataPathNodeUPFNodeID)
		tunnel.SrcEndPoint = endPoint
	}
	// TBA: recover dst endPoint
	return tunnel		
}

func RecoverFirstDPNode(nodeIDInDB NodeIDInDB) (dataPathNode *DataPathNode) {

	nodeInDB := GetNodeInDBFromDB(nodeIDInDB)
	dataPathNode = &DataPathNode{
		IsBranchingPoint: nodeInDB.IsBranchingPoint,
		UPF: RetrieveUPFNodeByNodeID(GetNodeID(nodeInDB.DataPathNodeUPFNodeID)),
	}
	var nilVal *TunnelInfo = nil
	if nodeInDB.ULTunnelInfo != nilVal {
		dataPathNode.UpLinkTunnel = RecoverTunnel(nodeInDB.ULTunnelInfo)
		dataPathNode.UpLinkTunnel.DestEndPoint = dataPathNode
	}
	if nodeInDB.DLTunnelInfo != nilVal {
		dataPathNode.DownLinkTunnel = RecoverTunnel(nodeInDB.DLTunnelInfo)
		dataPathNode.DownLinkTunnel.DestEndPoint = dataPathNode
	}
	return dataPathNode
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

func GetNodeInDBFromDB(nodeIDInDB NodeIDInDB) (dataPathNodeInDB *DataPathNodeInDB){
	filter := bson.M{}
	filter["nodeIDInDB"] = nodeIDInDB

	result := MongoDBLibrary.RestfulAPIGetOne(NodeInDBCol, filter)

	dataPathNodeInDB = new(DataPathNodeInDB)
	fmt.Println("GetNodeInDB, smf state json : ", result)

	err := json.Unmarshal(mapToByte(result), dataPathNodeInDB)
	if err != nil {
		logger.CtxLog.Errorf("GetNodeInDBFromDB unmarshall error: %v", err)
		return nil
	}
	return dataPathNodeInDB
}

func RecoverDataPathNode(dataPathNodeInDB *DataPathNodeInDB)  (dataPathNode *DataPathNode){
	fmt.Println("db - in RecoverDataPathNode")
	var nilValDpn *DataPathNodeInDB = nil
	var nilVarTunnelInfo *TunnelInfo = nil
	if dataPathNodeInDB != nilValDpn {
		
		fmt.Println("db - in RecoverDataPathNode dataPathNodeInDB != nilValDpn")
		fmt.Println("db - in RecoverDataPathNode dataPathNodeInDB.DataPathNodeUPFNodeID = %v", dataPathNodeInDB.DataPathNodeUPFNodeID)
		fmt.Println("db - in RecoverDataPathNode GetNodeID = %v", GetNodeID(dataPathNodeInDB.DataPathNodeUPFNodeID))
		fmt.Println("db - in RecoverDataPathNode retrieved UPF = %v", RetrieveUPFNodeByNodeID(GetNodeID(dataPathNodeInDB.DataPathNodeUPFNodeID)))

		dataPathNode :=  &DataPathNode{
			UPF: RetrieveUPFNodeByNodeID(GetNodeID(dataPathNodeInDB.DataPathNodeUPFNodeID)),
			IsBranchingPoint: dataPathNodeInDB.IsBranchingPoint,
		}

		upLinkTunnel := new(GTPTunnel)
		downLinkTunnel := new(GTPTunnel)
		
		uLTunnelInfo := dataPathNodeInDB.ULTunnelInfo
		dLTunnelInfo := dataPathNodeInDB.DLTunnelInfo
		fmt.Println("db - in RecoverDataPathNode uLTunnelInfo %v ", uLTunnelInfo)
		fmt.Println("db - in RecoverDataPathNode dLTunnelInfo %v ", dLTunnelInfo)
		
		if  uLTunnelInfo != nilVarTunnelInfo {
			fmt.Println("db - in RecoverDataPathNode uLTunnelInfo != nilVarTunnelInfo ")
			upLinkTunnel = RecoverTunnel(dataPathNodeInDB.ULTunnelInfo)
		}

		if dLTunnelInfo != nilVarTunnelInfo {
			fmt.Println("db - in RecoverDataPathNode dLTunnelInfo != nilVarTunnelInfo ")
			downLinkTunnel = RecoverTunnel(dataPathNodeInDB.DLTunnelInfo)
		}

		dataPathNode.UpLinkTunnel = upLinkTunnel
		dataPathNode.DownLinkTunnel = downLinkTunnel

		return dataPathNode
	}

	return nil
}

func StoreDataPathNode(dataPathNode *DataPathNode) (dataPathNodeInDB *DataPathNodeInDB){
	var nilValDpn *DataPathNode = nil
	var nilValTunnel *GTPTunnel = nil
	fmt.Println("db - in StoreDataPathNode")
	if dataPathNode != nilValDpn {
		fmt.Println("db - in StoreDataPathNode dataPathNode != nilValDpn %v", dataPathNode)

		dataPathNodeInDB :=  &DataPathNodeInDB{
			DataPathNodeUPFNodeID: GetNodeIDInDB(dataPathNode.UPF.NodeID),
			IsBranchingPoint: dataPathNode.IsBranchingPoint,
		}

		uLTunnelInfo := new(TunnelInfo)
		dLTunnelInfo := new(TunnelInfo)

		upLinkTunnel := dataPathNode.UpLinkTunnel
		downLinkTunnel := dataPathNode.DownLinkTunnel
		fmt.Println("db - in StoreDataPathNode checking upLinkTunnel")
		if upLinkTunnel != nilValTunnel {
			// fmt.Println("db - in StoreDataPathNode upLinkTunnel != nilValTunnel %v", upLinkTunnel)
			// store uLTunnelInfo
			uLTunnelInfo.TEID = upLinkTunnel.TEID
			uLTunnelInfo.PDR = upLinkTunnel.PDR

			// upLinkTunnelDEP := upLinkTunnel.DestEndPoint
			upLinkTunnelSEP := upLinkTunnel.SrcEndPoint
			// fmt.Println("db - in StoreDataPathNode upLinkTunnelSEP %v", upLinkTunnelSEP)
			// fmt.Println("db - in StoreDataPathNode upLinkTunnelDEP %v", upLinkTunnelDEP)
			// fmt.Println("db - in StoreDataPathNode upLinkTunnelSEP ==  %v", (upLinkTunnelSEP == dataPathNode))
			// fmt.Println("db - in StoreDataPathNode upLinkTunnelDEP ==  %v", (upLinkTunnelDEP == dataPathNode))
			if upLinkTunnelSEP != nilValDpn {
				// fmt.Println("db - in StoreDataPathNode upLinkTunnelSEP != nilValDpn")
				// fmt.Println("db - in StoreDataPathNode upLinkTunnelSEP != nilValDpn %v", upLinkTunnelDEP)
				uLTunnelInfo.DataPathNodeUPFNodeID = GetNodeIDInDB(upLinkTunnelSEP.UPF.NodeID)
			} 
			// else {
			// 	fmt.Println("db - in StoreDataPathNode upLinkTunnelSEP == nilValDpn")
			// }

			dataPathNodeInDB.ULTunnelInfo = uLTunnelInfo

		}
		
		fmt.Println("db - in StoreDataPathNode checking downLinkTunnel")
		if downLinkTunnel != nilValTunnel {
			// fmt.Println("db - in StoreDataPathNode downLinkTunnel != nilValTunnel %v ", downLinkTunnel)

			// store dLTunnelInfo
			dLTunnelInfo.TEID = downLinkTunnel.TEID
			dLTunnelInfo.PDR = downLinkTunnel.PDR

			dlLinkTunnelSEP := downLinkTunnel.SrcEndPoint
			// dlLinkTunnelDEP := downLinkTunnel.DestEndPoint
			// fmt.Println("db - in StoreDataPathNode dlLinkTunnelSEP %v", dlLinkTunnelSEP)
			// fmt.Println("db - in StoreDataPathNode dlLinkTunnelDEP %v", dlLinkTunnelDEP)
			// fmt.Println("db - in StoreDataPathNode dlLinkTunnelSEP ==  %v", (dlLinkTunnelSEP == dataPathNode))
			// fmt.Println("db - in StoreDataPathNode dlLinkTunnelDEP ==  %v", (dlLinkTunnelDEP == dataPathNode))
			if dlLinkTunnelSEP != nilValDpn {
				// fmt.Println("db - in StoreDataPathNode dlLinkTunnelSEP != nilValDpn")
				dLTunnelInfo.DataPathNodeUPFNodeID = GetNodeIDInDB(dlLinkTunnelSEP.UPF.NodeID)
			} 
			// else {
			// 	fmt.Println("db - in StoreDataPathNode dlLinkTunnelDEP == nilValDpn")
			// }
			dataPathNodeInDB.DLTunnelInfo = dLTunnelInfo
		}
		fmt.Println("db - in StoreDataPathNode return dataPathNodeInDB")

		return dataPathNodeInDB
	}
	return nil

}

