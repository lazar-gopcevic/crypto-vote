package transaction

import "errors"

type UTXO struct {
	TransactionID []byte
	PublicKeyHash []byte
	Value         int
	Vout          int
}

type UTXOs []UTXO

var ErrUTXONotFound = errors.New("UTXO not found")

var ErrInvalidTxAmount = errors.New("Sums of inputs and outputs for this transaction don't add up")

func (utxos UTXOs) Filter(criteria func(UTXO) bool) UTXOs {
	result := UTXOs{}
	for _, utxo := range utxos {
		if criteria(utxo) {
			result = append(result, utxo)
		}
	}
	return result
}

func (utxos UTXOs) Sum() (sum int) {
	for _, u := range utxos {
		sum += u.Value
	}
	return
}

type SaveUTXO func(UTXO) error

type GetUTXOS func(publicKeyHash []byte) ([]UTXO, error)
