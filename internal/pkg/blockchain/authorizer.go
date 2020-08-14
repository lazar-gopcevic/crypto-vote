package blockchain

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/nebser/crypto-vote/internal/pkg/wallet"
	"github.com/pkg/errors"

	"github.com/nebser/crypto-vote/internal/pkg/websocket"
)

func BlockchainAuthorizer(findBlock FindBlockFn) websocket.Authorizer {
	return func(request websocket.Request) error {
		rawPublicKey, err := base64.StdEncoding.DecodeString(request.Sender)
		if err != nil {
			return websocket.ErrUnauthorized("Invalid public key")
		}
		rawSignature, err := base64.StdEncoding.DecodeString(request.Signature)
		if err != nil {
			return websocket.ErrUnauthorized("Invalid signature")
		}

		if !wallet.Verify(request, rawSignature, rawPublicKey) {
			return websocket.ErrUnauthorized("Signature does not match the payload")
		}

		publicKeyHashed, err := wallet.HashedPublicKey(rawPublicKey)
		if err != nil {
			return err
		}
		criteria := func(b Block) bool {
			if _, ok := b.Body.Transactions.FindTransactionTo(publicKeyHashed); ok {
				return true
			}
			return false
		}
		switch _, ok, err := findBlock(criteria); {
		case err != nil:
			return errors.Errorf("Failed to find block. Error: %s", err)
		case !ok:
			return websocket.ErrUnauthorized(fmt.Sprintf("Node %s does not exist", request.Sender))
		default:
			log.Println("Authorized successfully")
			return nil
		}
	}
}
