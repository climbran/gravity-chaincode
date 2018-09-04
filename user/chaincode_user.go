package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const INIT_COIN = "100"

//peer chaincode invoke -C mychannel -n user -c '{"Function":"set","Args":["pubkey","{\"Name\":\"Yan\",\"Age\":\"18\",\"ID\":\"43070212345767\",\"Nickname\":\"swf\",\"Phonenum\":\"1376890876\"}","sign"]}'
//peer chaincode invoke -C mychannel -n user -c '{"Function":"get","Args":["pubkey"]}'

//peer chaincode invoke -C mychannel -n user -c '{"Function":"set","Args":["pubkey2","{\"Name\":\"Yan\",\"ID\":\"43070212345767\",\"Age\":\"18\",\"Nickname\":\"swf\",\"Phonenum\":\"1376890876\",\"CompanyName\":\"58Company\",\"CompanyID\":\"5858\"}","sign"]}'
type UserChaincode struct {
}

func (t *UserChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("UserChaincode Init")
	return shim.Success([]byte("success init"))
}

func (t *UserChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "get" {
		if len(args) != 1 {
			return shim.Error("Incorrect num of args, excepting 1")
		}
		return t.get(stub, args[0])
	} else if function == "set" {
		if len(args) != 3 {
			return shim.Error("Incorrect num of args, excepting 3")
		}
		return t.set(stub, args[0], args[1], args[2])
	}
	return shim.Success([]byte("error fuction"))
}

func (t *UserChaincode) set(stub shim.ChaincodeStubInterface, pubKey string, user_str string, sign string) pb.Response {
	//签名信息校验
	if !Verify(pubKey, user_str, sign) {
		return shim.Error("签名验证失败")
	}

	checkResponse := stub.InvokeChaincode("check_user_gr", [][]byte{[]byte("check"), []byte(user_str)}, "")
	//用户信息校验
	if checkResponse.GetStatus() != shim.OK {
		return checkResponse
	}

	old_user, err := stub.GetState(pubKey)
	if err != nil {
		return shim.Error("系统异常")
	}
	//第一次注册赠送100个币
	if old_user == nil || len(old_user) == 0 {
		issueResponse := stub.InvokeChaincode("coin", [][]byte{[]byte("issue"), []byte(pubKey), []byte(INIT_COIN)}, "")
		if issueResponse.GetStatus() != shim.OK {
			return issueResponse
		}
	}
	err = stub.PutState(pubKey, checkResponse.GetPayload())
	if err != nil {
		return shim.Error("写入数据失败")
	}
	return shim.Success([]byte("ok"))
}

func (t *UserChaincode) get(stub shim.ChaincodeStubInterface, pubKey string) pb.Response {
	user_str, err := stub.GetState(pubKey)
	if err != nil {
		return shim.Error("系统异常")
	}
	return shim.Success(user_str)
}

func main() {
	err := shim.Start(new(UserChaincode))
	if err != nil {
		fmt.Println("Could not start UserChaincode")
	} else {
		fmt.Println("UserChaincode successfully started")
	}
	// u := &User{"Yan", 18}
	// fmt.Printf("%s\n", u.toString())
}
