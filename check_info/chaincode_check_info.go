package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"time"
)

//test case
//["check","aaa","{\"Title\":\"bbb\",\"Content\":\"ccc\",\"Price\":100,\"City\":\"北京\"}"]
//peer chaincode query -C mychannel -n check_info -c '{"Function":"check","Args":["pubKey","{\"Title\":\"bbb\",\"Content\":\"ccc\",\"Price\":100,\"City\":\"Beijin\"}"]}'
//peer chaincode query -C mychannel -n check_info -c '{"Function":"check","Args":["pubkey1","{\"Title\":\"banjia\",\"Content\":\"上门搬家服务\",\"Price\":10,\"City\":\"Beijing\"}"]}'

type Info struct {
	PubKey      string
	Title       string
	Content     string
	CompanyName string
	City        string
	Price       int
	PublishTime time.Time
}
type User struct {
	Nickname string
	Name     string
	Age      string
	Phonenum string
	ID       string

	CompanyID   string
	CompanyName string
}

type InfoCheckChaincode struct {
}

func (t *InfoCheckChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("InfoCheckChaincode Init")
	return shim.Success([]byte("success init"))
}

func (t *InfoCheckChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "check" {
		return t.check(stub, args[0], args[1])
	}
	return shim.Success([]byte("error fuction"))
}

func (t *InfoCheckChaincode) check(stub shim.ChaincodeStubInterface, pubKey string, info_str string) pb.Response {
	i, err := jsonToInfo(info_str)
	if err != nil {
		error_str := fmt.Sprintf("信息转换失败: %s\n %s", err, info_str)
		fmt.Println(error_str)
		return shim.Error(error_str)
	}
	if len(i.Title) == 0 || len(i.Content) == 0 || len(i.City) == 0 {
		return shim.Error("Title,Content,City必须填写")
	}
	if i.Price < 0 {
		return shim.Error("价格不能小于0")
	}
	userResponse := stub.InvokeChaincode("user", [][]byte{[]byte("get"), []byte(pubKey)}, "")
	if userResponse.GetStatus() != shim.OK {
		return userResponse
	}
	user_str := userResponse.GetPayload()

	if len(user_str) <= 0 {
		return shim.Error("用户不存在")
	}
	u, err := jsonToUser(string(user_str))
	if err != nil {
		return shim.Error("用户数据异常")
	}
	if len(u.CompanyName) == 0 {
		return shim.Error("非商家账号不能发布信息")
	}

	i.CompanyName = u.CompanyName
	i.PubKey = pubKey

	tm, err := stub.GetTxTimestamp()
	i.PublishTime = time.Unix(tm.Seconds, 0)
	return shim.Success(i.toString())
}

func (u *Info) toString() []byte {
	if data, err := json.Marshal(u); err == nil {
		return data
	}
	return []byte("err")
}

func jsonToInfo(str string) (Info, error) {
	var i Info
	err := json.Unmarshal([]byte(str), &i)
	if err == nil {
		return i, nil
	}
	return i, err
}
func jsonToUser(str string) (User, error) {
	var u User
	err := json.Unmarshal([]byte(str), &u)
	if err == nil {
		return u, nil
	}
	return u, err
}

func main() {
	err := shim.Start(new(InfoCheckChaincode))
	if err != nil {
		fmt.Println("Could not start InfoCheckChaincode")
	} else {
		fmt.Println("InfoCheckChaincode successfully started")
	}
	// u := &User{"Yan", 18}
	// fmt.Printf("%s\n", u.toString())
}
