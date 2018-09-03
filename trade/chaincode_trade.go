package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"
	"time"
)

const SUFFIX_COIN = "_coin"
const SUFFIX_FREEZE = "_coin_freeze"

//test case
//{"Args":["mc","{\"kkkk\":{\"PublishTime\":\"2018-08-27T12:31:47Z\",\"City\":\"Shanghai\",\"Price\":100},\"nnnnn\":{\"PublishTime\":\"2019-06-27T12:31:47Z\",\"City\":\"Beijing\",\"Price\":300}}","Beijing","50","300"]}

type Info struct {
	PubKey      string
	Title       string
	Content     string
	CompanyName string
	City        string
	Price       int
	PublishTime time.Time
}

type Trade struct {
	Constumer   string
	Business    string
	InfoID      string
	Title       string
	Price       int
	SubmitTime  time.Time
	ConfirmTime time.Time
	FinishTIme  time.Time
	State       int
}

const STATE_SUBMIT = 1
const STATE_CONFIRM = 2
const STATE_FINISH = 3
const PRE_KEY_C = "trade_c"
const PRE_KEY_B = "trade_c"

type TradeChaincode struct {
}

func (t *TradeChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	return shim.Success([]byte("ok"))
}

func (t *TradeChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "submit" {
		return t.submit(stub, args[0], args[1], args[2])
	} else if function == "confirm" {
		return t.confirm(stub, args[0], args[1], args[2])
	} else if function == "finish" {
		return t.finish(stub, args[0], args[1], args[2])
	} else if function == "getTradeByConstumer" {
		return t.getTradeByConstumer(stub, args[0])
	} else if function == "getTradeByBusiness" {
		return t.getTradeByBusiness(stub, args[0])
	}
	return shim.Error("function error")
}

func (t *TradeChaincode) submit(stub shim.ChaincodeStubInterface, pubKey, infoId, sign string) pb.Response {
	if !Verify(pubKey, infoId, sign) {
		return shim.Error("签名验证失败")
	}
	infoResponse := stub.InvokeChaincode("info", [][]byte{[]byte("get"), []byte(infoId)}, "")
	if infoResponse.GetStatus() != shim.OK {
		return infoResponse
	}
	info_str := infoResponse.GetPayload()
	if len(info_str) <= 0 {
		return shim.Error("info not find")
	}
	info, err := jsonToInfo(string(info_str))

	if err != nil {
		return shim.Error("json error")
	}
	coinRs := stub.InvokeChaincode("coin", [][]byte{[]byte("freeze"), []byte(pubKey), []byte(strconv.Itoa(info.Price))}, "")
	if coinRs.Status != shim.OK {
		return coinRs
	}
	trade := &Trade{}
	trade.Constumer = pubKey
	trade.InfoID = infoId
	trade.Title = info.Title
	trade.Business = info.PubKey
	tm, err := stub.GetTxTimestamp()
	trade.SubmitTime = time.Unix(tm.Seconds, 0)
	trade.State = STATE_SUBMIT
	trade.Price = info.Price

	var tradeID, _ = stub.CreateCompositeKey(PRE_KEY_C, []string{pubKey, stub.GetTxID()})
	var tradeID_B, _ = stub.CreateCompositeKey(PRE_KEY_B, []string{info.PubKey, stub.GetTxID()})

	err = stub.PutState(tradeID, trade.toString())
	err = stub.PutState(tradeID_B, trade.toString())
	if err != nil {
		return shim.Error("写入数据失败")
	}
	return shim.Success([]byte("ok"))
}

func (t *TradeChaincode) confirm(stub shim.ChaincodeStubInterface, pubKey, tradeID, sign string) pb.Response {
	if !Verify(pubKey, tradeID, sign) {
		return shim.Error("签名验证失败")
	}
	trade_str, err := stub.GetState(tradeID)
	if err != nil {
		return shim.Error("query error")
	}
	if len(trade_str) <= 0 {
		return shim.Error("trade not find")
	}
	trade, err := jsonToTrade(string(trade_str))
	if err != nil {
		return shim.Error("json error")
	}
	if trade.State != STATE_SUBMIT {
		return shim.Error("state not submit")
	}
	_, attrArray, _ := stub.SplitCompositeKey(tradeID)
	var tradeID_C, _ = stub.CreateCompositeKey(PRE_KEY_C, []string{trade.Constumer, attrArray[1]})

	trade.State = STATE_CONFIRM
	tm, err := stub.GetTxTimestamp()
	trade.ConfirmTime = time.Unix(tm.Seconds, 0)

	err = stub.PutState(tradeID, trade.toString())
	err = stub.PutState(tradeID_C, trade.toString())

	if err != nil {
		return shim.Error("写入数据失败")
	}

	return shim.Success([]byte("ok"))
}

func (t *TradeChaincode) finish(stub shim.ChaincodeStubInterface, pubKey, tradeID, sign string) pb.Response {
	if !Verify(pubKey, tradeID, sign) {
		return shim.Error("签名验证失败")
	}
	trade_str, err := stub.GetState(tradeID)
	if err != nil {
		return shim.Error("query error")
	}
	if len(trade_str) <= 0 {
		return shim.Error("trade not find")
	}
	trade, err := jsonToTrade(string(trade_str))
	if err != nil {
		return shim.Error("json error")
	}
	if trade.State != STATE_SUBMIT {
		return shim.Error("state not submit")
	}
	coinRs := stub.InvokeChaincode("coin", [][]byte{[]byte("confirm"), []byte(trade.Constumer), []byte(trade.Business), []byte(strconv.Itoa(trade.Price))}, "")
	if coinRs.Status != shim.OK {
		return coinRs
	}

	_, attrArray, _ := stub.SplitCompositeKey(tradeID)
	var tradeID_B, _ = stub.CreateCompositeKey(PRE_KEY_B, []string{trade.Business, attrArray[1]})

	trade.State = STATE_FINISH
	tm, err := stub.GetTxTimestamp()
	trade.FinishTIme = time.Unix(tm.Seconds, 0)

	err = stub.PutState(tradeID, trade.toString())
	err = stub.PutState(tradeID_B, trade.toString())

	if err != nil {
		return shim.Error("写入数据失败")
	}

	return shim.Success([]byte("ok"))
}

func (t *TradeChaincode) getTradeByConstumer(stub shim.ChaincodeStubInterface, pubKey string) pb.Response {
	rs, err := stub.GetStateByPartialCompositeKey(PRE_KEY_C, []string{pubKey})
	if err != nil {
		return shim.Error("系统异常")
	}
	var trade_map = make(map[string]string)
	defer rs.Close()

	for rs.HasNext() {
		responseRange, err := rs.Next()

		if err != nil {
			error_str := fmt.Sprintf("find error: %s", err)
			fmt.Println(error_str)
			return shim.Error(error_str)
		}
		trade_map[responseRange.Key] = string(responseRange.Value)
	}
	json_trades, err := json.Marshal(trade_map)
	if err == nil {
		fmt.Printf("%s\n", json_trades)
	}
	return shim.Success(json_trades)
}

func (t *TradeChaincode) getTradeByBusiness(stub shim.ChaincodeStubInterface, pubKey string) pb.Response {
	rs, err := stub.GetStateByPartialCompositeKey(PRE_KEY_C, []string{pubKey})
	if err != nil {
		return shim.Error("系统异常")
	}
	var trade_map = make(map[string]string)
	defer rs.Close()

	for rs.HasNext() {
		responseRange, err := rs.Next()

		if err != nil {
			error_str := fmt.Sprintf("find error: %s", err)
			fmt.Println(error_str)
			return shim.Error(error_str)
		}
		trade_map[responseRange.Key] = string(responseRange.Value)
	}
	json_trades, err := json.Marshal(trade_map)
	if err == nil {
		fmt.Printf("%s\n", json_trades)
	}
	return shim.Success(json_trades)
}

func jsonToInfo(str string) (Info, error) {
	var i Info
	err := json.Unmarshal([]byte(str), &i)
	if err == nil {
		return i, nil
	}
	return i, err
}

func jsonToTrade(str string) (Trade, error) {
	var t Trade
	err := json.Unmarshal([]byte(str), &t)
	if err == nil {
		return t, nil
	}
	return t, err
}

func (u *Trade) toString() []byte {
	if data, err := json.Marshal(u); err == nil {
		return data
	}
	return []byte("err")
}

func main() {
	err := shim.Start(new(TradeChaincode))
	if err != nil {
		fmt.Println("Could not start TradeChaincode")
	} else {
		fmt.Println("TradeChaincode successfully started")
	}
}
