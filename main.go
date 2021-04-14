package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/ldsec/lattigo/v2/bfv"
)

var (
	channelName = "mychannel"
	orgName     = "Org1"
	ccName      = "wkx-demo33"
)

var (
	clientAdmin *channel.Client
	clientUser1 *channel.Client
	sdk         *fabsdk.FabricSDK
	err         error
	params      *bfv.Parameters
)

func init() {
	c := config.FromFile("./fhe-demo.yaml")
	sdk, err = fabsdk.New(c)
	if err != nil {
		fmt.Printf("Failed to create new SDK: %s\n", err)
		os.Exit(1)
	}

	ccc1 := sdk.ChannelContext(channelName, fabsdk.WithUser("Admin"), fabsdk.WithOrg(orgName))
	if err != nil {
		fmt.Printf("Failed to create channel [%s] client: %#v", channelName, err)
		os.Exit(1)
	}

	clientAdmin, err = channel.New(ccc1)
	if err != nil {
		fmt.Printf("Failed to create channel [%s]:%s", channelName, err)
	}

	ccc2 := sdk.ChannelContext(channelName, fabsdk.WithUser("User1"), fabsdk.WithOrg(orgName))
	if err != nil {
		fmt.Printf("Failed to create channel [%s] client: %#v", channelName, err)
		os.Exit(1)
	}

	clientUser1, err = channel.New(ccc2)
	if err != nil {
		fmt.Printf("Failed to create channel [%s]:%s", channelName, err)
	}
	params = bfv.DefaultParams[bfv.PN13QP218].WithT(0x3ee0001)
}

func deserializePk() *bfv.PublicKey {
	content, err := ioutil.ReadFile("pk.key")
	if err != nil {
		log.Fatalln(err)
	}

	pk := bfv.NewPublicKey(params)
	if err != nil {
		log.Fatalln(err)
	}

	pk.UnmarshalBinary(content)

	return pk
}

func deserializeSk() *bfv.SecretKey {
	content, err := ioutil.ReadFile("sk.key")
	if err != nil {
		log.Fatalln(err)
	}

	sk := bfv.NewSecretKey(params)
	sk.UnmarshalBinary(content)

	return sk
}

func main() {
	// creater(clientAdmin, "clientAdmin.txt")
	// creater(clientUser1, "clientUser1.txt")

	enroll(clientAdmin, "./pk.key")
	enroll(clientUser1, "./pk.key")

	pk := deserializePk()
	fromBalance, err := encryptAmount(pk, 100)
	if err != nil {
		log.Fatalln(err)
	}

	fromTransfer, err := encryptAmount(pk, -44)
	if err != nil {
		log.Fatalln(err)
	}

	toTransfer, err := encryptAmount(pk, 44)
	if err != nil {
		log.Fatalln(err)
	}

	transfer(clientAdmin, "User1@org1.example.com", fromBalance, fromTransfer, toTransfer)

	balance(clientAdmin)
	balance(clientUser1)
	sdk.Close()
}

func encryptAmount(pk *bfv.PublicKey, num int64) ([]byte, error) {
	pt := bfv.NewPlaintext(params)
	rider := []int64{num}

	encoder := bfv.NewEncoder(params)
	encoder.EncodeInt(rider, pt)

	encryptorPk := bfv.NewEncryptorFromPk(params, pk)
	tmp := encryptorPk.EncryptNew(pt)
	return tmp.MarshalBinary()
}

func enroll(client *channel.Client, pkPath string) {
	content, err := ioutil.ReadFile(pkPath)
	if err != nil {
		fmt.Println(err)
	}

	queryArgs := [][]byte{content}
	response := execute(client, "enroll", queryArgs)

	fmt.Println(response.ChaincodeStatus, string(response.Payload))
}

func creater(client *channel.Client, path string) {
	queryArgs := [][]byte{[]byte("no used")}
	response := execute(client, "creater", queryArgs)

	if err := ioutil.WriteFile(path, response.Payload, 0666); err != nil {
		log.Fatalf("WriteFile %s: %v", path, err)
	}
}

func execute(client *channel.Client, fcn string, queryArgs [][]byte) channel.Response {
	response, err := client.Execute(channel.Request{
		ChaincodeID: ccName,
		Fcn:         fcn,
		Args:        queryArgs,
	})

	if err != nil {
		log.Fatalln(err)
	}

	return response
}

func transfer(client *channel.Client, userName string, fromBalance, fromTransfer, toTransfer []byte) {
	queryArgs := [][]byte{
		[]byte(userName),
		fromBalance,
		fromTransfer,
		toTransfer,
	}

	response := execute(client, "transfer", queryArgs)
	fmt.Println(response.ChaincodeStatus, string(response.Payload))
}

func balance(client *channel.Client) {
	queryArgs := [][]byte{}

	response := execute(client, "balance", queryArgs)

	balanceCiperText := new(bfv.Ciphertext)
	err = balanceCiperText.UnmarshalBinary(response.Payload)
	if err != nil {
		log.Fatalln(err)
	}
	sk := deserializeSk()
	decryptor := bfv.NewDecryptor(params, sk)
	resultEncoded := decryptor.DecryptNew(balanceCiperText)

	encoder := bfv.NewEncoder(params)
	result := encoder.DecodeIntNew(resultEncoded)
	fmt.Println(response.ChaincodeStatus, result[:1])

}
