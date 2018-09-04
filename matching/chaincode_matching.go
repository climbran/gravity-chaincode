package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const PRE_KEY = "mc_"

type MatchingChaincode struct {
}

func (t *MatchingChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("MatchingChaincode Init")
	return shim.Success([]byte("success init"))
}

func (t *MatchingChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "matchingList" {
		return t.matchingList(stub)
	} else if function == "getAddr" {
		if len(args) != 1 {
			return shim.Error("Incorrect num of args, excepting 1")
		}
		return t.getAddr(stub, args[0])
	} else if function == "signup" {
		if len(args) != 2 {
			return shim.Error("Incorrect num of args, excepting 2")
		}
		return t.signup(stub, args[0], args[1])
	}
	return shim.Success([]byte("error fuction"))
}

func (t *MatchingChaincode) signup(stub shim.ChaincodeStubInterface, mc_id, mc_name string) pb.Response {
	var mc_key, _ = stub.CreateCompositeKey(PRE_KEY, []string{mc_id})

	err := stub.PutState(mc_key, []byte(mc_name))
	if err != nil {
		return shim.Error("写入数据失败")
	}
	return shim.Success([]byte("ok"))
}

func (t *MatchingChaincode) matchingList(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("start get matchingList")

	rs, err := stub.GetStateByPartialCompositeKey(PRE_KEY, []string{})

	if err != nil {
		return shim.Error("查询异常")
	}
	fmt.Println("query success")
	var mc_list []string
	defer rs.Close()
	var i int

	for i = 0; rs.HasNext(); i++ {
		responseRange, err := rs.Next()
		if err != nil {
			error_str := fmt.Sprintf("find error: %s", err)
			fmt.Println(error_str)
			return shim.Error(error_str)
		}
		mc_list = append(mc_list, responseRange.Key)
	}
	fmt.Println("start to json")

	json_mcs, err := json.Marshal(mc_list)
	if err == nil {
		fmt.Printf("%s\n", json_mcs)
	}
	fmt.Println("to json success")

	return shim.Success(json_mcs)
}

func (t *MatchingChaincode) getAddr(stub shim.ChaincodeStubInterface, mc_id string) pb.Response {

	addr, err := stub.GetState(mc_id)
	if err != nil {
		return shim.Error("读取数据失败")
	}
	return shim.Success(addr)
}

func main() {
	err := shim.Start(new(MatchingChaincode))
	if err != nil {
		fmt.Println("Could not start MatchingChaincode")
	} else {
		fmt.Println("MatchingChaincode successfully started")
	}
}
