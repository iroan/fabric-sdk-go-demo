/*
 * Copyright IBM Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/ldsec/lattigo/v2/bfv"
)

type SimpleAsset struct {
}

var (
	params             *bfv.Parameters
	initAccountBalance int64
)

func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()

	if fn == "enroll" {
		return enroll(stub, args)
	} else if fn == "transfer" {
		return transfer(stub, args)
	} else if fn == "balance" {
		return balance(stub)
	} else if fn == "creater" {
		return creater(stub)
	}

	return shim.Error("no such function")
}

func creatorName(stub shim.ChaincodeStubInterface) (string, error) {
	rawCert, err := stub.GetCreator()
	if err != nil {
		return "GetCreator err:", err
	}

	tmp := bytes.Index(rawCert, []byte{'-'})

	block, _ := pem.Decode(rawCert[tmp:])
	if block == nil || block.Type != "CERTIFICATE" {
		return "", fmt.Errorf("WRONG PEM ENCODING:%s", block.Type)

	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}

	return cert.Subject.CommonName, nil
}

func enroll(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// get username
	name, err := creatorName(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	// add to account map
	publicKey := []byte(args[0])
	err = stub.PutState(getEnrollKey(name), publicKey)
	if err != nil {
		return shim.Error(fmt.Sprintf("%s enroll failed,err:%s", getEnrollKey(name), err))
	}

	// restore the public key
	pk := bfv.NewPublicKey(params)
	if err != nil {
		return shim.Error(err.Error())
	}

	pk.UnmarshalBinary(publicKey)

	// set account balance
	pt := bfv.NewPlaintext(params)
	balance := []int64{initAccountBalance}

	encoder := bfv.NewEncoder(params)
	encoder.EncodeInt(balance, pt)

	encryptorPk := bfv.NewEncryptorFromPk(params, pk)
	balanceCiper := encryptorPk.EncryptNew(pt)

	balanceBin, err := balanceCiper.MarshalBinary()
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(getBalanceKey(name), balanceBin)
	if err != nil {
		return shim.Error(err.Error())
	}

	// enroll success
	return shim.Success([]byte(fmt.Sprintf("%s enroll successfully", name)))
}

func getBalanceKey(userName string) string {
	return strings.Join([]string{userName, "Balance"}, "-")
}

func getEnrollKey(userName string) string {
	return strings.Join([]string{userName, "Balance"}, "-")
}

func isEnrolled(stub shim.ChaincodeStubInterface, userName string) bool {
	val, _ := stub.GetState(getEnrollKey(userName))
	return val != nil
}

func creater(stub shim.ChaincodeStubInterface) peer.Response {
	tmp, e := stub.GetCreator()
	if e != nil {
		return shim.Error("on creater:" + e.Error())
	}
	return shim.Success(tmp)
}

func balance(stub shim.ChaincodeStubInterface) peer.Response {
	name, err := creatorName(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	balanceCiper, err := stub.GetState(getBalanceKey(name))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(balanceCiper)
}

func transfer(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// check params
	if len(args) != 3 {
		tmp := fmt.Sprintf("current args len:%d,%s", len(args), "Incorrect arguments. Args: fn toUser,fromAmount,toAmount")
		return shim.Error(tmp)
	}

	fromUser, err := creatorName(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	toUser := args[0]

	// check whether the two account has been enrolled.
	if !isEnrolled(stub, fromUser) || !isEnrolled(stub, toUser) {
		tmp := fmt.Sprintf("%s or %s is not enrolled.", fromUser, toUser)
		return shim.Error(tmp)
	}

	// TODO need check dec(fromTranster) + dec(toTransfer) == 0
	fromAmount := []byte(args[1])
	toAmount := []byte(args[2])

	// update balance
	if updateBalance(stub, fromUser, fromAmount) != nil || updateBalance(stub, toUser, toAmount) != nil {
		return shim.Error("failure")
	}

	return shim.Success([]byte("success"))
}

func updateBalance(stub shim.ChaincodeStubInterface, userName string, amountCiper []byte) error {
	tmp, err := stub.GetState(getBalanceKey(userName))
	if err != nil {
		return err
	}

	ciper1 := new(bfv.Ciphertext)
	err = ciper1.UnmarshalBinary(tmp)
	if err != nil {
		return err
	}

	ciper2 := new(bfv.Ciphertext)
	err = ciper2.UnmarshalBinary(amountCiper)
	if err != nil {
		return err
	}

	degree := ciper2.Degree()
	ciper3 := bfv.NewCiphertext(params, degree)

	evaluator := bfv.NewEvaluator(params)
	evaluator.Add(ciper1, ciper2, ciper3)

	balance, err := ciper3.Ciphertext().MarshalBinary()
	if err != nil {
		return err
	}

	stub.PutState(getBalanceKey(userName), balance)

	return nil
}

func init() {
	params = bfv.DefaultParams[bfv.PN13QP218].WithT(0x3ee0001)
	initAccountBalance = 100
}

func main() {
	if err := shim.Start(new(SimpleAsset)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
