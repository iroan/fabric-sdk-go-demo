package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"testing"
)

func TestStrHandle(t *testing.T) {
	tmp1, err := ioutil.ReadFile("clientAdmin.txt")
	if err != nil {
		log.Fatalln(err)
	}

	clientAdmin := string(tmp1)
	tmp2, err := ioutil.ReadFile("clientUser1.txt")
	if err != nil {
		log.Fatalln(err)
	}

	clientUser1 := string(tmp2)

	fmt.Println(clientUser1[strings.Index(clientUser1, "-"):])
	fmt.Println(clientAdmin[strings.Index(clientAdmin, "-"):])

	tmp3 := bytes.Index(tmp1, []byte{'-'})
	fmt.Println(tmp3)
	fmt.Println(tmp1[tmp3:])
	fmt.Println(string(tmp1[tmp3:]))
}
