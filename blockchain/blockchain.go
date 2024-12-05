package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type Transaction struct {
	ID     string
	Amount int
}

type Block struct {
	Index        int
	Timestamp    string
	Transactions []Transaction
	PreviousHash string
	Hash         string
	MinedBy      string
}

type Blockchain struct {
	Blocks         []Block
	TransactionPool []Transaction
}

func NewBlockchain() *Blockchain {
	return &Blockchain{
		Blocks:         []Block{createGenesisBlock()},
		TransactionPool: []Transaction{},
	}
}

func createGenesisBlock() Block {
	return Block{
		Index:        0,
		Timestamp:    time.Now().String(),
		Transactions: []Transaction{},
		PreviousHash: "0",
		Hash:         calculateHash(0, time.Now().String(), []Transaction{}, "0"),
		MinedBy:      "Genesis",
	}
}

func calculateHash(index int, timestamp string, transactions []Transaction, previousHash string) string {
	record := fmt.Sprintf("%d%s%v%s", index, timestamp, transactions, previousHash)
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func (bc *Blockchain) AddTransaction(tx Transaction) {
	bc.TransactionPool = append(bc.TransactionPool, tx)
}

func (bc *Blockchain) MineBlock(minerName string) {
	if len(bc.TransactionPool) == 0 {
		fmt.Println("No transactions to mine")
		return
	}

	prevBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock := Block{
		Index:        prevBlock.Index + 1,
		Timestamp:    time.Now().String(),
		Transactions: bc.TransactionPool,
		PreviousHash: prevBlock.Hash,
		Hash:         calculateHash(prevBlock.Index+1, time.Now().String(), bc.TransactionPool, prevBlock.Hash),
		MinedBy:      minerName,
	}
	bc.Blocks = append(bc.Blocks, newBlock)
	bc.TransactionPool = []Transaction{} // Clear the transaction pool
	fmt.Printf("Mined new block: %+v\n", newBlock)
}

func (bc *Blockchain) IsValid() bool {
	for i := 1; i < len(bc.Blocks); i++ {
		currentBlock := bc.Blocks[i]
		prevBlock := bc.Blocks[i-1]

		if currentBlock.Hash != calculateHash(currentBlock.Index, currentBlock.Timestamp, currentBlock.Transactions, currentBlock.PreviousHash) {
			return false
		}

		if currentBlock.PreviousHash != prevBlock.Hash {
			return false
		}
	}
	return true
} 