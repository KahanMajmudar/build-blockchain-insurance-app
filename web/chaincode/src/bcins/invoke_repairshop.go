package main

import (
	"encoding/json"

	"github.com/hyperledger/fabric/core/chaincode/shim"

	pb "github.com/hyperledger/fabric/protos/peer"
)

func listRepairOrders(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// GetStateByPartialCompositeKey queries the state in the ledger based on
	// a given partial composite key. This function returns an iterator
	// which can be used to iterate over all composite keys whose prefix(here it is prefixRepairOrder)
	// matches the given partial composite key
	resultsIterator, err := stub.GetStateByPartialCompositeKey(prefixRepairOrder, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	// delay the execution of this line till nearby functions returns
	// usually used for cleanup work
	defer resultsIterator.Close()

	// array of interfaces
	results := []interface{}{}
	for resultsIterator.HasNext() {
		kvResult, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		repairOrder := repairOrder{}
		// unmarshal converts the input into json which is defined by the 2nd param (here it is repairOrder)
		err = json.Unmarshal(kvResult.Value, &repairOrder)
		if err != nil {
			return shim.Error(err.Error())
		}
		if repairOrder.Ready {
			continue
		}

		// create a result struct
		result := struct {
			UUID         string `json:"uuid"`
			ClaimUUID    string `json:"claim_uuid"`
			ContractUUID string `json:"contract_uuid"`
			Item         item   `json:"item"`
		}{}
		// unmarshal converts the input into json which is defined by the 2nd param (here it is result)
		err = json.Unmarshal(kvResult.Value, &result)
		if err != nil {
			return shim.Error(err.Error())
		}
		// SplitCompositeKey splits the specified key into attributes on which the
		// composite key was formed.
		prefix, keyParts, err := stub.SplitCompositeKey(kvResult.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		if len(keyParts) == 0 {
			result.UUID = prefix
		} else {
			result.UUID = keyParts[0]
		}
		// push the remaining things into result
		results = append(results, result)
	}

	resultsAsBytes, err := json.Marshal(results)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(resultsAsBytes)
}

func completeRepairOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Invalid argument count.")
	}

	input := struct {
		UUID string `json:"uuid"`
	}{}
	// unmarshal converts the input into json which is defined by the 2nd param (here it is input)
	err := json.Unmarshal([]byte(args[0]), &input)
	if err != nil {
		return shim.Error(err.Error())
	}

	// composite key allows to create a key with muktiple values which can later be used to store key value data
	repairOrderKey, err := stub.CreateCompositeKey(prefixRepairOrder, []string{input.UUID})
	if err != nil {
		return shim.Error(err.Error())
	}

	// GetState returns the value of the specified `key` from the ledger
	repairOrderBytes, _ := stub.GetState(repairOrderKey)
	if len(repairOrderBytes) == 0 {
		return shim.Error("Could not find the repair order")
	}

	repairOrder := repairOrder{}
	// unmarshal converts the input into json which is defined by the 2nd param (here it is repairOrder)
	err = json.Unmarshal(repairOrderBytes, &repairOrder)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Marking repair order as ready
	repairOrder.Ready = true

	// convert the data inot byte array (JSON encoding)
	repairOrderBytes, err = json.Marshal(repairOrder)
	if err != nil {
		return shim.Error(err.Error())
	}
	// PutState puts the specified `key` and `value` into the transaction
	err = stub.PutState(repairOrderKey, repairOrderBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Reflect changes in the corresponding claim
	// composite key allows to create a key with muktiple values which can later be used to store key value data
	claimKey, err := stub.CreateCompositeKey(prefixClaim, []string{repairOrder.ContractUUID, repairOrder.ClaimUUID})
	if err != nil {
		return shim.Error(err.Error())
	}
	// GetState returns the value of the specified `key` from the ledger
	claimBytes, _ := stub.GetState(claimKey)
	if claimBytes != nil {
		claim := Claim{}
		// unmarshal converts the input into json which is defined by the 2nd param (here it is claim)
		err := json.Unmarshal(claimBytes, &claim)
		if err != nil {
			return shim.Error(err.Error())
		}

		claim.Repaired = true
		// convert the data inot byte array (JSON encoding)
		claimBytes, err = json.Marshal(claim)
		if err != nil {
			return shim.Error(err.Error())
		}

		// PutState puts the specified `key` and `value` into the transaction
		err = stub.PutState(claimKey, claimBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	return shim.Success(nil)
}
