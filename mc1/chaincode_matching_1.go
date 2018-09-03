package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"sort"
	"strconv"
	"strings"
	"time"
)

const PRE_KEY = "mc_"

//test case
//{"Args":["mc","{\"kkkk\":{\"PublishTime\":\"2018-08-27T12:31:47Z\",\"City\":\"Shanghai\",\"Price\":100},\"nnnnn\":{\"PublishTime\":\"2019-06-27T12:31:47Z\",\"City\":\"Beijing\",\"Price\":300}}","Beijing","50","300"]}

type Info struct {
	ID          string
	PubKey      string
	Title       string
	Content     string
	CompanyName string
	City        string
	Price       int
	PublishTime time.Time
}

type Infos []Info

func (infos Infos) Len() int {
	return len(infos)
}

//根据时间排序
func (infos Infos) Less(i, j int) bool {
	return infos[i].PublishTime.After(infos[j].PublishTime)
}
func (infos Infos) Swap(i, j int) {
	infos[i], infos[j] = infos[j], infos[i]
}

func (i *Info) toString() []byte {
	if data, err := json.Marshal(i); err == nil {
		return data
	}
	return []byte("err")
}

type D1_MatchingChaincode struct {
}

func (t *D1_MatchingChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	//目前不支持跨链码交易，需手动注册
	//checkResponse := stub.InvokeChaincode("matching", [][]byte{[]byte("signup"), []byte("1"), []byte("matching_1")}, "")

	return shim.Success([]byte("ok"))
}

func (t *D1_MatchingChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	_, args := stub.GetFunctionAndParameters()
	infos_str := args[0]
	city := args[1]
	price_lower, err := strconv.Atoi(args[2])
	price_upper, err := strconv.Atoi(args[3])
	if price_lower > price_upper || price_lower < 0 {
		return shim.Error(fmt.Sprintf("价格输入有误 %s, %s", price_lower, price_upper))
	}
	if err != nil {
		return shim.Error("price参数转换为int类型异常")
	}
	return t.matching(stub, infos_str, city, price_lower, price_upper)
}

func (t *D1_MatchingChaincode) matching(stub shim.ChaincodeStubInterface, infos_str, city string, price_lower, price_upper int) pb.Response {
	var infos Infos
	var info_map = make(map[string]Info)

	err := json.Unmarshal([]byte(infos_str), &info_map)
	if err != nil {
		error_str := fmt.Sprintf("string to json error: %s", err)
		fmt.Println(error_str)
		return shim.Error(error_str)
	}

	for k, v := range info_map {

		if strings.EqualFold(v.City, city) && v.Price >= price_lower && v.Price <= price_upper {
			v.ID = k
			infos = append(infos, v)
		}
	}
	sort.Sort(infos)

	var info_sortMap = make(map[string]string)
	for index, i := range infos {
		info_sortMap[strconv.Itoa(index)] = string(i.toString())
	}
	result_str, err := json.Marshal(info_sortMap)
	if err != nil {
		error_str := fmt.Sprintf("map to string error: %s", err)
		fmt.Println(error_str)
		return shim.Error(error_str)
	}
	return shim.Success(result_str)
}

func main() {
	err := shim.Start(new(D1_MatchingChaincode))
	if err != nil {
		fmt.Println("Could not start D1_MatchingChaincode")
	} else {
		fmt.Println("D1_MatchingChaincode successfully started")
	}
	// u := &User{"Yan", 18}
	// fmt.Printf("%s\n", u.toString())
}
