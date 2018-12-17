package util

import (
	"fmt"
	"errors"
	"encoding/pem"
	"strings"
)

type Signer interface {
	Sign(src []byte) ([]byte, error)
	Verify(src []byte, sign []byte) error
	SetPrivKey(privatekey []byte) error
	SetPubKey(publickey []byte) error
}

type Sign_Type int
const (
	SIGN_TYPE_ECC Sign_Type = 1 + iota
	SIGN_TYPE_RSA
)

/*
type Key_Type int

const (
	PRIVATE_KEY Key_Type = iota
	PUBLIC_KEY
)
*/

func NewFromPem(key string,t Sign_Type)(Signer,error)  {
	blockkey, rest := pem.Decode([]byte(key))
	if blockkey == nil {
		fmt.Println(string(rest))
		return nil, errors.New("key error")
	}
	var sg Signer
	switch t {
	case SIGN_TYPE_ECC:
		//sg = new(eccClient)
	case SIGN_TYPE_RSA:
		sg = new(rsaClient)
	}
	if strings.Contains(blockkey.Type,"PRIVATE KEY") {
		sg.SetPrivKey(blockkey.Bytes)
	}else {
		sg.SetPubKey(blockkey.Bytes)
	}
	return sg,nil
}