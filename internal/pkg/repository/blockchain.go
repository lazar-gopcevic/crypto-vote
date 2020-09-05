package repository

import (
	"encoding/json"

	"github.com/boltdb/bolt"
	"github.com/nebser/crypto-vote/internal/pkg/blockchain"
	"github.com/pkg/errors"
)

func blocksBucket() []byte {
	return []byte("blocks")
}

func tipKey() []byte {
	return []byte("l")
}

func GetTip(db *bolt.DB) blockchain.GetTipFn {
	return func() []byte {
		var tip []byte
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(blocksBucket())
			if b == nil {
				return nil
			}
			tip = b.Get(tipKey())
			return nil
		})
		return tip
	}
}

func InitBlockchain(db *bolt.DB) blockchain.InitBlockchainFn {
	return func(genesis blockchain.Block) ([]byte, error) {
		var tip []byte
		err := db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucket(blocksBucket())
			if err != nil {
				return errors.Wrap(err, "Failed to create blocks bucket")
			}
			rawBlock, err := json.Marshal(newBlock(genesis))
			if err != nil {
				return errors.Wrapf(err, "Failed to marshal block %#v", genesis)
			}
			if err := b.Put(genesis.Header.Hash, rawBlock); err != nil {
				return errors.Wrap(err, "Failed to put genesis block")
			}
			if err := b.Put(tipKey(), genesis.Header.Hash); err != nil {
				return errors.Wrap(err, "Failed to update tip")
			}
			tip = genesis.Header.Hash
			return nil
		})

		return tip, err
	}
}

func addBlock(tx *bolt.Tx, block blockchain.Block) ([]byte, error) {
	b := tx.Bucket(blocksBucket())
	if b == nil {
		created, err := tx.CreateBucket(blocksBucket())
		if err != nil {
			return nil, errors.Wrap(err, "Failed to create blocks bucket")
		}
		b = created
	}
	rawBlock, err := json.Marshal(newBlock(block))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to marshal block %#v", block)
	}
	if err := b.Put(block.Header.Hash, rawBlock); err != nil {
		return nil, errors.Wrapf(err, "Failed to put block %#v", block)
	}
	if err := b.Put(tipKey(), block.Header.Hash); err != nil {
		return nil, errors.Wrap(err, "Failed to update tip")
	}
	return block.Header.Hash, nil
}

func AddBlock(db *bolt.DB) blockchain.AddBlockFn {
	return func(block blockchain.Block) ([]byte, error) {
		var tip []byte
		err := db.Update(func(tx *bolt.Tx) error {
			created, err := addBlock(tx, block)
			if err != nil {
				return err
			}
			tip = created
			for _, transaction := range block.Body.Transactions {
				if err := deleteTransaction(tx, transaction); err != nil {
					return err
				}
				if err := saveUTXOS(tx, transaction.UTXOs()); err != nil {
					return err
				}
			}
			return nil
		})
		return tip, err
	}
}

func GetBlock(db *bolt.DB) blockchain.GetBlockFn {
	return func(hash []byte) (*blockchain.Block, error) {
		var result *blockchain.Block
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(blocksBucket())
			if b == nil {
				return errors.New("Blocks bucket does not exist")
			}
			rawBlock := b.Get(hash)
			if rawBlock == nil {
				return nil
			}
			var serialized block
			if err := json.Unmarshal(rawBlock, &serialized); err != nil {
				return errors.Wrapf(err, "Failed to unmarshal serialized block %s", rawBlock)
			}
			bl := serialized.toBlock()
			result = &bl
			return nil
		})
		return result, err
	}
}
