package handlers

import (
	"encoding/json"

	"github.com/nebser/crypto-vote/internal/pkg/blockchain"
	"github.com/nebser/crypto-vote/internal/pkg/websocket"
	"github.com/pkg/errors"
)

type blockForgedBody struct {
	Height int              `json:"height"`
	Block  blockchain.Block `json:"block"`
}

func BlockForged(getTip blockchain.GetTipFn, getBlock blockchain.GetBlockFn, verifyBlock blockchain.VerifyBlockFn, addNewBlock blockchain.AddNewBlockFn) websocket.Handler {
	return func(ping websocket.Ping, _ string) (*websocket.Pong, error) {
		var body blockForgedBody
		if err := json.Unmarshal(ping.Body, &body); err != nil {
			return nil, errors.Wrapf(err, "Failed to unmarshal block forged body %s", ping.Body)
		}
		height, err := blockchain.GetHeight(getTip, getBlock)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to get height")
		}
		if height+1 < body.Height {
			return nil, errors.Errorf("Blockchain height is too low %d", height)
		}
		if !verifyBlock(body.Block) {
			return websocket.NewDisconnectPong(), nil
		}
		switch err := addNewBlock(body.Block); {
		case errors.Is(err, blockchain.ErrInvalidBlock):
			return websocket.NewDisconnectPong(), nil
		case err != nil:
			return nil, errors.Wrap(err, "Failed to add new block to blockchain")
		default:
			return websocket.NewNoActionPong(), nil
		}
	}
}
