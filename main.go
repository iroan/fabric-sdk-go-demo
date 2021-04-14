package main

import (
	"fmt"
	"os"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

var (
	cc          = "wkx-demo1"
	user        = "Admin" //此处Admin，但实际中应使用User1
	secret      = ""
	channelName = "mychannel"
	lvl         = logging.INFO
	orgName     = "Org1"
)

func main() {
	c := config.FromFile("./fhe-demo.yaml")
	sdk, err := fabsdk.New(c)
	if err != nil {
		fmt.Printf("Failed to create new SDK: %s\n", err)
		os.Exit(1)
	}
	defer sdk.Close()
	clientChannelContext := sdk.ChannelContext(channelName, fabsdk.WithUser(user), fabsdk.WithOrg(orgName))
	if err != nil {
		fmt.Printf("Failed to create channel [%s] client: %#v", channelName, err)
		os.Exit(1)
	}

	client, err := channel.New(clientChannelContext)
	if err != nil {
		fmt.Printf("Failed to create channel [%s]:%s", channelName, err)
	}
	queryCC(client, []byte("a"))
	invokeCC(client, "ff55")
	queryCC(client, []byte("a"))
}

func invokeCC(client *channel.Client, newValue string) {
	invokeArgs := [][]byte{[]byte("a"), []byte(newValue)}

	_, err := client.Execute(channel.Request{
		ChaincodeID: cc,
		Fcn:         "set",
		Args:        invokeArgs,
	})

	if err != nil {
		fmt.Printf("Failed to invoke: %+v\n", err)
	}
}

func queryCC(client *channel.Client, name []byte) string {
	var queryArgs = [][]byte{name}
	response, err := client.Query(channel.Request{
		ChaincodeID: cc,
		Fcn:         "creater",
		Args:        queryArgs,
	})

	if err != nil {
		fmt.Println("Failed to query: ", err)
	}

	ret := string(response.Payload)
	fmt.Println("Chaincode status: ", response.ChaincodeStatus)
	fmt.Println("Payload: ", ret)
	return ret
}
