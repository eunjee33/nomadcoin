package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
)

func HandleErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func ToBytes(i interface{}) []byte { //파라미터로 뭐든지 받아라
	var aBuffer bytes.Buffer
	encoder := gob.NewEncoder(&aBuffer)
	HandleErr(encoder.Encode(i))
	return aBuffer.Bytes()
}

//여기서 interface i는 복원된(디코딩한) 데이터를 저장할 포인터이다
func FromBytes(i interface{}, data []byte) {
    decoder := gob.NewDecoder(bytes.NewReader(data))
	HandleErr(decoder.Decode(i))
}

func Hash(i interface{}) string {
	s := fmt.Sprintf("%v", i) // %v는 기본 formatter
	hash := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", hash)
}