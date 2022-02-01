package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"

	"github.com/eunjee33/nomadcoin/utils"
)

const (
	fileName = "nomadcoin.wallet"
)

type wallet struct {
	privateKey *ecdsa.PrivateKey
	Address    string // public key (16진수 문자열) = 사람들이 돈을 보내는 주소
}

var w *wallet

func hasWalletFile() bool {
	_, err := os.Stat(fileName) // 해당 파일이 없다면 에러 반환
	return !os.IsNotExist(err)  // 위에서 발생한 에러가 파일이 존재하지 않을 때 발생하는 에러라면 true 반환
}

func createPrivKey() *ecdsa.PrivateKey {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	utils.HandleErr(err)
	return privKey
}

func persistKey(key *ecdsa.PrivateKey) {
	bytes, err := x509.MarshalECPrivateKey(key) // privateKey -> byet[]
	utils.HandleErr(err)
	err = os.WriteFile(fileName, bytes, 0644) // 읽기와 쓰기 허용
	utils.HandleErr(err)
}

func restoreKey() (key *ecdsa.PrivateKey) { // name return : variable을 미리 초기화시켜주고 해당 variable을 알아서 return
	keyAsBytes, err := os.ReadFile(fileName)
	utils.HandleErr(err)
	key, err = x509.ParseECPrivateKey(keyAsBytes)
	utils.HandleErr(err)
	return
}

func encodeBigInts(a, b []byte) string {
	z := append(a, b...)
	return fmt.Sprintf("%x", z)
}

func aFromK(key *ecdsa.PrivateKey) string {
	return encodeBigInts(key.X.Bytes(), key.Y.Bytes()) 
}

func Sign(payload string, w *wallet) string {
	payloadAsB, err := hex.DecodeString(payload)
	utils.HandleErr(err)
	r, s, err := ecdsa.Sign(rand.Reader, w.privateKey, payloadAsB)
	utils.HandleErr(err)
	return encodeBigInts(r.Bytes(), s.Bytes())
}

func restoreBigInts(payload string) (*big.Int, *big.Int, error) {
	bytes, err := hex.DecodeString(payload)
	if err != nil {
		return nil, nil, err
	}
	firstHalfBytes := bytes[:len(bytes)/2]
	secondHalfBytes := bytes[len(bytes)/2:]
	bigA, bigB := big.Int{}, big.Int{}
	bigA.SetBytes(firstHalfBytes)
	bigB.SetBytes(secondHalfBytes)
	return &bigA, &bigB, nil
}

func Verify(signature, payload, address string) bool {
	r, s, err := restoreBigInts(signature) // 서명을 R과 S로 복원
	utils.HandleErr(err)
	x, y, err := restoreBigInts(address) // public key의 x와 y 만들기
	utils.HandleErr(err)
	publicKey := ecdsa.PublicKey{ // private key를 모르는 상태로 public key 만드는 법
		Curve: 	elliptic.P256(),
		X: 		x,
		Y:		y,
	}
	payloadBytes, err := hex.DecodeString(payload)
	utils.HandleErr(err)
	ok := ecdsa.Verify(&publicKey, payloadBytes, r, s) // 검증하기
	return ok
}

func Wallet() *wallet {
	if w == nil {
		w = &wallet{}
		if hasWalletFile() {
			w.privateKey = restoreKey()
		} else {
			key := createPrivKey()
			persistKey(key)
			w.privateKey = key
		}
		w.Address = aFromK(w.privateKey)
	}
	return w
}
