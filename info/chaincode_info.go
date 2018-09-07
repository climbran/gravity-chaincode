package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"
)

const INIT_COIN = 100
const PRE_KEY = "info_"

//peer chaincode invoke -C mychannel -n info -c '{"Function":"set","Args":["pubkey1","{\"Title\":\"banjia\",\"Content\":\"上门搬家服务\",\"Price\":10,\"City\":\"Beijing\"}","sign"]}'
// type Info struct {
// 	PubKey      string
// 	Title       string
// 	Content     string
// 	CompanyName string
// 	City        string
// 	Price       int
// 	PublishTime time.Time
// }

type InfoChaincode struct {
}

func (t *InfoChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("InfoChaincode Init")
	return shim.Success([]byte("success init"))
}

func (t *InfoChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
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
	} else if function == "getByOwner" {
		if len(args) != 1 {
			return shim.Error("Incorrect num of args, excepting 1")
		}
		return t.getByOwner(stub, args[0])
	} else if function == "matching" {
		if len(args) != 4 {
			return shim.Error("Incorrect num of args, excepting 4")
		}
		return t.matching(stub, args[0], args[1], args[2], args[3])
	}
	error_str := fmt.Sprintf("function error: %s\n", function)
	return shim.Error(error_str)
}

func (t *InfoChaincode) set(stub shim.ChaincodeStubInterface, pubKey string, info_str string, sign string) pb.Response {
	//签名信息校验
	if !Verify(pubKey, info_str, sign) {
		return shim.Error("签名验证失败")
	}
	checkResponse := stub.InvokeChaincode("check_info", [][]byte{[]byte("check"), []byte(pubKey), []byte(info_str)}, "")
	//用户信息校验
	if checkResponse.GetStatus() != shim.OK {
		return checkResponse
	}

	var info_key, _ = stub.CreateCompositeKey(PRE_KEY, []string{pubKey, stub.GetTxID()})

	err := stub.PutState(info_key, checkResponse.GetPayload())
	if err != nil {
		return shim.Error("写入数据失败")
	}
	return shim.Success([]byte("ok"))
}

func (t *InfoChaincode) get(stub shim.ChaincodeStubInterface, key string) pb.Response {
	info_str, err := stub.GetState(key)
	if err != nil {
		return shim.Error("系统异常")
	}
	return shim.Success(info_str)
}

func (t *InfoChaincode) getByOwner(stub shim.ChaincodeStubInterface, pubKey string) pb.Response {
	rs, err := stub.GetStateByPartialCompositeKey(PRE_KEY, []string{pubKey})
	if err != nil {
		return shim.Error("系统异常")
	}
	var info_map = make(map[string]string)
	defer rs.Close()

	for rs.HasNext() {
		responseRange, err := rs.Next()

		if err != nil {
			error_str := fmt.Sprintf("find error: %s", err)
			fmt.Println(error_str)
			return shim.Error(error_str)
		}
		info_map[responseRange.Key] = string(responseRange.Value)
	}
	json_infos, err := json.Marshal(info_map)
	if err == nil {
		fmt.Printf("%s\n", json_infos)
	}
	return shim.Success(json_infos)
}

func (t *InfoChaincode) matching(stub shim.ChaincodeStubInterface, mc_id, city, price_lower, price_upper string) pb.Response {

	lower, err := strconv.Atoi(price_lower)
	upper, err := strconv.Atoi(price_upper)
	if lower > upper || lower < 0 {
		return shim.Error(fmt.Sprintf("价格输入有误 %s, %s", lower, upper))
	}

	rs, err := stub.GetStateByPartialCompositeKey(PRE_KEY, []string{})
	if err != nil {
		return shim.Error("系统异常")
	}
	var info_map = make(map[string]string)
	defer rs.Close()

	for rs.HasNext() {
		responseRange, err := rs.Next()

		if err != nil {
			error_str := fmt.Sprintf("find error: %s", err)
			fmt.Println(error_str)
			return shim.Error(error_str)
		}
		info_map[responseRange.Key] = string(responseRange.Value)
	}
	json_infos, err := json.Marshal(info_map)
	if err == nil {
		fmt.Printf("%s\n", json_infos)
	}
	//根据匹配合约ID获取匹配合约地址
	mcRs := stub.InvokeChaincode("matching", [][]byte{[]byte("getAddr"), []byte(mc_id)}, stub.GetChannelID())
	if mcRs.GetStatus() != shim.OK {
		error_str := fmt.Sprintf("获取地址异常: %s", string(mcRs.Payload))
		fmt.Println(error_str)
		return shim.Error(error_str)
	}
	mc_addr := mcRs.GetPayload()

	//调用匹配合约
	checkResponse := stub.InvokeChaincode(string(mc_addr), [][]byte{[]byte("matching"), json_infos, []byte(city), []byte(price_lower), []byte(price_upper)}, stub.GetChannelID())

	return checkResponse
}

func main() {
	err := shim.Start(new(InfoChaincode))
	if err != nil {
		fmt.Println("Could not start InfoChaincode")
	} else {
		fmt.Println("InfoChaincode successfully started")
	}

}
