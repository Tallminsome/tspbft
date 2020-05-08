package cmf

import (
	"errors"
	"flag"
	"fmt"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/message"
	"log"
	"os"
	"strconv"
	"strings"
)

type ShareConfig struct {
	//ClientServer bool
	Port           int
	Id             message.Identify
	Primary        message.Primary
	SubPrimary     message.SubPrimary
	Group          message.Group
	Table          map[message.Group][]string
	FaultNum       uint
	ThresholdNum   uint
	ExecuteMaxNum  int
	CheckPointNum  message.Sequence
	WaterLow       message.Sequence
	WaterHigh      message.Sequence
}

func LoadConfig() *ShareConfig {
	port,_ := GetConfigPort()
	id,_   := GetConfigID()
	pri,_  := GetConfigPrimary()
	sub,_  := GetConfigSubPrimary()
	grp,_  := GetConfigGroup()
	table,_:= GetConfigTable()

	//check if node is enough
	tablecheck := make(map[message.Group][]string)
	var fault  = 0
	for k, v := range table{
		tablecheck[message.Group(k)] = v
		fault = (len(tablecheck[message.Group(k)]) - 1) / 3
		if len(tablecheck[message.Group(k)]) % 3 != 1 {
			log.Fatalf("Error:Incorrect node number:%d,it should be 3f + 1",len(tablecheck))
			return nil
		}
	}

	flag.Parse()
	return &ShareConfig{
		Port:          port,
		Id:            message.Identify(id),
		Primary:       message.Primary(pri),
		SubPrimary:    message.SubPrimary(sub),
		Group:         message.Group(grp),
		Table:         tablecheck,
		FaultNum:      uint(fault),
		ThresholdNum:  uint(2*((len(tablecheck) - 1) / 3)+1),
		ExecuteMaxNum: 1,
		CheckPointNum: 20000,
		WaterLow:      0,
		WaterHigh:     40000,
	}
}
//
func GetConfigPort() (port int,err error)  {
	StrPort := os.Getenv("TSPBFT_LISTEN_PORT") //get env various
	if port, err = strconv.Atoi(StrPort);err != nil {
		return
	} //parse int
	return
}
// getconfigID
func GetConfigID() (id int,err error){
	StrID := os.Getenv("TSPBFT_NODE_ID")
	if id, err = strconv.Atoi(StrID);err != nil {
		return
	}
	return
}
// zhujiedian chushihua
func GetConfigPrimary() (bool,error)  {
	G := os.Getenv("TSPBFT_NODE_GROUP")
	if ( G == "A" ) {
		return true, nil
	}
	return false, nil
}

// xiacengzhujiedian chushihua
func GetConfigSubPrimary() (bool,error)  {
	G := os.Getenv("TSPBFT_NODE_GROUP")
	if ( G == "AB" || G == "AC" || G == "AD" ) {
		return true, nil
	}
	return false, nil
}
// de dao suo wei de tong dao
func GetConfigGroup() (string,error)  {
	G := os.Getenv("TSPBFT_NODE_GROUP")
	return G, nil
}

func GetConfigTable() (map[string][]string, error) {
	g := os.Getenv("TSPBFT_NODE_GROUP")
	var NodeTable = make(map[string][]string)
	switch g {
	case "A" :
		RawTable := os.Getenv("TSPBFT_NODE_TABLE_A")
		Tables  := strings.Split(RawTable,";")
		NodeTable["A"] = make([]string,len(Tables))
		for k,v := range Tables {
			NodeTable["A"][k] = v
		}
		if len(Tables)< 3 || len(Tables) % 3 != 1 {
			return nil, errors.New("Error GroupA Node Number,Node number should be 3f + 1")
		}
	case "AB" :
		RawTableA := os.Getenv("TSPBFT_NODE_TABLE_A")
		RawTableB := os.Getenv("TSPBFT_NODE_TABLE_B")

		TablesA  := strings.Split(RawTableA,";")
		TablesB  := strings.Split(RawTableB,";")

		NodeTable["A"] = make([]string,len(TablesA))
		NodeTable["B"] = make([]string,len(TablesB))

		for k,v := range TablesA {
			NodeTable["A"][k] = v
		}
		for k,v := range TablesB {
			NodeTable["B"][k] = v
		}
		if len(TablesA)< 3 || len(TablesA) % 3 != 1 {
			return nil, errors.New("Error GroupA Node Number,Node number should be 3f + 1")
		}
		if len(TablesB)< 3 || len(TablesB) % 3 != 1 {
			return nil, errors.New("Error GroupB Node Number,Node number should be 3f + 1")
		}
	case "AC" :
		RawTableA := os.Getenv("TSPBFT_NODE_TABLE_A")
		RawTableC := os.Getenv("TSPBFT_NODE_TABLE_C")

		TablesA  := strings.Split(RawTableA,";")
		TablesC  := strings.Split(RawTableC,";")

		NodeTable["A"] = make([]string,len(TablesA))
		NodeTable["C"] = make([]string,len(TablesC))

		for k,v := range TablesA {
			NodeTable["A"][k] = v
		}
		for k,v := range TablesC {
			NodeTable["C"][k] = v
		}
		if len(TablesA)< 3 || len(TablesA) % 3 != 1 {
			return nil, errors.New("Error GroupA Node Number,Node number should be 3f + 1")
		}
		if len(TablesC)< 3 || len(TablesC) % 3 != 1 {
			return nil, errors.New("Error GroupC Node Number,Node number should be 3f + 1")
		}
	case "AD" :
		RawTableA := os.Getenv("TSPBFT_NODE_TABLE_A")
		RawTableD := os.Getenv("TSPBFT_NODE_TABLE_D")

		TablesA  := strings.Split(RawTableA,";")
		TablesD  := strings.Split(RawTableD,";")

		NodeTable["A"] = make([]string,len(TablesA))
		NodeTable["D"] = make([]string,len(TablesD))

		for k,v := range TablesA {
			NodeTable["A"][k] = v
		}
		for k,v := range TablesD {
			NodeTable["D"][k] = v
		}
		fmt.Println("Check AD")
		if len(TablesA)< 3 || len(TablesA) % 3 != 1 {
			return nil, errors.New("Error GroupA Node Number,Node number should be 3f + 1")
		}
		if len(TablesD)< 3 || len(TablesD) % 3 != 1 {
			return nil, errors.New("Error GroupD Node Number,Node number should be 3f + 1")
		}
	case "B" :
		RawTable := os.Getenv("TSPBFT_NODE_TABLE_B")
		Tables  := strings.Split(RawTable,";")

		NodeTable["B"] = make([]string,len(Tables))

		for k,v := range Tables {
			NodeTable["B"][k] = v
		}
		if len(Tables)< 3 || len(Tables) % 3 != 1 {
			return nil, errors.New("Error GroupB Node Number,Node number should be 3f + 1")
		}
	case "C" :
		RawTable := os.Getenv("TSPBFT_NODE_TABLE_C")
		Tables  := strings.Split(RawTable,";")

		NodeTable["C"] = make([]string,len(Tables))

		for k,v := range Tables {
			NodeTable["C"][k] = v
		}
		if len(Tables)< 3 || len(Tables) % 3 != 1 {
			return nil, errors.New("Error GroupC Node Number,Node number should be 3f + 1")
		}
	case "D" :
		RawTable := os.Getenv("TSPBFT_NODE_TABLE_D")
		Tables  := strings.Split(RawTable,";")

		NodeTable["D"] = make([]string,len(Tables))

		for k,v := range Tables {
			NodeTable["D"][k] = v
		}
		if len(Tables)< 3 || len(Tables) % 3 != 1 {
			return nil, errors.New("Error GroupD Node Number,Node number should be 3f + 1")
		}
	default:
		log.Printf("Error in TSPBFT_NODE_GROUP")
	}
	return NodeTable,nil
}