/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"errors"
	"fmt"
	"strconv"
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var cprIndexStr = "_cprindex"				//name for the key/value that will store a list of all known cprs

type Cpr struct{
	id string `json:"id"`					//the fieldtags are needed to keep case from bouncing around
	hash string `json:"hash"`
	owner string `json:"owner"`
	grower string `json:"grower"`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var Aval int
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState("abc", []byte(strconv.Itoa(Aval)))				//making a test var "abc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}
	
	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)								//marshal an emtpy array of strings to clear the index
	err = stub.PutState(cprIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}
	
	return nil, nil
}

// ============================================================================================================================
// Invoke - Our entry point
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		return t.Init(stub, function, args)
	} else if function == "delete" {										//deletes an entity from its state
		return t.Delete(stub, args)
	} else if function == "write" {											//writes a value to the chaincode state
		return t.Write(stub, args)
	} else if function == "init_cpr" {					  				    //create a new cpr
		fmt.Println("pre call cpr");
		return t.init_cpr(stub, args)

	} else if function == "set_user" {										//change owner of a cpr
		return t.set_user(stub, args)
	}
	fmt.Println("run did not find func: " + function)						//error

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" {													//read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil													//send it onward
}

// ============================================================================================================================
// Delete - remove a key/value pair from state
// ============================================================================================================================
func (t *SimpleChaincode) Delete(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	
	name := args[0]
	err := stub.DelState(name)													//remove the key from chaincode state
	if err != nil {
		return nil, errors.New("Failed to delete state")
	}

	//get the cpr index
	cprsAsBytes, err := stub.GetState(cprIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get cpr index")
	}
	var cprIndex []string
	json.Unmarshal(cprsAsBytes, &cprIndex)								//un stringify it aka JSON.parse()
	
	//remove cpr from index
	for i,val := range cprIndex{
		fmt.Println(strconv.Itoa(i) + " - looking at " + val + " for " + name)
		if val == name{															//find the correct cpr
			fmt.Println("found cpr")
			cprIndex = append(cprIndex[:i], cprIndex[i+1:]...)			//remove it
			for x:= range cprIndex{											//debug prints...
				fmt.Println(string(x) + " - " + cprIndex[x])
			}
			break
		}
	}
	jsonAsBytes, _ := json.Marshal(cprIndex)									//save new index
	err = stub.PutState(cprIndexStr, jsonAsBytes)
	return nil, nil
}

// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) Write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var id, value string // Entities
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}

	id = args[0]														//rename for funsies
	value = args[1]
	err = stub.PutState(id, []byte(value))								//write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ============================================================================================================================
// Init CPR - create a new cpr, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) init_cpr(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	

	fmt.Println("- STARTINGGGGGG INIT_CPR METHOOOOD! - 1")


	var err error
 
	//   0         1           2,      4,
	// "id", "hash do pdf",  owner   grower

	fmt.Println("- STARTINGGGGGG INIT_CPR METHOOOOD! - 2")

	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}
	
	fmt.Println("- STARTINGGGGGG INIT_CPR METHOOOOD! - 3")

	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return nil, errors.New("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return nil, errors.New("4rd argument must be a non-empty string")
	}

	fmt.Println("- STARTINGGGGGG INIT_CPR METHOOOOD! - 4")
		
	str := `{"id": "` + args[0] + `", "hash": "` + args[1] + `",  "owner": "`+ args[2] + `", "grower": "`+ args[3] +`"}`
	err = stub.PutState(args[0], []byte(str))								//store cpr with id as key
	if err != nil {
		return nil, err
	}
		
	//get the cpr index
	cprAsBytes, err := stub.GetState(cprIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get cpr index")
	}
	var cprIndex []string
	json.Unmarshal(cprAsBytes, &cprIndex)							    //un stringify it aka JSON.parse()
	
	//append
	cprIndex = append(cprIndex, args[0])								//add cpr id to index list
	fmt.Println("! cpr index: ", cprIndex)
	jsonAsBytes, _ := json.Marshal(cprIndex)
	err = stub.PutState(cprIndexStr, jsonAsBytes)						//store name of cpr

	fmt.Println("- end init cpr")
	return nil, nil
}

// ============================================================================================================================
// Set User Permission on cpr
// ============================================================================================================================
func (t *SimpleChaincode) set_user(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	
	//   0     1
	// "id", "owner"
	if len(args) < 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}
	
	fmt.Println("- start set owner")
	fmt.Println(args[0] + " - " + args[1])
	cprAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get thing")
	}
	res := Cpr{}
	json.Unmarshal(cprAsBytes, &res)										//un stringify it aka JSON.parse()
	res.owner = args[1]														//change the owner
	
	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes)								//rewrite the cpr with id as key
	if err != nil {
		return nil, err
	}
	
	fmt.Println("- end set owner")
	return nil, nil
}