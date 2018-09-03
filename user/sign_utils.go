package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"fmt"
	"math/big"
)

/**
验证签名
*/
func Verify(pubkey, json, sign string) bool {
	curve := elliptic.P256()
	pubkeyByte, e := base64.StdEncoding.DecodeString(pubkey)
	if e != nil {
		fmt.Println(e)
	}

	//拆分签名文件
	r := big.Int{}
	s := big.Int{}
	signByte, e := base64.StdEncoding.DecodeString(sign)
	if e != nil {
		fmt.Println(e)
	}
	sigLen := len(signByte)
	r.SetBytes(signByte[:(sigLen / 2)])
	s.SetBytes(signByte[(sigLen / 2):])

	//拆分公钥
	x := big.Int{}
	y := big.Int{}
	keyLen := len(pubkeyByte)
	x.SetBytes(pubkeyByte[:(keyLen / 2)])
	y.SetBytes(pubkeyByte[(keyLen / 2):])
	//还原为原始公钥
	rawPubKey := ecdsa.PublicKey{curve, &x, &y}
	//公钥、签名文件、原始数据确认签名有效性
	if ecdsa.Verify(&rawPubKey, []byte(json), &r, &s) == false {
		return false
	}
	return true
}

/**
公钥base64解码
*/
func PubkeyToBytes(pubkey string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(pubkey)
}
