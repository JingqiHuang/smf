# db changes README

## db change related files:

context/db.go
context/db_tunnel.go
context/sm_context.go
fsm/txnFsm.go

##### Major DB operations are written as functions in db.go, including
    MarshalJSON() // customized mashaller for sm context
	UnmarshalJSON() // customized unmashaller for sm context
	StoreSmContextInDB() // Store SmContext In DB
	StoreSeidContextInDB() // Store SmContext with SEID as key in DB
	StoreRefToSeidInDB() // Store SmContext with ref string as key in DB
	GetSeidByRefInDB() // Get SEID using ref string as key from DB
	GetSMContextByRefInDB() // Get SmContext using ref string as key from DB 
	GetSMContextBySEIDInDB() // Get SmContext using SEID string as key from DB  
	DeleteSmContextInDBBySEID() // Delete SmContext using SEID string as key from DB   
	DeleteSmContextInDBByRef() // Delete SmContext using ref string as key from DB 
	ClearSMContextInMem() // Delete SMContext in smContextPool and seidSMContextMap, for test purpose
	
##### Major DB operation related to store Tunnel variable of sm context are written as functions in db_tunnel.go

Why do we need smf.data.nodeInDB?

- In the struct UplinkTunnel/DownlinkTunnel *GTPTunnel, it contains SrcEndPoint *DataPathNode. and the SrcEndPoint also contains UplinkTunnel/DownlinkTunnel *GTPTunnel. This causes a cyclic structure that cannot be stored into the database.

- To recover the SrcEndPoint *DataPathNode:
	* Store the ID that can track the SrcEndPoint in the GTPTunnel, which is SrcEndPoint.UPF.NodeID. 
	* Store the DataPathNodeInDB with the ID into the database

    
### Note

- Clear sm context in memory using  after each successful transaction for test

- Convert SEID, which is uint64, into hex string for DB storage

- Store sm context in fsm/txnFsm.go/TxnSave() (after the transaction is successful)

- Load sm context from db at the beginning of TxnProcess() is sm context if not found in memory

### Known issues
We test the code with the gnbsim profile in aiab and found the tests of the following events cannot pass. Will fix them later
    nwtriggeruedereg: timeout
    uereqpdusessrelease and dereg: runtime error: invalid memory address or nil pointer dereference in smf/context/ip_allocator.go:86; Error raise when HandlePDUSessionReleaseRequest

