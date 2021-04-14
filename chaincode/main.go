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

var params *bfv.Parameters

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

func userName(rawCert []byte) (string, error) {
	block, _ := pem.Decode(rawCert)
	if block == nil || block.Type != "CERTIFICATE" {
		return "", fmt.Errorf("Wrong PEM encoding:", block.Type)

	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}

	return cert.Subject.CommonName, nil
}

// return shim.Error(err.Error())
// return shim.Success([]byte(result))

func enroll(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	rawCert, err := stub.GetCreator()
	if err != nil {
		return shim.Error("GetCreator:" + err.Error())
	}
	tmp := bytes.Index(rawCert, []byte{'-'})
	userName, err := userName(rawCert[tmp:])

	if err != nil {
		return shim.Error("on userName:" + err.Error())
	}

	pk := args[0]
	err = stub.PutState(userName, []byte(pk))
	if err != nil {
		return shim.Error(fmt.Sprintf("%s enroll failed,err:%s", userName, err))
	}

	return shim.Success([]byte(fmt.Sprintf("%s enroll successfully", userName)))
}

func balanceAccount(userName string) string {
	return strings.Join([]string{userName, "Balance"}, "-")
}

func fromAccouunt(stub shim.ChaincodeStubInterface) (string, error) {
	rawCert, err := stub.GetCreator()
	if err != nil {
		return "GetCreator err:", err
	}

	tmp := bytes.Index(rawCert, []byte{'-'})
	userName, err := userName(rawCert[tmp:])
	if err != nil {
		return "", err
	}

	return balanceAccount(userName), nil
}

func creater(stub shim.ChaincodeStubInterface) peer.Response {
	tmp, e := stub.GetCreator()
	if e != nil {
		return shim.Error("on creater:" + e.Error())
	}
	return shim.Success(tmp)
}

func balance(stub shim.ChaincodeStubInterface) peer.Response {
	fromAcc, err := fromAccouunt(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	balanceCiper, err := stub.GetState(fromAcc)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(balanceCiper)
}

func transfer(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 4 {
		tmp := fmt.Sprintf("current args len:%d,%s", len(args), "Incorrect arguments. Args: fn toAccount x1,x2,y1")
		return shim.Error(tmp)
	}

	fromBalance := []byte(args[1])
	// TODO need check fromTransfer ?= toTransfer, fromTransfer >= 0 use ZKP
	// fromTransfer := []byte(args[2])
	toTransfer := []byte(args[3])

	fromAcc, err := fromAccouunt(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	toAcc := balanceAccount(args[0])

	balanceCiper, err := stub.GetState(toAcc)
	if err != nil {
		return shim.Error("get balance failure")
	}

	if balanceCiper == nil {
		stub.PutState(toAcc, toTransfer)
		stub.PutState(fromAcc, fromBalance)
		return shim.Success([]byte("transfer successfully(first reception)"))
	}

	toBalCiphertext := new(bfv.Ciphertext)
	err = toBalCiphertext.UnmarshalBinary(balanceCiper)
	if err != nil {
		tmp := fmt.Sprintf("balanceCiper UnmarshalBinary err:%s", err.Error())
		return shim.Error(tmp)
	}

	toTransCiphertext := new(bfv.Ciphertext)
	err = toTransCiphertext.UnmarshalBinary(toTransfer)
	if err != nil {
		tmp := fmt.Sprintf("y1 UnmarshalBinary err:%s", err.Error())
		return shim.Error(tmp)
	}

	degree := toTransCiphertext.Degree()
	resultToCiphertext := bfv.NewCiphertext(params, degree)

	// return fmt.Sprintf("1, len:%d,%d degree:%d", len(toBalBytes), len(y1), degree), nil
	evaluator := bfv.NewEvaluator(params)
	evaluator.Add(toBalCiphertext, toTransCiphertext, resultToCiphertext)

	resultBytes, err := resultToCiphertext.Ciphertext().MarshalBinary()
	if err != nil {
		tmp := fmt.Sprintf("resultToCiphertext.Ciphertext().MarshalBinary err:%s", err.Error())
		return shim.Error(tmp)
	}

	stub.PutState(toAcc, resultBytes)
	return shim.Success([]byte("transfer successfully"))
}

func init() {
	params = bfv.DefaultParams[bfv.PN13QP218].WithT(0x3ee0001)
}

func main() {
	if err := shim.Start(new(SimpleAsset)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
