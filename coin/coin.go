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
//{"Args":["mc","{\"kkkk\":{\"PublishTime\":\"2018-08-27T12:31:47Z\",\"City\":\"Shanghai\",\"Price\":100},\"nnnnn\":{\"PublishTime\":\"2019-06-27T12:31:47Z\",\"City\":\"Beijing\",\"Price\":300}}","Beijing","50","300"]}

type CoinChaincode struct {
}

func (t *CoinChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	return shim.Success([]byte("ok"))
}

func (t *CoinChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "issue" {
		amount, err := strconv.Atoi(string(args[1]))
		if err != nil {
			return shim.Error("参数转换为int类型异常")
		}
		return t.issue(stub, args[0], amount)
	} else if function == "freeze" {
		amount, err := strconv.Atoi(string(args[1]))
		if err != nil {
			return shim.Error("参数转换为int类型异常")
		}
		return t.freeze(stub, args[0], amount)
	} else if function == "confirm" {
		amount, err := strconv.Atoi(string(args[2]))
		if err != nil {
			return shim.Error("参数转换为int类型异常")
		}
		return t.confirm(stub, args[0], args[1], amount)
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
	freeze, err := strconv.Atoi(string(fz))
	if err != nil {
		return shim.Error("issue coin fail")
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
