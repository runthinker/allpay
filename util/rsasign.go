package util

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"crypto/rand"
	"crypto"
	"crypto/sha256"
	"errors"
)

func GenRsaKey(bits int)([]byte,[]byte)  {
	prk,err := rsa.GenerateKey(rand.Reader,bits)
	if err != nil {
		return nil,nil
	}
	derPriv := x509.MarshalPKCS1PrivateKey(prk)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derPriv,
	}
	privkey := pem.EncodeToMemory(block)
	derPub,err := x509.MarshalPKIXPublicKey(&prk.PublicKey)
	block = &pem.Block{
		Type:"PUBLIC KEY",
		Bytes:derPub,
	}
	pubkey := pem.EncodeToMemory(block)
	return privkey,pubkey
}

type rsaClient struct {
	privkey *rsa.PrivateKey
	pubkey *rsa.PublicKey
}

func (r *rsaClient)Sign(src []byte)([]byte,error)  {
	if r.privkey == nil {
		return nil,errors.New("private key is nil")
	}
	h := sha256.New()
	h.Write(src)
	hashed := h.Sum(nil)
	return rsa.SignPKCS1v15(rand.Reader,r.privkey,crypto.SHA256,hashed)
}

func (r *rsaClient)Verify(src []byte,sign []byte) error  {
	if r.pubkey == nil {
		return errors.New("public key is nil")
	}
	h := sha256.New()
	h.Write(src)
	hashed := h.Sum(nil)
	return rsa.VerifyPKCS1v15(r.pubkey,crypto.SHA256,hashed,sign)
}

func (r *rsaClient)SetPubKey(publickey []byte)(error)  {
	pub,err := x509.ParsePKIXPublicKey(publickey)
	if err != nil {
		return err
	}
	r.pubkey = pub.(*rsa.PublicKey)
	return nil
}

func (r* rsaClient)SetPrivKey(privatekey []byte) error {
	pri,err := x509.ParsePKCS1PrivateKey(privatekey)
	r.privkey = pri
	r.pubkey = &pri.PublicKey
	return err
}

func NewrsaClient() *rsaClient  {
	return &rsaClient{}
}