package blockchain

import (
	"errors"
	"strings"
	"time"

	"github.com/eunjee33/nomadcoin/db"
	"github.com/eunjee33/nomadcoin/utils"
)

type Block struct {
	Hash		string	`json:"hash"`
	PrevHash	string	`json:"prevHash,omitempty"`
	Height		int		`json:"height"`
	Difficulty	int		`json:"difficulty`
	Nonce		int 	`json:"nonce`
	Timestamp 	int		`json:"timestamp`
	Transactions []*Tx 	`json:"transactions"` // data string 대신 transaction을 넣어준다
}

func (b *Block) persist(){
	db.SaveBlock(b.Hash, utils.ToBytes(b))
}

var ErrNotFound = errors.New("block not found")

func (b *Block) restore(data []byte){
	utils.FromBytes(b, data)
}

func FindBlock(hash string) (*Block, error){
	blockBytes := db.Block(hash) //해당 해쉬를 가진 블록을 찾음
	if blockBytes == nil { 
		return nil, ErrNotFound	// -> 못찾으면 nil을 return
	}
	block := &Block{}
	block.restore(blockBytes)
	return block, nil	// 찾으면 해당 블록을 return
}

func (b *Block) mine(){
	target := strings.Repeat("0", b.Difficulty)
	for {
		b.Timestamp = int(time.Now().Unix()) //채굴할 때 마다 timestamp 설정
		hash := utils.Hash(b)
		if strings.HasPrefix(hash, target){
			b.Hash = hash
			break
		} else {
			b.Nonce++
		}
	}
}

func createBlock(prevHash string, height, diff int) *Block {
	block := &Block{
		Hash:     "",
		PrevHash: prevHash,
		Height:   height,
		Difficulty: diff,
		Nonce: 0,	
	}
	block.mine()
	block.Transactions =  Mempool.TxToConfirm() //mine()이 언제끝날지 모르므로 transactions은 mine() 후에 선언
	block.persist()
	return block
}