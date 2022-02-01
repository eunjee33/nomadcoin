package blockchain

import (
	"sync"

	"github.com/eunjee33/nomadcoin/db"
	"github.com/eunjee33/nomadcoin/utils"
)

const (
    defaultDifficulty int = 2
    difficultyInterval int = 5
    blockInterval int = 2 // 매 2분마다 블록 1개가 생성되게끔
    allowedRange int = 2
)

type blockchain struct{
    NewestHash string `json:"newstHash"`
    Height int `json:"height"`
    CurrentDifficulty int `json:"currentDifficulty"`
}

var b *blockchain
var once sync.Once

func (b *blockchain) restore(data []byte){
    utils.FromBytes(b, data)
}

func persistBlockchain(b *blockchain) {
    db.SaveCheckpoint(utils.ToBytes(b))
}

func (b *blockchain) AddBlock () {
    block := createBlock(b.NewestHash, b.Height + 1, getDifficulty(b))
    b.NewestHash = block.Hash
    b.Height = block.Height
    b.CurrentDifficulty = block.Difficulty // 최신 difficulty를 가져온다
    persistBlockchain(b)
}

//모든 Block 가져오기
func Blocks(b *blockchain) []*Block {
    var blocks []*Block
    hashCursor := b.NewestHash
    for {
        block, _ := FindBlock(hashCursor) //blockchain에 있는 hash로 블록을 찾기때문에 에러가 날 일 없다
        blocks = append(blocks ,block)
        if block.PrevHash != "" { 
            hashCursor = block.PrevHash
        } else { //Genesis block이 이라면 처음 block까지 다 찾은 것이므로 종료
            break
        }
    }
    return blocks
}

//모든 transaction들을 찾아주는 함수
func Txs(b *blockchain) []*Tx {
    var txs []*Tx
    for _, block := range Blocks(b) {
        txs = append(txs, block.Transactions...)
    }
    return txs
}

// 하나의 transaction만 찾아주는 함수
func FindTx(b *blockchain, targetID string) *Tx {
    for _, tx := range Txs(b) {
        if tx.ID == targetID {
            return tx
        }
    }
    return nil
}

func recalculateDifficulty(b *blockchain) int {
    allBlocks := Blocks(b)
    newestBlock := allBlocks[0] // 가장 최근에 추가된 블록
    lastRecalculatedBlock := allBlocks[difficultyInterval - 1] // 가장 최근에 difficulty가 재설정된 블록
    actualTime := (newestBlock.Timestamp - lastRecalculatedBlock.Timestamp) / 60
    expectedTime := difficultyInterval * blockInterval
    if actualTime <= (expectedTime - allowedRange) {
        return b.CurrentDifficulty + 1
    } else if actualTime >= (expectedTime + allowedRange) {
        return b.CurrentDifficulty - 1
    }
    return b.CurrentDifficulty
}

func getDifficulty(b *blockchain) int {
    if b.Height == 0 {
        return defaultDifficulty 
    } else if (b.Height % difficultyInterval == 0) { 
        return recalculateDifficulty(b)
    } else {
        return b.CurrentDifficulty
    }
}

// 특정 주소의 소유주를 가진 출력값들만 필터링 -> API에서 호출할거라 export해주기(첫글자 대문자)
func UTxOutsByAddress (address string, b *blockchain) []*UTxOut {
    var uTxOuts []*UTxOut
    creatorTxs := make(map[string]bool) //map 구조체 생성

    for _, block := range Blocks(b){
        for _, tx := range block.Transactions {
            for _, input := range tx.TxIns {
                if input.Signature == "COiNBASE" { // coinbase의 경우 input을 확인하지 않아도 됨
                    break
                }
                if FindTx(b, input.TxID).TxOuts[input.Index].Address == address { //input엔 더이상 Owner이 없으니 input으로 전 output을 찾아가 그 address와 같은지 확인
                    creatorTxs[input.TxID] = true
                }
            }
            for index, output:= range tx.TxOuts {
                if output.Address == address {// output이 creatorTxs안에 있는 트랜잭션 내에 없다는 것을 확인
                    if _, ok := creatorTxs[tx.ID]; !ok { // 이 output을 생성한 트랜잭션을 caretorTxs에서 못찾았을 때
                        //unspent transaction output 찾음!! (input으로 참조되지 않은 output의 transaction 찾음)   
                        uTxOut := &UTxOut{tx.ID, index, output.Amount}
                        if !isOnMempool(uTxOut) {
                            uTxOuts = append(uTxOuts, uTxOut)
                        }
                    }
                }
            }
        }
    }    
    return uTxOuts
}

// 출력값들을 모아 전재산이 얼만지 보여줌
func BalanceByAddress(address string, b *blockchain) int {
    txOuts := UTxOutsByAddress(address, b) //새로 만들어준 함수 이름으로 바꾸기
    var amount int
    for _, txOut := range txOuts {
        amount += txOut.Amount
    }
    return amount
}

//blockchain 초기화 및 추출 (export)
func Blockchain() *blockchain {
    once.Do(func() {
        b = &blockchain{ //height만 0으로 해주면 됨
            Height: 0,
        }
        checkpoint := db.Checkpoint() 
        if checkpoint == nil {
            b.AddBlock()
        } else{
            b.restore(checkpoint)
        }            
    })
    return b
}
