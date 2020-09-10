package repository

import (
	"encoding/base64"
	"encoding/json"

	"github.com/boltdb/bolt"
	"github.com/nebser/crypto-vote/internal/pkg/transaction"
	"github.com/pkg/errors"
)

func transactionsBucket() []byte {
	return []byte("ether")
}

type tx struct {
	ID      string              `json:"id"`
	Inputs  []transactionInput  `json:"inputs"`
	Outputs []transactionOutput `json:"outputs"`
}

func (t tx) toTransaction() transaction.Transaction {
	inputs := transaction.Inputs{}
	for _, in := range t.Inputs {
		inputs = append(inputs, in.toInput())
	}
	outputs := transaction.Outputs{}
	for _, out := range t.Outputs {
		outputs = append(outputs, out.toOutput())
	}
	id, _ := base64.StdEncoding.DecodeString(t.ID)
	return transaction.Transaction{
		ID:      id,
		Inputs:  inputs,
		Outputs: outputs,
	}
}

func newTX(transaction transaction.Transaction) tx {
	inputs := []transactionInput{}
	for _, in := range transaction.Inputs {
		inputs = append(inputs, newTransactionInput(in))
	}
	outputs := []transactionOutput{}
	for _, out := range transaction.Outputs {
		outputs = append(outputs, newTransactionOutput(out))
	}
	return tx{
		ID:      base64.StdEncoding.EncodeToString(transaction.ID),
		Inputs:  inputs,
		Outputs: outputs,
	}
}

type transactionInput struct {
	TransactionID string `json:"transactionId"`
	Vout          int    `json:"vout"`
	PublicKeyHash string `json:"publicKeyHash"`
	Signature     string `json:"signature"`
}

func (ti transactionInput) toInput() transaction.Input {
	transactionID, _ := base64.StdEncoding.DecodeString(ti.TransactionID)
	publicKeyHash, _ := base64.StdEncoding.DecodeString(ti.PublicKeyHash)
	signature, _ := base64.StdEncoding.DecodeString(ti.Signature)
	return transaction.Input{
		TransactionID: transactionID,
		Vout:          ti.Vout,
		PublicKeyHash: publicKeyHash,
		Signature:     signature,
	}
}

func newTransactionInput(input transaction.Input) transactionInput {
	return transactionInput{
		TransactionID: base64.StdEncoding.EncodeToString(input.TransactionID),
		Vout:          input.Vout,
		PublicKeyHash: base64.StdEncoding.EncodeToString(input.PublicKeyHash),
		Signature:     base64.StdEncoding.EncodeToString(input.Signature),
	}
}

type transactionOutput struct {
	Value         int    `json:"value"`
	PublicKeyHash string `json:"publicKeyHash"`
}

func (to transactionOutput) toOutput() transaction.Output {
	publicKeyHash, _ := base64.StdEncoding.DecodeString(to.PublicKeyHash)
	return transaction.Output{
		Value:         to.Value,
		PublicKeyHash: publicKeyHash,
	}
}

func newTransactionOutput(output transaction.Output) transactionOutput {
	return transactionOutput{
		Value:         output.Value,
		PublicKeyHash: base64.StdEncoding.EncodeToString(output.PublicKeyHash),
	}
}

func CastVote(db *bolt.DB) transaction.CastVote {
	return func(from, to, signature []byte) error {
		return db.Update(func(tx *bolt.Tx) error {
			utxos, err := getUTXOs(tx, from)
			switch {
			case err != nil:
				return errors.Wrapf(err, "Failed to retrieve utxos for %x", from)
			case len(utxos) == 0:
				return transaction.ErrInsufficientVotes
			}
			usedUTXO := utxos[0]
			inputs := transaction.Inputs{transaction.Input{
				PublicKeyHash: from,
				Signature:     signature,
				TransactionID: usedUTXO.TransactionID,
				Vout:          usedUTXO.Vout,
			}}
			outputs := transaction.Outputs{
				transaction.Output{
					PublicKeyHash: to,
					Value:         1,
				},
			}
			if usedUTXO.Value > 1 {
				outputs = append(outputs, transaction.Output{
					PublicKeyHash: from,
					Value:         usedUTXO.Value - 1,
				})
			}
			tr, err := transaction.NewTransaction(inputs, outputs)
			if err != nil {
				return errors.Wrap(err, "Failed to create new transaction")
			}
			if err := overwriteUTXOs(tx, usedUTXO.PublicKeyHash, transaction.UTXOs{usedUTXO}); err != nil {
				return errors.Wrap(err, "Failed to overwrite transaction")
			}
			if err := saveTransaction(tx, *tr); err != nil {
				return errors.Wrap(err, "Failed to save transaction")
			}
			if err := saveUTXOs(tx, tr.UTXOs()); err != nil {
				return errors.Wrap(err, "Failed to save UTXOs")
			}
			return nil
		})
	}
}

func saveTransaction(tx *bolt.Tx, transaction transaction.Transaction) error {
	b := tx.Bucket(transactionsBucket())
	if b == nil {
		created, err := tx.CreateBucket(transactionsBucket())
		if err != nil {
			return errors.Wrapf(err, "Failed to create bucket %s", transactionsBucket())
		}
		b = created
	}
	raw, err := json.Marshal(newTX(transaction))
	if err != nil {
		return errors.Wrapf(err, "Failed to serialize transaction %#v", transactionsBucket())
	}
	if err := b.Put(transaction.ID, raw); err != nil {
		return errors.Wrapf(err, "Failed to save transaction %s", transaction)
	}
	return nil
}

func deleteTransaction(tx *bolt.Tx, transaction transaction.Transaction) error {
	b := tx.Bucket(transactionsBucket())
	if b == nil {
		return nil
	}
	if err := b.Delete(transaction.ID); err != nil {
		return errors.Wrapf(err, "Failed to delete transaction %s", transaction)
	}
	return nil
}
