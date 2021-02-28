package main

// import all the necessary packages
import (
	"fmt"

	"encoding/json"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// This are all the constants used throught the chaincode (cc)
const prefixContractType = "contract_type"
const prefixContract = "contract"
const prefixClaim = "claim"
const prefixUser = "user"
const prefixRepairOrder = "repair_order"

var logger = shim.NewLogger("main")

type SmartContract struct {
}

// create a maping for functions, so we can call 'N' contract functions from a single contract
var bcFunctions = map[string]func(shim.ChaincodeStubInterface, []string) pb.Response{

	//This is the function defination, Function declaration is in invoke_insurance.go file
	// You can find all the below mentioned function there
	// Insurance Peer
	"contract_type_ls":         listContractTypes,
	"contract_type_create":     createContractType,
	"contract_type_set_active": setActiveContractType,
	"contract_ls":              listContracts,
	"claim_ls":                 listClaims,
	"claim_file":               fileClaim,
	"claim_process":            processClaim,
	"user_authenticate":        authUser,
	"user_get_info":            getUser,

	//This is the function defination, Function declaration is in invoke_shop.go file
	// You can find createContract() and createUser() function there
	// Shop Peer
	"contract_create": createContract,
	"user_create":     createUser,

	//This is the function defination, Function declaration is in invoke_repairshop.go file
	// You can find listRepairOrders() and completeRepairOrder() function there
	// Repair Shop Peer
	"repair_order_ls":       listRepairOrders,
	"repair_order_complete": completeRepairOrder,

	//This is the function defination, Function declaration is in invoke_police.go file
	// You can find listTheftClaims() and processTheftClaim() function there
	// Police Peer
	"theft_claim_ls":      listTheftClaims,
	"theft_claim_process": processTheftClaim,
}

// Init callback representing the invocation of a chaincode
// SmartContract is a receiver, tha means the Init method is defined on SmartContract and * means it is passed as pointer
func (t *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	// GetFunctionAndParameters returns the first argument as the function
	// name and the rest of the arguments as parameters in a string array.
	// _ means you are not going to use it but only using the args
	_, args := stub.GetFunctionAndParameters()

	if len(args) == 1 {
		// create a struct to store uuid and the supplied contract type
		var contractTypes []struct {
			UUID string `json:"uuid"`
			*ContractType
		}
		// unmarshal converts the input into json which is defined by the 2nd param (here it is contracTypes)
		err := json.Unmarshal([]byte(args[0]), &contractTypes)
		if err != nil {
			return shim.Error(err.Error())
		}
		// for multiple loop over all the types
		for _, ct := range contractTypes {
			// composite key allows to create a key with mutiple values which can later be used to store key value data
			contractTypeKey, err := stub.CreateCompositeKey(prefixContractType, []string{ct.UUID})
			if err != nil {
				return shim.Error(err.Error())
			}
			// convert the data inot byte array (JSON encoding)
			contractTypeAsBytes, err := json.Marshal(ct.ContractType)
			if err != nil {
				return shim.Error(err.Error())
			}
			// PutState puts the specified `key` and `value` into the transaction
			err = stub.PutState(contractTypeKey, contractTypeAsBytes)
			if err != nil {
				return shim.Error(err.Error())
			}
		}
	}
	return shim.Success(nil)
}

// Invoke Function accept blockchain code invocations.
func (t *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "init" {
		return t.Init(stub)
	}
	bcFunc := bcFunctions[function]
	if bcFunc == nil {
		return shim.Error("Invalid invoke function.")
	}
	return bcFunc(stub, args)
}

func main() {
	logger.SetLevel(shim.LogInfo)

	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
