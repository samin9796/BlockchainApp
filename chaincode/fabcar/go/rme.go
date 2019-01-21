/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

/*
 * The sample smart contract for documentation topic:
 * Writing Your First Blockchain Application
 */

package main

/* Imports
 * 4 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts
 */
import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/cd1/utils-golang"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
	"strconv"
	"time"
)

// Define the Smart Contract structure
type SmartContract struct {
}
type area struct {
	Doctype string
	Name string
	Division string
	District string
	Thana string
	key string
}
type candidate struct {
	Doctype string
	Name string
	AreaName string
	Totalvote int
	key string
}

type election struct {
	Doctype string
	Name string
	key string
}

type commission struct {
	Doctype string
	Name string
	email string
	password string
}

type user struct {
	Doctype string
	Name string
	Email string
	Key string
	Password string
	Balance float64
}

type Transaction struct {
	Doctype string
	SenderEmail string
	ReceiverEmail string
	Amount float64
	TransactionId string
	Shomoy string
}

/*
 * The Init method is called when the Smart Contract "fabcar" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "fabcar"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "register" {
		return s.register(APIstub, args)
	} else if function == "login" {
		return s.login(APIstub, args)
	}else if function == "makeTransaction" {
		return s.makeTransaction(APIstub, args)
	}else if function == "getData" {
		return s.getDataFromArgs(APIstub, args)
	} else if function == "setData" {
		return s.setData(APIstub, args)
	} else if function == "getBalance" {
		return s.getBalance(APIstub, args)
	} else if function == "subtractBalance" {
		return s.subtractBalance(APIstub, args)
	} else if function == "getHistory" {
		return s.getHistory(APIstub, args)
	} else if function == "getReceiveHistory" {
		return s.getReceiveHistory(APIstub, args)
	} else if function == "checkBalance" {
		return s.checkBalance(APIstub, args)
	}


	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) register(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments, required 11, given "+strconv.Itoa(len(args)))
	}

	mUser := user{}

	mUser.Doctype = "user"
	//mUser.Nid, _ = strconv.ParseInt(args[0], 10, 64)
	mUser.Name= args[0]
	mUser.Email = args[1]
	mUser.Password = args[2]


	h := sha256.New()
	h.Write([]byte(mUser.Password))
	mUser.Password = fmt.Sprintf("%x", h.Sum(nil))

	userKey := utils.RandomString()
	mUser.Key = userKey

	mUser.Balance = 100.0

	jsonUser, _ := json.Marshal(mUser)

	_ = APIstub.PutState(userKey, jsonUser)
	return shim.Success(nil)
}


func (s *SmartContract) makeTransaction(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	currentTransaction := Transaction{}
	currentTransaction.Doctype = "transaction"
	currentTransaction.Shomoy = time.Now().Format(time.RFC850)
	currentTransaction.SenderEmail = args[0]
	currentTransaction.ReceiverEmail = args[1]
	currentTransaction.Amount, _ = strconv.ParseFloat(args[2], 64)
	currentTransaction.TransactionId = utils.RandomString()

	jsonTransaction, _ := json.Marshal(currentTransaction)

	_ = APIstub.PutState(currentTransaction.TransactionId, jsonTransaction)

	userQuery := newCouchQueryBuilder().addSelector("Doctype", "user").addSelector("Email", currentTransaction.SenderEmail).getQueryString()

	currentUserData, _ := firstQueryValueForQueryString(APIstub, userQuery)
	var currentUser user
	_ = json.Unmarshal(currentUserData, &currentUser)

	currentUser.Balance -= currentTransaction.Amount

	jsonUser, _ := json.Marshal(currentUser)
	_ = APIstub.PutState(currentUser.Key, jsonUser)

	userQuery2 := newCouchQueryBuilder().addSelector("Doctype", "user").addSelector("Email", currentTransaction.ReceiverEmail).getQueryString()

	currentUserData2, _ := firstQueryValueForQueryString(APIstub, userQuery2)
	var currentUser2 user
	_ = json.Unmarshal(currentUserData2, &currentUser2)

	currentUser2.Balance += currentTransaction.Amount


	jsonUser2, _ := json.Marshal(currentUser2)
	_ = APIstub.PutState(currentUser2.Key, jsonUser2)




	return shim.Success(nil)
}

func (s *SmartContract) getBalance(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	email := args[0]
	balanceQuery := newCouchQueryBuilder().addSelector("Doctype", "transaction").addSelector("ReceiverEmail", email).getQueryString()
	var totalBalance float64 = 100.0
	iterator, _ := APIstub.GetQueryResult(balanceQuery)
	for iterator.HasNext() {
		element, _ := iterator.Next()
		val := element.Value
		var transaction Transaction
		_ = json.Unmarshal(val, &transaction)
		totalBalance += transaction.Amount
	}

	userQuery := newCouchQueryBuilder().addSelector("Doctype", "user").addSelector("Email", email).getQueryString()

	currentUserData, _ := firstQueryValueForQueryString(APIstub, userQuery)
	var currentUser user
	_ = json.Unmarshal(currentUserData, &currentUser)

	currentUser.Balance = totalBalance
	response := fmt.Sprintf("{\"balance\": %v}", totalBalance)


	return shim.Success( []byte(response) )
}

func (s *SmartContract) checkBalance(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	email := args[0]
	userQuery := newCouchQueryBuilder().addSelector("Doctype", "user").addSelector("Email", email).getQueryString()

	currentUserData, _ := firstQueryValueForQueryString(APIstub, userQuery)
	var currentUser user
	_ = json.Unmarshal(currentUserData, &currentUser)

	response := fmt.Sprintf("{\"balance\": %v}", currentUser.Balance)

	return shim.Success( []byte(response) )
}


func (s *SmartContract) subtractBalance(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	email := args[0]
	balanceQuery := newCouchQueryBuilder().addSelector("Doctype", "transaction").addSelector("SenderEmail", email).getQueryString()
	var totalBalance float64 = 100.0
	iterator, _ := APIstub.GetQueryResult(balanceQuery)
	for iterator.HasNext() {
		element, _ := iterator.Next()
		val := element.Value
		var transaction Transaction
		_ = json.Unmarshal(val, &transaction)
		totalBalance -= transaction.Amount
	}


	userQuery := newCouchQueryBuilder().addSelector("Doctype", "user").addSelector("Email", email).getQueryString()

	currentUserData, _ := firstQueryValueForQueryString(APIstub, userQuery)
	var currentUser user
	_ = json.Unmarshal(currentUserData, &currentUser)

	currentUser.Balance = totalBalance
	response := fmt.Sprintf("{\"balance\": %v}", totalBalance)

	return shim.Success( []byte(response) )
}

func (s *SmartContract) addArea(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	addArea := area{}
		addArea.Doctype="Area"
		addArea.Name=args[0]
		addArea.Division=args[1]
		addArea.District=args[2]
		addArea.Thana=args[3]
		key:=utils.RandomString()
	    addArea.key=key


	jsonUser, _ := json.Marshal(addArea)

	_ = APIstub.PutState(key, jsonUser)
	return shim.Success(nil)
}

func (s *SmartContract) addElection(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, required 1, given "+strconv.Itoa(len(args)))
	}
	addElection := election{}
	addElection.Doctype="election"
	addElection.Name=args[0]
	key:=utils.RandomString()
	addElection.key=key


	jsonElection, _ := json.Marshal(addElection)

	_ = APIstub.PutState(key, jsonElection)
	return shim.Success(nil)
}

func (s *SmartContract) getReceiveHistory(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	key:=args[0]

	currentUserData, _ := APIstub.GetState(key)

	var currentUser user
	_ = json.Unmarshal(currentUserData, &currentUser)

	//email := currentUser.Email

	//areaQuery := newCouchQueryBuilder().addSelector("Doctype", "area").addSelector("Thana", presentThana).getQueryString()

	//currentAreaData, _ := firstQueryValueForQueryString(APIstub, areaQuery)
	//var currentArea area
	//_ = json.Unmarshal(currentAreaData, &currentArea)

	myHistory := newCouchQueryBuilder().addSelector("Doctype", "transaction").addSelector("ReceiverEmail", currentUser.Email).getQueryString()

	myHistoryData, _ := getJSONQueryResultForQueryString(APIstub, myHistory)

	//print the output
	fmt.Println( string(myHistoryData) )

	return shim.Success([]byte(myHistoryData))
}


func (s *SmartContract) getHistory(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	//key:=args[0]

	//currentUserData, _ := APIstub.GetState(key)

	email := args[0]
	myHistory := newCouchQueryBuilder().addSelector("Doctype", "transaction").addSelector("SenderEmail", email).getQueryString()

	//currentUserData, _ := firstQueryValueForQueryString(APIstub, userQuery)
	//var currentUser user
	//_ = json.Unmarshal(currentUserData, &currentUser)


	//email := currentUser.Email

	//areaQuery := newCouchQueryBuilder().addSelector("Doctype", "area").addSelector("Thana", presentThana).getQueryString()

	//currentAreaData, _ := firstQueryValueForQueryString(APIstub, areaQuery)
	//var currentArea area
	//_ = json.Unmarshal(currentAreaData, &currentArea)

	//myHistory := newCouchQueryBuilder().addSelector("Doctype", "transaction").addSelector("SenderEmail", currentUser.Email).getQueryString()

	myHistoryData, _ := getJSONQueryResultForQueryString(APIstub, myHistory)

	//print the output
	fmt.Println( string(myHistoryData) )

	return shim.Success([]byte(myHistoryData))
}

func (s *SmartContract) addCandidate(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments, required 2, given "+strconv.Itoa(len(args)))
	}
	addCandidate := candidate{}
	addCandidate.Doctype ="candidate"
	addCandidate.Name=args[0]
	addCandidate.AreaName =args[1]
	addCandidate.Totalvote=0
	key:=utils.RandomString()
	addCandidate.key=key


	jsonCandidate, _ := json.Marshal(addCandidate)

	_ = APIstub.PutState(key,jsonCandidate)
	return shim.Success(nil)
}

func (s *SmartContract) login(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("you need 2 arguments, but you have given "+ strconv.Itoa(len(args)) )
	}

	//Nid, _ := strconv.ParseInt(args[0], 10, 64)
	Email := args[0]
	Password := args[1]

	h := sha256.New()
	h.Write([]byte(Password))
	Password = fmt.Sprintf("%x", h.Sum(nil))

	//queryFormat := "{ \"selector\": { \"Nid\": %d, \"Birthdate\": \"%s\", \"Password\": \"%s\" } }"



	queryString :=newCouchQueryBuilder().addSelector("Doctype","user").addSelector("Email",Email).addSelector("Password",Password).getQueryString()

	//iterator,_:=APIstub.GetQueryResult(queryString)
	//var lastValue []byte
	//var lastKey string
	//for i:=0; iterator.HasNext(); i++ {
//		currentData, _:=iterator.Next()
//		//lastKey = currentData.Key
//		lastValue = currentData.Value
//	}

	userData, _ := firstQueryValueForQueryString(APIstub, queryString)
	return shim.Success(userData)
	//return shim.Success([]byte(queryString))

	//jsonData, _ := getQueryResultForQueryString(APIstub, queryString)
	//return shim.Success([]byte(jsonData))
}

func (s *SmartContract) getData(APIstub shim.ChaincodeStubInterface, args ...string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	key := args[0]

	data, err := APIstub.GetState(key)
	if err != nil {
		return shim.Error("There was an error")
	}

	return shim.Success(data)
}
//////

//func (s *SmartContract) getCandidates(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	//key:=args[0]

	//currentUserData, _ := APIstub.GetState(key)
	//var currentUser user
	//_ = json.Unmarshal(currentUserData, &currentUser)

	//presentThana := currentUser.PresentThana

	//areaQuery := newCouchQueryBuilder().addSelector("Doctype", "area").addSelector("Thana", presentThana).getQueryString()

 	//currentAreaData, _ := firstQueryValueForQueryString(APIstub, areaQuery)
 	//var currentArea area
 	//_ = json.Unmarshal(currentAreaData, &currentArea)

 	//candidateQuery := newCouchQueryBuilder().addSelector("Doctype", "candidate").addSelector("AreaName", currentArea.Name).getQueryString()

 	//candidatesData, _ := getJSONQueryResultForQueryString(APIstub, candidateQuery)

 	//print the output
 	//fmt.Println( string(candidatesData) )

 	//res1, res2 := doSomething(2,3)
 	//fmt.Println(res1, res2)

	//return shim.Success(candidatesData)
//}



////

func (s *SmartContract) getDataFromArgs(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	key := args[0]

	data, err := APIstub.GetState(key)
	if err != nil {
		return shim.Error("There was an error")
	}

	return shim.Success(data)
}

func (s *SmartContract) setData(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	key := args[0]
	val := args[1]

	err := APIstub.PutState(key, []byte(val))
	if err != nil {
		return shim.Error("There was an error")
	}

	str := "operation successful"

	return shim.Success([]byte(str))
}

func MockInvoke(stub *shim.MockStub, function string, args []string) sc.Response {
	input := args
	output := make([][]byte, len(input)+1)
	output[0]= []byte(function)
	for i, v := range input {
		output[i+1] = []byte(v)
	}

	fmt.Println("final arguments: ", output) // [[102 111 111] [98 97 114]]

	return stub.MockInvoke("1", output)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {
	//scc := new(SmartContract)
	//stub := shim.NewMockStub("mychannel", scc)
	//res := MockInvoke(stub, "register", []string {"Tanmoy Krishna Das", "tanmoykrishnadas@gmail.com", "12345678"})
	//if res.Status != shim.OK {
	//	fmt.Println("bad status received, expected: 200; received:" + strconv.FormatInt(int64(res.Status), 10))
	//	fmt.Println("response: " + string(res.Message))
	//}
	//fmt.Println("Payload", string(res.Payload))
	//fmt.Println("Message", res.Message)

	//res = MockInvoke(stub, "login", []string {"tanmoykrishnadas@gmail.com", "12345678"})
	//if res.Status != shim.OK {
	//	fmt.Println("bad status received, expected: 200; received:" + strconv.FormatInt(int64(res.Status), 10))
	//	fmt.Println("response: " + string(res.Message))
	//}
	//fmt.Println("Payload", string(res.Payload))
	//fmt.Println("Message", res.Message)

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
