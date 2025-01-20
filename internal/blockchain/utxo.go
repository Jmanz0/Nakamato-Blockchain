package blockchain

import "fmt"

// UTXOs for an address
type UTXOSet struct {
	utxos map[string][]UTXO
}

type UTXO struct {
	TxID    string
	Index   int
	Amount  int64
	Address string
}

func NewUTXOSet() *UTXOSet {
	return &UTXOSet{utxos: make(map[string][]UTXO)}
}

func (u *UTXOSet) PrintAllUTXOs() {
	fmt.Println("[UTXOSet] Current UTXOs:")
	for address, utxos := range u.utxos {
		fmt.Printf("Address: %s\n", address)
		for _, utxo := range utxos {
			fmt.Printf("\tTxID: %s, Index: %d, Amount: %d\n", utxo.TxID, utxo.Index, utxo.Amount)
		}
	}
}

func (u *UTXOSet) Get(address string) []UTXO {
	if val, ok := u.utxos[address]; ok {
		return val
	}
	return []UTXO{}
}

func (u *UTXOSet) CheckUTXO(utxo UTXO) bool {
	for _, ut := range u.Get(utxo.Address) {
		if ut == utxo {
			return true
		}
	}
	return false
}

func (u *UTXOSet) CheckTransaction(tx Transaction) bool {
	if !tx.Verify() {
		return false
	}

	for _, utxo := range tx.Content.InputUTXOs {
		if !u.CheckUTXO(utxo) {
			return false
		}
	}
	return true
}

func (u *UTXOSet) AddUTXO(utxo UTXO) {
	u.utxos[utxo.Address] = append(u.Get(utxo.Address), utxo)
}

func (u *UTXOSet) RemoveUTXO(utxo UTXO) error {
	if !u.CheckUTXO(utxo) {
		return fmt.Errorf("UTXO not found")
	}

	utxos := u.Get(utxo.Address)
	for i, ut := range utxos {
		if ut == utxo {
			utxos = append(utxos[:i], utxos[i+1:]...)
			break
		}
	}
	u.utxos[utxo.Address] = utxos
	return nil
}

func (u *UTXOSet) AddTransaction(tx Transaction) error {
	if !u.CheckTransaction(tx) {
		return fmt.Errorf("invalid transaction")
	}

	for _, utxo := range tx.Content.InputUTXOs {
		err := u.RemoveUTXO(utxo)
		if err != nil {
			return err
		}
	}

	for i := range tx.Content.OutputUTXOs {
		utxo, _ := tx.GetUTXO(i)
		u.AddUTXO(utxo)
	}

	return nil
}

func (u *UTXOSet) AddBlock(b *Block) error {
	for _, tx := range b.Content.Transactions {
		if !u.CheckTransaction(tx) {
			return fmt.Errorf("invalid transaction in block")
		}
	}
	for _, tx := range b.Content.Transactions {
		err := u.AddTransaction(tx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *UTXOSet) RemoveBlock(b *Block) error {
	for i := len(b.Content.Transactions) - 1; i >= 0; i-- {
		tx := b.Content.Transactions[i]

		for _, utxo := range tx.Content.OutputUTXOs {
			err := u.RemoveUTXO(utxo)
			if err != nil {
				return fmt.Errorf("failed to remove UTXO: %v", err)
			}
		}

		// Restore inputs: Add UTXOs that were consumed by this transaction
		for _, utxo := range tx.Content.InputUTXOs {
			u.AddUTXO(utxo)
		}
	}

	return nil
}

func (u *UTXOSet) Balance(address string) int64 {
	balance := int64(0)
	for _, utxo := range u.Get(address) {
		balance += utxo.Amount
	}
	return balance
}

func (u *UTXOSet) GetUTXOs(address string, amount int64) ([]UTXO, error) {
	if u.Balance(address) < amount {
		return nil, fmt.Errorf("not enough funds")
	}

	utxos := u.Get(address)
	var selectedUTXOs []UTXO
	var total int64

	for _, utxo := range utxos {
		if total >= amount {
			break
		}
		selectedUTXOs = append(selectedUTXOs, utxo)
		total += utxo.Amount
	}

	return selectedUTXOs, nil
}
