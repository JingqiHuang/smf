
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
	"github.com/omec-project/smf/transaction"
	"github.com/omec-project/smf/msgtypes/svcmsgtypes"
	"github.com/sirupsen/logrus"
)


// store/get/recover Txn

func ToBsonMTxnInDB(data *TransactionInDB) (ret bson.M){
	// Marshal data into json format
	tmp, err := json.Marshal(data)
	if err != nil {
		logger.CtxLog.Errorf("ToBsonMTxnInDB marshall error: %v", err)
	}

	// unmarshal data into bson format
	err = json.Unmarshal(tmp, &ret)
	if err != nil {
		logger.CtxLog.Errorf("ToBsonMTxnInDB unmarshall error: %v", err)
	}

	return
}

func StoreTxnInDB(txnInDB *TransactionInDB) {
	fmt.Println("db - StoreTxnInDB In DB!!")
	itemBsonA := ToBsonMTxnInDB(txnInDB)
	filter := bson.M{"txnInDB": txnInDB.TxnId}
	logger.CtxLog.Infof("filter : ", filter)

	MongoDBLibrary.RestfulAPIPost(TransactionDataCol, filter, itemBsonA)
	fmt.Println("db - finished StoreTxnInDB In DB!!")
}

func GetTxnInDBFromDB(txnID uint32) *TransactionInDB {
	filter := bson.M{}
	filter["txnID"] = txnID

	result := MongoDBLibrary.RestfulAPIGetOne(TransactionDataCol, filter)

	txnInDB := &TransactionInDB{}
	fmt.Println("GetTxnInDBFromDB, smf state json : ", result)

	err := json.Unmarshal(mapToByte(result), txnInDB)
	if err != nil {
		logger.CtxLog.Errorf("GetTxnInDBFromDB unmarshall error: %v", err)
		return nil
	}

	return txnInDB
}

func RecoverActiveTxn(txnID uint32) *transaction.Transaction {

	txn := &transaction.Transaction{}
	txnInDB := GetTxnInDBFromDB(txnID)

	if txn.NextTxn !=nil {
		txn.NextTxn = RecoverActiveTxn(txnInDB.NextTxnId)
	}

	txn.SetStartTime(txnInDB.StartTime())
	txn.SetEndTime(txnInDB.EndTime())
	txn.Priority = txnInDB.Priority
	txn.Req = txnInDB.Req
	txn.Rsp = txnInDB.Rsp
	txn.Ctxt = txnInDB.Ctxt
	txn.MsgType = svcmsgtypes.SmfMsgTypeType(txnInDB.MsgType)
	txn.CtxtKey = txnInDB.CtxtKey
	txn.Err = txnInDB.Err
	txn.Status = make(chan bool)

	subField := logrus.Fields{"txnid": txnID,
	"txntype": txnInDB.MsgType, "ctxtkey": txn.CtxtKey}

	txn.TxnFsmLog = logger.TxnFsmLog.WithFields(subField)

	return txn

}
