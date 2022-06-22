// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package fsm

import (
	"fmt"
	"net/http"

	"github.com/omec-project/http_wrapper"
	"github.com/omec-project/openapi/models"
	smf_context "github.com/omec-project/smf/context"
	"github.com/omec-project/smf/logger"
	"github.com/omec-project/smf/msgtypes/svcmsgtypes"
	"github.com/omec-project/smf/producer"
	"github.com/omec-project/smf/transaction"
)

func (SmfTxnFsm) TxnInit(txn *transaction.Transaction) (transaction.TxnEvent, error) {
	txn.TxnFsmLog.Debugf("handle event[%v] ", transaction.TxnEventInit.String())
	fmt.Println("db - TxnInit handle event[%v] ", transaction.TxnEventInit.String())
	return transaction.TxnEventDecode, nil
}

func (SmfTxnFsm) TxnDecode(txn *transaction.Transaction) (transaction.TxnEvent, error) {

	fmt.Println("db - TxnInit TxnDecode ")
	return transaction.TxnEventLoadCtxt, nil
}

func (SmfTxnFsm) TxnLoadCtxt(txn *transaction.Transaction) (transaction.TxnEvent, error) {
	fmt.Println("db - TxnInit TxnLoadCtxt ")
	switch txn.MsgType {
	case svcmsgtypes.CreateSmContext:
		req := txn.Req.(models.PostSmContextsRequest)
		createData := req.JsonData
		if smCtxtRef, err := smf_context.ResolveRef(createData.Supi, createData.PduSessionId); err == nil {
			//Previous context exist
			producer.HandlePduSessionContextReplacement(smCtxtRef)
		}
		//Create fresh context
		txn.Ctxt = smf_context.NewSMContext(createData.Supi, createData.PduSessionId)

	case svcmsgtypes.UpdateSmContext:
		fallthrough
	case svcmsgtypes.ReleaseSmContext:
		fallthrough
	case svcmsgtypes.SmPolicyUpdateNotification:
		txn.Ctxt = smf_context.GetSMContext(txn.CtxtKey)
		// fmt.Println("db - TxnInit TxnLoadCtxt txn.Ctxt = ", txn.Ctxt)
	case svcmsgtypes.PfcpSessCreate:
		fallthrough
	case svcmsgtypes.N1N2MessageTransfer:
		//Pre-loaded- No action
	case svcmsgtypes.N1N2MessageTransferFailureNotification:
		txn.Ctxt = smf_context.GetSMContext(txn.CtxtKey)
	default:
		txn.TxnFsmLog.Errorf("handle event[%v], next-event[%v], unknown msgtype [%v] ",
			transaction.TxnEventLoadCtxt.String(), transaction.TxnEventFailure.String(), txn.MsgType)
		return transaction.TxnEventFailure, fmt.Errorf("invalid Msg to load Txn")
	}

	if txn.Ctxt.(*smf_context.SMContext) == nil {
		txn.TxnFsmLog.Errorf("handle event[%v], ctxt [%v] not found", transaction.TxnEventLoadCtxt.String(), txn.CtxtKey)
		return transaction.TxnEventFailure, fmt.Errorf("ctxt not found")
	} else {
		fmt.Println("db - TxnInit TxnLoadCtxt txn.Ctxt found ")
	}

	return transaction.TxnEventCtxtPost, nil
}

func (SmfTxnFsm) TxnCtxtPost(txn *transaction.Transaction) (transaction.TxnEvent, error) {
	fmt.Println("db - in TxnCtxtPost")
	smContext := txn.Ctxt.(*smf_context.SMContext)

	//Lock the bus before modifying
	smContext.SMTxnBusLock.Lock()
	defer smContext.SMTxnBusLock.Unlock()

	//If already Active Txn running then post it to SMF Txn Bus

	if smContext.ActiveTxn != nil {

		smContext.TxnBus = smContext.TxnBus.AddTxn(txn)

		//Txn has been posted and shall be scheduled later
		fmt.Println("db - in TxnCtxtPost event[%v], next-event[%v], txn queued ", transaction.TxnEventCtxtPost.String(), transaction.TxnEventExit.String())
		return transaction.TxnEventQueue, nil
	} else {
		fmt.Println("db - in TxnCtxtPost smContext.ActiveTxn == nil")

	}

	//No other Txn running, lets proceed with current Txn

	return transaction.TxnEventRun, nil
}

func (SmfTxnFsm) TxnCtxtRun(txn *transaction.Transaction) (transaction.TxnEvent, error) {
	fmt.Println("db - in TxnCtxtRun")
	smContext := txn.Ctxt.(*smf_context.SMContext)

	//There shouldn't be any active Txn if current Txn has reached to Run state
	//Probably, abort it
	smContext.SMTxnBusLock.Lock()
	defer smContext.SMTxnBusLock.Unlock()

	if smContext.ActiveTxn != nil {
		logger.TxnFsmLog.Errorf("active transaction [%v] not completed", smContext.ActiveTxn)
	}

	//make current txn as Active now, move it to processing
	smContext.ActiveTxn = txn
	return transaction.TxnEventProcess, nil
}

func (SmfTxnFsm) TxnProcess(txn *transaction.Transaction) (transaction.TxnEvent, error) {
	fmt.Println("db - in TxnProcess")
	smContext := txn.Ctxt.(*smf_context.SMContext)
	if smContext == nil {
		txn.TxnFsmLog.Errorf("event[%v], next-event[%v], SM context invalid ", transaction.TxnEventProcess.String(), transaction.TxnEventFailure.String())
		return transaction.TxnEventFailure, fmt.Errorf("TxnProcess, invalid SM Ctxt")
	}

	var event SmEvent
	fmt.Println("db - in TxnProcess txn.MsgType %s ", txn.MsgType)
	switch txn.MsgType {
	case svcmsgtypes.CreateSmContext:
		event = SmEventPduSessCreate
	case svcmsgtypes.UpdateSmContext:
		event = SmEventPduSessModify
		//req := txn.Req.(models.UpdateSmContextRequest)
	case svcmsgtypes.ReleaseSmContext:
		event = SmEventPduSessRelease
	case svcmsgtypes.PfcpSessCreate:
		event = SmEventPfcpSessCreate
	case svcmsgtypes.N1N2MessageTransfer:
		event = SmEventPduSessN1N2Transfer
	case svcmsgtypes.N1N2MessageTransferFailureNotification:
		event = SmEventPduSessN1N2TransferFailureIndication
	case svcmsgtypes.SmPolicyUpdateNotification:
		event = SmEventPolicyUpdateNotify
	default:
		event = SmEventInvalid

	}

	eventData := SmEventData{Txn: txn}

	if err := HandleEvent(smContext, event, eventData); err != nil {
		fmt.Println("db - in TxnProcess HandleEvent %v err %s", transaction.TxnEventProcess.String(), err.Error())
		smContext.SubFsmLog.Errorf("handle event[%v], err [%s]", transaction.TxnEventProcess.String(), err.Error())
		return transaction.TxnEventFailure, err
	} else {
		fmt.Println("db - in TxnProcess HandleEvent %v", transaction.TxnEventProcess.String())
	}
	return transaction.TxnEventSuccess, nil
}

func (SmfTxnFsm) TxnSuccess(txn *transaction.Transaction) (transaction.TxnEvent, error) {
	fmt.Println("db - in TxnSuccess")
	switch txn.MsgType {
	case svcmsgtypes.PfcpSessCreate:
		fmt.Println("db - in TxnSuccess -  PfcpSessCreate")
		nextTxn := transaction.NewTransaction(nil, nil, svcmsgtypes.SmfMsgType(svcmsgtypes.N1N2MessageTransfer))
		nextTxn.Ctxt = txn.Ctxt
		smContext := txn.Ctxt.(*smf_context.SMContext)
		// fmt.Println("db - in TxnSuccess -  smContext ", smContext)
		smContext.SMTxnBusLock.Lock()
		smContext.TxnBus = smContext.TxnBus.AddTxn(nextTxn)
		smContext.SMTxnBusLock.Unlock()
		go func(nextTxn *transaction.Transaction) {
			//Initiate N1N2 Transfer

			//nextTxn.StartTxnLifeCycle(SmfTxnFsmHandle)
			<-nextTxn.Status
		}(nextTxn)
	}

	//put Success Rsp
	fmt.Println("db - in TxnSuccess -  smContext.TxnBus", txn.Ctxt.(*smf_context.SMContext).TxnBus)
	// smf_context.StoreSmContextInDB(txn.Ctxt.(*smf_context.SMContext))
	txn.Status <- true
	return transaction.TxnEventSave, nil
}

func (SmfTxnFsm) TxnFailure(txn *transaction.Transaction) (transaction.TxnEvent, error) {
	fmt.Println("db - in TxnFailure")
	//Put Failure Rsp
	switch txn.MsgType {
	case svcmsgtypes.UpdateSmContext:
		if txn.Ctxt == nil {
			logger.PduSessLog.Warnf("PDUSessionSMContextUpdate, SMContext[%s] is not found", txn.CtxtKey)

			httpResponse := &http_wrapper.Response{
				Header: nil,
				Status: http.StatusNotFound,
				Body: models.UpdateSmContextErrorResponse{
					JsonData: &models.SmContextUpdateError{
						UpCnxState: models.UpCnxState_DEACTIVATED,
						Error: &models.ProblemDetails{
							Type:   "Resource Not Found",
							Title:  "SMContext Ref is not found",
							Status: http.StatusNotFound,
						},
					},
				},
			}
			txn.Rsp = httpResponse
		}

	case svcmsgtypes.ReleaseSmContext:
		if txn.Ctxt == nil {
			logger.PduSessLog.Warnf("PDUSessionSMContextRelease [%s] is not found", txn.CtxtKey)

			//4xx/5xx Error not defined in spec 29502 for Release SM ctxt error
			//Send Not Found
			httpResponse := &http_wrapper.Response{
				Header: nil,
				Status: http.StatusNotFound,

				Body: &models.ProblemDetails{
					Type:   "Resource Not Found",
					Title:  "SMContext Ref is not found",
					Status: http.StatusNotFound,
				},
			}
			txn.Rsp = httpResponse
		}
	}
	txn.Status <- false
	return transaction.TxnEventEnd, nil
}

func (SmfTxnFsm) TxnAbort(txn *transaction.Transaction) (transaction.TxnEvent, error) {
	return transaction.TxnEventEnd, nil
}

func (SmfTxnFsm) TxnSave(txn *transaction.Transaction) (transaction.TxnEvent, error) {
	smf_context.StoreSmContextInDB(txn.Ctxt.(*smf_context.SMContext))
	return transaction.TxnEventEnd, nil
}

func (SmfTxnFsm) TxnTimeout(txn *transaction.Transaction) (transaction.TxnEvent, error) {
	return transaction.TxnEventEnd, nil
}

func (SmfTxnFsm) TxnCollision(txn *transaction.Transaction) (transaction.TxnEvent, error) {
	return transaction.TxnEventEnd, nil
}

func (SmfTxnFsm) TxnEnd(txn *transaction.Transaction) (transaction.TxnEvent, error) {

	txn.TransactionEnd()

	smContext := txn.Ctxt.(*smf_context.SMContext)
	if smContext == nil {
		return transaction.TxnEventExit, nil
	}

	//Lock txnbus to access
	smContext.SMTxnBusLock.Lock()
	defer smContext.SMTxnBusLock.Unlock()

	//Reset Active Txn
	smContext.ActiveTxn = nil

	var nextTxn *transaction.Transaction
	//Active Txn is over, now Pull out head Txn and Run it
	if len(smContext.TxnBus) > 0 {

		nextTxn, smContext.TxnBus = smContext.TxnBus.PopTxn()
		txn.NextTxn = nextTxn
		return transaction.TxnEventRun, nil
	}

	return transaction.TxnEventExit, nil
}

/// Suggestions
//1. Global pipeline for txns
//2. Memory alloc pool for txns
