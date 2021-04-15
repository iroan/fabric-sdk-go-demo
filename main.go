package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/ldsec/lattigo/v2/bfv"
)

var (
	channelName = "mychannel"
	orgName     = "Org1"
	ccName      = "demo02"
)

var (
	clientAdmin *channel.Client
	clientUser1 *channel.Client
	sdk         *fabsdk.FabricSDK
	err         error
	params      *bfv.Parameters
)

func init() {
	c := config.FromFile("./config.yaml")
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

func pkPath(name string) string {
	return strings.Join([]string{name, "pk.key"}, "-")
}

func skPath(name string) string {
	return strings.Join([]string{name, "sk.key"}, "-")
}

func newKeyPairAndSave(fileName string) (pk *bfv.PublicKey, sk *bfv.SecretKey) {
	kgen := bfv.NewKeyGenerator(params)
	sk, pk = kgen.GenKeyPair()
	bsk, err := sk.MarshalBinary()
	if err != nil {
		log.Fatalln(err)
	}

	bpk, err := pk.MarshalBinary()
	if err != nil {
		log.Fatalln(err)
	}

	err = ioutil.WriteFile(pkPath(fileName), bpk, 0666)
	if err != nil {
		log.Fatalln(err)
	}

	err = ioutil.WriteFile(skPath(fileName), bsk, 0666)
	if err != nil {
		log.Fatalln(err)
	}

	return
}

func deserializePk(fileName string) *bfv.PublicKey {
	content, err := ioutil.ReadFile(pkPath(fileName))
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

func deserializeSk(fileName string) *bfv.SecretKey {
	content, err := ioutil.ReadFile(skPath(fileName))
	if err != nil {
		log.Fatalln(err)
	}

	sk := bfv.NewSecretKey(params)
	sk.UnmarshalBinary(content)

	return sk
}

func transferDemo() {
	adminPk := deserializePk("admin")
	adminSk := deserializeSk("admin")
	user1Pk := deserializePk("user1")
	user1Sk := deserializeSk("user1")

	outLine("admin balance", "user1 balance")
	adminBal := balance(clientAdmin, adminSk)
	user1Bal := balance(clientUser1, user1Sk)
	outLine(strconv.FormatInt(adminBal, 10), strconv.FormatInt(user1Bal, 10))

	from1, _ := encryptAmount(adminPk, -5)
	to1, _ := encryptAmount(user1Pk, 5)
	transfer(clientAdmin, "User1@org1.example.com", from1, to1)

	adminBal = balance(clientAdmin, adminSk)
	user1Bal = balance(clientUser1, user1Sk)
	outLine(strconv.FormatInt(adminBal, 10), strconv.FormatInt(user1Bal, 10))

	from2, _ := encryptAmount(user1Pk, -8)
	to2, _ := encryptAmount(adminPk, 8)
	transfer(clientUser1, "Admin@org1.example.com", from2, to2)

	adminBal = balance(clientAdmin, adminSk)
	user1Bal = balance(clientUser1, user1Sk)
	outLine(strconv.FormatInt(adminBal, 10), strconv.FormatInt(user1Bal, 10))
}

func enrollDemo() {
	adminPk, adminSk := newKeyPairAndSave("admin")
	user1Pk, user1Sk := newKeyPairAndSave("user1")

	enroll(clientAdmin, "admin")
	enroll(clientUser1, "user1")

	_ = adminPk
	_ = adminSk
	_ = user1Pk
	_ = user1Sk
}

func main() {
	// enrollDemo()
	transferDemo()

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

func enroll(client *channel.Client, path string) {
	content, err := ioutil.ReadFile(pkPath(path))
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

func transfer(client *channel.Client, userName string, fromAmountCiper, toAmountCiper []byte) {
	queryArgs := [][]byte{
		[]byte(userName),
		fromAmountCiper,
		toAmountCiper,
	}

	response := execute(client, "transfer", queryArgs)
	// fmt.Println("to", userName, ":", response.ChaincodeStatus, string(response.Payload))
	_ = response
}

func balance(client *channel.Client, sk *bfv.SecretKey) int64 {
	queryArgs := [][]byte{}

	response := execute(client, "balance", queryArgs)

	balanceCiperText := new(bfv.Ciphertext)
	err = balanceCiperText.UnmarshalBinary(response.Payload)
	if err != nil {
		log.Fatalln(err)
	}
	decryptor := bfv.NewDecryptor(params, sk)
	resultEncoded := decryptor.DecryptNew(balanceCiperText)

	encoder := bfv.NewEncoder(params)
	result := encoder.DecodeIntNew(resultEncoded)
	if response.ChaincodeStatus == 200 {
		return result[0]
	}
	return 0
}

func outLine(c1, c2 string) {
	fmt.Printf("|%-20s|%-20s|\n", c1, c2)
}
