package blockchain

import (
	"errors"
	"time"

	"github.com/eunjee33/nomadcoin/utils"
	"github.com/eunjee33/nomadcoin/wallet"
)

const (
	minerReward int = 50
)

type mempool struct {
	Txs []*Tx
}

var Mempool *mempool = &mempool{} // mempool은 메모리에 저장 (DB에서 복원 X) 따라서 초기화 과정도 X

type Tx struct {
	ID        string	`json:"id"`
	Timestamp int		`json:"timestamp"`
	TxIns     []*TxIn	`json:"txIns"`
	TxOuts    []*TxOut	`json:"txOuts"`
}

type TxIn struct {
	TxID string		`json:"txId` 
	Index int		`json:"index"`
	Signature string `json:"signature` // 아무나 넣을 수 없게 Owner를 서명으로 바꾼다
}

type TxOut struct {
	Address string 	`json:"address"` // 사람들이 코인 보내는 곳
	Amount int		`json:"amount"`
}

//어떤 output이 쓰였는지 안쓰였는지 확인 가능하게 해주는 struct
type UTxOut struct {
	TxID string	`json:"txId"`
	Index int	`json:"index"`
	Amount int	`json:amount"`
}

func (t *Tx) getId() {
	t.ID = utils.Hash(t)
}

func (t *Tx)sign() {
	for _, txIn := range t.TxIns { // 모든 transaction input들에 대해 서명함
		txIn.Signature = wallet.Sign(t.ID, wallet.Wallet())
	}
}

// transaction output을 소유하고 있다는 걸 증명해야 함
// == 사용할 transaction output의 address를 가져와 서명을 verify 할 수 있어야 한다
func validate(tx *Tx)  bool { 
	valid := true
	for _, txIn := range tx.TxIns {
		prevTx := FindTx(Blockchain(), txIn.TxID) 
		if prevTx == nil { // 이전 transaction이 blockchain에 없다면
			valid = false
			break
		}
		address := prevTx.TxOuts[txIn.Index].Address // == public key
		valid = wallet.Verify(txIn.Signature, tx.ID, address)
		if !valid {
			break
		}
	}
	return valid
}

// unspent transaction output이 mempool에서 사용되고 있는지 확인하는 함수
func isOnMempool(UTxOut *UTxOut) bool {
	exists := false
Outer: //여러 개의 for loop이 중첩됐을 경우 바깥쪽 for loop롤 종료시킬 수 있다!!
	for _, tx := range Mempool.Txs {
		for _, input := range tx.TxIns {
			if input.TxID == UTxOut.TxID && input.Index == UTxOut.Index {
				exists = true
				break Outer
			}
		}
	}
	return exists
}

var ErrorNoMoney = errors.New("not enough money")
var ErrorNotValid = errors.New("Tx Invalid")

// 채굴자를 주소로 삼는 코인베이스 거래내역을 생성해서 Tx 포인터 return
func makeCoinbaseTx(address string) *Tx {
	txIns := []*TxIn{
		{"", -1, "COINBASE"}, // 
	}
	txOuts := []*TxOut{
		{address, minerReward},
	}
	tx := Tx{
		ID: "",
		Timestamp: int(time.Now().Unix()),
		TxIns: txIns,
		TxOuts: txOuts,
	}
	tx.getId()
	return &tx
}

//output에 있는 금액들로 내가 보내고 싶은 금액보다 크거나 같을때까지 input을 만들어야 함
func makeTx(from, to string, amount int) (*Tx, error) {
	if BalanceByAddress(from, Blockchain()) < amount {
		return nil, ErrorNoMoney
	}
	var txOuts []*TxOut
	var txIns []*TxIn
	total := 0
	uTxOuts := UTxOutsByAddress(from, Blockchain())
	for _, uTxOut := range uTxOuts {
		if total >= amount {
			break
		}
		txIn := &TxIn{uTxOut.TxID, uTxOut.Index, from} //transaction output마다 transaction input 생성
		txIns = append(txIns, txIn)
		total += uTxOut.Amount
	}
	if change := total - amount; change != 0 {
		changeTxOut := &TxOut{from, change}
		txOuts = append(txOuts, changeTxOut)
	}
	txOut := &TxOut{to, amount}
	txOuts = append(txOuts, txOut)
	tx := &Tx{
		ID: "",
		Timestamp: int(time.Now().Unix()),
		TxIns: txIns,
		TxOuts: txOuts,
	}
	tx.getId()
	tx.sign() // transaction id를 만든 후 그 id에 서명
	valid := validate(tx) //해당 transaction 검증
	if !valid {
		return nil, ErrorNotValid
	}
	return tx, nil
}

// 단순히 mempool에 transaction을 추가해주는 함수 (transaction을 만들어주는 함수 X)
func (m *mempool) AddTx(to string, amount int) error { 
	tx, err := makeTx(wallet.Wallet().Address, to, amount)
	if err != nil {
		return err
	}
	m.Txs = append(m.Txs, tx)
	return nil
}
//승인할 transacion 가져오기 == 모든 transaction을 건네주고 mempool을 비워야함
func (m *mempool) TxToConfirm() []*Tx {
	coinbase := makeCoinbaseTx(wallet.Wallet().Address) // 마찬가지
	txs := m.Txs 
	txs = append(txs, coinbase) 
	m.Txs = nil
	return txs
}