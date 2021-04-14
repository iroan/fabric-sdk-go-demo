/*
 * Copyright IBM Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/ldsec/lattigo/v2/bfv"
)

var params *bfv.Parameters
var numCiperTextFilename string

func genTwoCiperBytes(pk *bfv.PublicKey) {
	pt := bfv.NewPlaintext(params)
	rider := make([]int64, 2)
	rider[0] = -3
	rider[1] = 11

	encoder := bfv.NewEncoder(params)
	encoder.EncodeInt(rider, pt)
	encryptorPk := bfv.NewEncryptorFromPk(params, pk)
	num1 := encryptorPk.EncryptNew(pt)

	data, err := num1.MarshalBinary()
	if err != nil {
		fmt.Errorf("%v", err)
		return
	}

	if err := ioutil.WriteFile(numCiperTextFilename, data, 0666); err != nil {
		log.Fatalf("WriteFile %s: %v", numCiperTextFilename, err)
	}
}

func deserializePk() *bfv.PublicKey {
	content, err := ioutil.ReadFile("pk.key")
	if err != nil {
		log.Fatalln(err)
	}

	pk := bfv.NewPublicKey(params)
	if err != nil {
		log.Fatalln("%v", err)
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

func serializeNewKeyPair() {
	kgen := bfv.NewKeyGenerator(params)
	sk, pk := kgen.GenKeyPair()
	bsk, err := sk.MarshalBinary()
	if err != nil {
		fmt.Errorf("%v", err)
	}

	bpk, err := pk.MarshalBinary()
	if err != nil {
		fmt.Errorf("%v", err)
	}

	// tmp1 := hex.EncodeToString(bpk)
	// fmt.Println("PublicKey Len:", len(tmp1))
	// err = ioutil.WriteFile("pk.key", []byte(tmp1), 0666)
	fmt.Println("PublicKey Len:", len(bpk))
	err = ioutil.WriteFile("pk.key", bpk, 0666)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("SecretKey Len:", len(bsk))
	err = ioutil.WriteFile("sk.key", bsk, 0666)
	if err != nil {
		log.Fatalln(err)
	}

}

func enroll() {
	rawCert, err := ioutil.ReadFile("my.cert")
	if err != nil {
		fmt.Errorf("%v", err)
		return
	}
	block, _ := pem.Decode(rawCert)
	if block == nil || block.Type != "CERTIFICATE" {
		fmt.Errorf("%s: wrong PEM encoding")
		return
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		fmt.Errorf("%s: wrong DER encoding")
		return
	}

	fmt.Println("CommonName:", cert.Subject.CommonName)
}

func transfer() {
	// fromBalanceCiperText := ""
	// toBalanceCiperText := "amountCipertext" + prevBalanceCiperText
	// stub.PutState(fromCommonName,fromBalanceCiperText)
	// stub.PutState(toCommonName,toBalanceCiperText)
}

func init() {
	params = bfv.DefaultParams[bfv.PN12QP101pq].WithT(0x3ee0001)
	fmt.Println("params", params)
	numCiperTextFilename = "numCiper.bin"

}

func main() {
	serializeNewKeyPair()
	pk := deserializePk()
	sk := deserializeSk()
	// enroll()
	// transfer()
	genTwoCiperBytes(pk)
	add(sk, pk)
}

func add(sk *bfv.SecretKey, pk *bfv.PublicKey) {
	encoder := bfv.NewEncoder(params)

	decryptor := bfv.NewDecryptor(params, sk)
	evaluator := bfv.NewEvaluator(params)

	tmp1, err := ioutil.ReadFile(numCiperTextFilename)
	if err != nil {
		log.Fatalln(err)
	}

	numCiperText := new(bfv.Ciphertext)
	err = numCiperText.UnmarshalBinary(tmp1)
	if err != nil {
		log.Fatalln(err)
	}

	degree := numCiperText.Degree()
	resultCiper := bfv.NewCiphertext(params, degree)

	// 同态加
	evaluator.Add(numCiperText, numCiperText, resultCiper)

	// 用sk解密
	resultEncoded := decryptor.DecryptNew(resultCiper)
	result := encoder.DecodeIntNew(resultEncoded)

	fmt.Println(result[:2])
}
