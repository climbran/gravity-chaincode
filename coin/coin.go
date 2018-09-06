package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"
)

const SUFFIX_COIN = "_coin"
const SUFFIX_FREEZE = "_coin_freeze"

//test case
//peer chaincode query -C mychannel -n coin -c '{"Args":["get","Yanweiqing"]}'

type CoinChaincode struct {
}

func (t *CoinChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	return shim.Success([]byte("ok"))
}

func (t *CoinChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "issue" {
		if len(args) != 2 {
			return shim.Error("Incorrect num of args, excepting 2")
		}
		amount, err := strconv.Atoi(string(args[1]))
		if err != nil {
			return shim.Error("参数转换为int类型异常")
		}
		return t.issue(stub, args[0], amount)
	} else if function == "freeze" {
		if len(args) != 2 {
			return shim.Error("Incorrect num of args, excepting 2")
		}
		amount, err := strconv.Atoi(string(args[1]))
		if err != nil {
			return shim.Error("参数转换为int类型异常")
		}
		return t.freeze(stub, args[0], amount)
	} else if function == "confirm" {
		if len(args) != 3 {
			return shim.Error("Incorrect num of args, excepting 3")
		}
		amount, err := strconv.Atoi(string(args[2]))
		if err != nil {
			return shim.Error("参数转换为int类型异常")
		}
		return t.confirm(stub, args[0], args[1], amount)
	} else if function == "get" {
		if len(args) != 1 {
			return shim.Error("Incorrect num of args, excepting 1")
		}
		return t.get(stub, args[0])
	}
	return shim.Error("function error")
}

func (t *CoinChaincode) issue(stub shim.ChaincodeStubInterface, pubKey string, amount int) pb.Response {
	err := stub.PutState(pubKey+SUFFIX_COIN, []byte(strconv.Itoa(amount)))

	if err != nil {
		return shim.Error("issue coin fail")
	}
	return shim.Success([]byte("ok"))
}

func (t *CoinChaincode) get(stub shim.ChaincodeStubInterface, pubKey string) pb.Response {
	b, err := stub.GetState(pubKey + SUFFIX_COIN)

	if err != nil {
		return shim.Error("get coin fail")
	}
	return shim.Success(b)
}

func (t *CoinChaincode) freeze(stub shim.ChaincodeStubInterface, pubKey string, amount int) pb.Response {
	//校验用户是否存在
	checkRs := checkUser(stub, pubKey)
	if checkRs.GetStatus() != shim.OK {
		return shim.Error("用户不存在")
	}
	//获取用户余额
	b, err := stub.GetState(pubKey + SUFFIX_COIN)
	balance, err := strconv.Atoi(string(b))
	if err != nil {
		return shim.Error("参数转换为int类型异常")
	}
	if balance < amount {
		return shim.Error("balance not enough")
	}
	fz, err := stub.GetState(pubKey + SUFFIX_FREEZE)
	if err != nil {
		error_str := fmt.Sprintf("get data error: %s\n", err)
		fmt.Println(error_str)
		return shim.Error(error_str)
	}
	freeze := 0
	if len(fz) > 0 {
		freeze, err = strconv.Atoi(string(fz))
		if err != nil {
			error_str := fmt.Sprintf("string to int error: %s\n, %s", err, fz)
			fmt.Println(error_str)
			return shim.Error(error_str)
		}
	}

	//计算冻结后余额
	newBalance := balance - amount
	newFreeze := freeze + amount
	//更新数据
	err = stub.PutState(pubKey+SUFFIX_COIN, []byte(strconv.Itoa(newBalance)))
	err = stub.PutState(pubKey+SUFFIX_FREEZE, []byte(strconv.Itoa(newFreeze)))
	if err != nil {
		return shim.Error("put fail")
	}
	return shim.Success([]byte("ok"))
}

func (t *CoinChaincode) confirm(stub shim.ChaincodeStubInterface, from, to string, amount int) pb.Response {
	//校验用户是否存在
	checkFrom := checkUser(stub, from)
	if checkFrom.GetStatus() != shim.OK {
		return shim.Error("from不存在")
	}
	checkTo := checkUser(stub, to)
	if checkTo.GetStatus() != shim.OK {
		return shim.Error("to不存在")
	}
	//获取用户余额
	f_fz, err := stub.GetState(from + SUFFIX_FREEZE)
	from_freeze, err := strconv.Atoi(string(f_fz))
	if err != nil {
		return shim.Error("参数转换为int类型异常")
	}
	if from_freeze < amount {
		return shim.Error("from freeze not enough")
	}
	t_b, err := stub.GetState(to + SUFFIX_COIN)
	to_balance, err := strconv.Atoi(string(t_b))
	if err != nil {
		return shim.Error("get fail")
	}
	//计算冻结后余额
	new_fz := from_freeze - amount
	new_tb := to_balance + amount
	//更新数据
	err = stub.PutState(from+SUFFIX_FREEZE, []byte(strconv.Itoa(new_fz)))
	err = stub.PutState(to+SUFFIX_COIN, []byte(strconv.Itoa(new_tb)))
	if err != nil {
		return shim.Error("put fail")
	}
	return shim.Success([]byte("ok"))
}

func checkUser(stub shim.ChaincodeStubInterface, pubKey string) pb.Response {
	userResponse := stub.InvokeChaincode("user", [][]byte{[]byte("get"), []byte(pubKey)}, "")
	if userResponse.GetStatus() != shim.OK {
		return userResponse
	}
	user_str := userResponse.GetPayload()

	if len(user_str) > 0 {
		return shim.Success([]byte("ok"))
	} else {
		return shim.Error("用户不存在")
	}
}

func main() {
	err := shim.Start(new(CoinChaincode))
	if err != nil {
		fmt.Println("Could not start CoinChaincode")
	} else {
		fmt.Println("CoinChaincode successfully started")
	}
}
