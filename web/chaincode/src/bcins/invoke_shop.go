package main

import (
	"encoding/json"

	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func createContract(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// need to pass 1 arg
	if len(args) != 1 {
		return shim.Error("Invalid argument count.")
	}

	// create a struct to store useful info
	// UUID -> key (always camel case)
	// string -> data type
	// json:"uuid" -> key name as stored in couchdb
	dto := struct {
		UUID             string    `json:"uuid"`
		ContractTypeUUID string    `json:"contract_type_uuid"`
		Username         string    `json:"username"`
		Password         string    `json:"password"`
		FirstName        string    `json:"first_name"`
		LastName         string    `json:"last_name"`
		Item             item      `json:"item"`
		StartDate        time.Time `json:"start_date"`
		EndDate          time.Time `json:"end_date"`
	}{}

	// unmarshal converts the input into json which is defined by the 2nd param (here it is dto)
	err := json.Unmarshal([]byte(args[0]), &dto)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Create new user if necessary
	var u user
	requestUserCreate := len(dto.Username) > 0 && len(dto.Password) > 0
	// composite key allows to create a key with mutiple values which can later be used to store key value data
	userKey, err := stub.CreateCompositeKey(prefixUser, []string{dto.Username})
	if requestUserCreate {
		// Check if a user with the same username exists
		if err != nil {
			return shim.Error(err.Error())
		}
		userAsBytes, _ := stub.GetState(userKey)
		if userAsBytes == nil {
			// Create new user
			u = user{
				Username:  dto.Username,
				Password:  dto.Password,
				FirstName: dto.FirstName,
				LastName:  dto.LastName,
			}
			// Persist the new user
			// convert the data inot byte array (JSON encoding)
			userAsBytes, err := json.Marshal(u)
			if err != nil {
				return shim.Error(err.Error())
			}
			// PutState puts the specified `key` and `value` into the transaction
			err = stub.PutState(userKey, userAsBytes)
			if err != nil {
				return shim.Error(err.Error())
			}
		} else {
			// Parse the existing user
			err := json.Unmarshal(userAsBytes, &u)
			if err != nil {
				return shim.Error(err.Error())
			}
		}
	} else {
		// Validate if the user with the provided username exists
		userAsBytes, _ := stub.GetState(userKey)
		if userAsBytes == nil {
			return shim.Error("User with this username does not exist.")
		}
	}

	// create contract struct to store useful info
	contract := Contract{
		Username:         dto.Username,
		ContractTypeUUID: dto.ContractTypeUUID,
		Item:             dto.Item,
		StartDate:        dto.StartDate,
		EndDate:          dto.EndDate,
		Void:             false,
		ClaimIndex:       []string{},
	}

	// composite key allows to create a key with muktiple values which can later be used to store key value data
	contractKey, err := stub.CreateCompositeKey(prefixContract, []string{dto.Username, dto.UUID})
	if err != nil {
		return shim.Error(err.Error())
	}
	// convert the data inot byte array (JSON encoding)
	contractAsBytes, err := json.Marshal(contract)
	if err != nil {
		return shim.Error(err.Error())
	}
	// PutState puts the specified `key` and `value` into the transaction
	err = stub.PutState(contractKey, contractAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Return success, if the new user has been created
	// (the user variable "u" should be blank)
	if !requestUserCreate {
		return shim.Success(nil)
	}

	// declaraton + defining values
	response := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: u.Username,
		Password: u.Password,
	}
	responseAsBytes, err := json.Marshal(response)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(responseAsBytes)
}

func createUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// need to pass 1 arg
	if len(args) != 1 {
		return shim.Error("Invalid argument count.")
	}

	user := user{}
	// unmarshal converts the input into json which is defined by the 2nd param (here it is user)
	err := json.Unmarshal([]byte(args[0]), &user)
	if err != nil {
		return shim.Error(err.Error())
	}
	// composite key allows to create a key with muktiple values which can later be used to store key value data
	key, err := stub.CreateCompositeKey(prefixUser, []string{user.Username})
	if err != nil {
		return shim.Error(err.Error())
	}

	// Check if the user already exists
	// GetState returns the value of the specified `key` from the ledger
	userAsBytes, _ := stub.GetState(key)
	// User does not exist, attempting creation
	if len(userAsBytes) == 0 {
		// convert the data inot byte array (JSON encoding)
		userAsBytes, err = json.Marshal(user)
		if err != nil {
			return shim.Error(err.Error())
		}
		// PutState puts the specified `key` and `value` into the transaction
		err = stub.PutState(key, userAsBytes)
		if err != nil {
			return shim.Error(err.Error())
		}

		// Return nil, if user is newly created
		return shim.Success(nil)
	}

	// unmarshal converts the input into json which is defined by the 2nd param (here it is user)
	err = json.Unmarshal(userAsBytes, &user)
	if err != nil {
		return shim.Error(err.Error())
	}

	userResponse := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: user.Username,
		Password: user.Password,
	}

	userResponseAsBytes, err := json.Marshal(userResponse)
	if err != nil {
		return shim.Error(err.Error())
	}
	// Return the username and the password of the already existing user
	return shim.Success(userResponseAsBytes)
}
