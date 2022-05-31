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

	"github.com/free5gc/amf/logger"
	// "github.com/free5gc/openapi/models"
)

const (
	SmContextDataColl = "smf.data.smContext"
	GTPTunnelDataColl = "smf.data.gtpTunnel"
	TransactionDataCol = "smf.data.transaction"
)

func SetupSmfCollection() {
	fmt.Println("db - SetupSmfCollection!!")
	MongoDBLibrary.SetMongoDB("sdcore", "mongodb://mongodb")
	_, err := MongoDBLibrary.CreateIndex(SmContextDataColl, "supi")
	if err != nil {
		logger.ContextLog.Errorf("Create index failed on Supi field.")
	}
	MongoDBLibrary.SetMongoDB("sdcore", "mongodb://mongodb")
	_, err = MongoDBLibrary.CreateIndex(SmContextDataColl, "identifier")
	if err != nil {
		logger.ContextLog.Errorf("Create index failed on Identifier field.")
	}
	MongoDBLibrary.SetMongoDB("sdcore", "mongodb://mongodb")
	_, err = MongoDBLibrary.CreateIndex(SmContextDataColl, "pduSessionID")
	if err != nil {
		logger.ContextLog.Errorf("Create index failed on PDUSessionID field.")
	}

	MongoDBLibrary.SetMongoDB("sdcore", "mongodb://mongodb")
	_, err = MongoDBLibrary.CreateIndex(GTPTunnelDataColl, "teid")
	if err != nil {
		logger.ContextLog.Errorf("Create index failed on teid field.")
	}
	MongoDBLibrary.SetMongoDB("sdcore", "mongodb://mongodb")
	_, err = MongoDBLibrary.CreateIndex(TransactionDataCol, "txnId")
	if err != nil {
		logger.ContextLog.Errorf("Create index failed on TxnId field.")
	}
	
}


func ToBsonM(data *SMContext) (ret bson.M) {
	tmp, err := json.Marshal(data)
	if err != nil {
		logger.ContextLog.Errorf("SMContext marshall error: %v", err)
	}
	err = json.Unmarshal(tmp, &ret)
	if err != nil {
		logger.ContextLog.Errorf("SMContext unmarshall error: %v", err)
	}

	return
}

func StoreContextInDB(smContext *SMContext) {
	fmt.Println("db - Store SMContext In DB!!")
	smContextBsonA := ToBsonM(smContext)
	filter := bson.M{"supi": smContext.Supi}
	logger.ContextLog.Infof("filter : ", filter)

	MongoDBLibrary.RestfulAPIPost(SmContextDataColl, filter, smContextBsonA)
	fmt.Println("db - finished Store SMContext In DB!!")
}

