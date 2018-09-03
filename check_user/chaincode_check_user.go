package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

//['check','{"name":"Yan","Age":"18","ID":"43070212345767","Nickname":"成本","Phonenum":"1376890876"}']
//'{"Function":"check","Args":["{\"Name\":\"Yan\",\"Age\":\"18\",\"ID\":\"43070212345767\",\"Nickname\":\"swf\",\"Phonenum\":\"1376890876\"}"]}'
type User struct {
	Nickname string
	Name     string
	Age      string
	Phonenum string
	ID       string

	CompanyID   string
	CompanyName string
}

type UserCheckChaincode struct {
}

func (t *UserCheckChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("UserCheckChaincode Init")
	return shim.Success([]byte("success init"))
}

func (t *UserCheckChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "check" {
		return t.check(stub, args[0])
	}
	return shim.Success([]byte("error fuction"))
}

func (t *UserCheckChaincode) check(stub shim.ChaincodeStubInterface, user_str string) pb.Response {
	u, err := jsonToUser(user_str)
	if err != nil {
		error_str := fmt.Sprintf("map to string error: %s\n %s", err, user_str)
		fmt.Println(error_str)
		return shim.Error(error_str)
	}
	if len(u.Name) == 0 || len(u.Age) == 0 || len(u.Phonenum) == 0 || len(u.ID) == 0 || len(u.Nickname) == 0 {
		return shim.Error("Name,Age,Phonenum,ID,Nickname必须填写")
	}
	if (len(u.CompanyID) == 0 && len(u.CompanyName) > 0) || (len(u.CompanyID) > 0 && len(u.CompanyName) == 0) {
		return shim.Error("商家信息不完整,CompanyID、CompanyName必须填写")
	}

	return shim.Success(u.toString())
}

func (u *User) toString() []byte {
	if data, err := json.Marshal(u); err == nil {
		return data
	}
	return []byte("err")
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
	err := shim.Start(new(UserCheckChaincode))
	if err != nil {
		fmt.Println("Could not start UserCheckChaincode")
	} else {
		fmt.Println("UserCheckChaincode successfully started")
	}
}
