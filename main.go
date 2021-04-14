package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

var (
	cc          = "wkx-demo9"
	user        = "Admin" //此处Admin，但实际中应使用User1
	secret      = ""
	channelName = "mychannel"
	lvl         = logging.INFO
	orgName     = "Org1"
)

var (
	client *channel.Client
	sdk    *fabsdk.FabricSDK
	err    error
)

func init() {
	c := config.FromFile("./fhe-demo.yaml")
	sdk, err = fabsdk.New(c)
	if err != nil {
		fmt.Printf("Failed to create new SDK: %s\n", err)
		os.Exit(1)
	}

	clientChannelContext := sdk.ChannelContext(channelName, fabsdk.WithUser(user), fabsdk.WithOrg(orgName))
	if err != nil {
		fmt.Printf("Failed to create channel [%s] client: %#v", channelName, err)
		os.Exit(1)
	}

	client, err = channel.New(clientChannelContext)
	if err != nil {
		fmt.Printf("Failed to create channel [%s]:%s", channelName, err)
	}
}

func main() {
	enroll()

	sdk.Close()
}

func enroll() {
	content, err := ioutil.ReadFile("./pk.key")
	if err != nil {
		fmt.Println(err)
	}

	var queryArgs = [][]byte{content}
	response, err := client.Query(channel.Request{
		ChaincodeID: cc,
		Fcn:         "enroll",
		Args:        queryArgs,
	})

	fmt.Println(response.ChaincodeStatus, string(response.Payload))
}
