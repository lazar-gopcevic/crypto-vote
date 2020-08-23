package operations

import (
	"encoding/base64"

	"github.com/gorilla/websocket"
	"github.com/nebser/crypto-vote/internal/pkg/wallet"
	_websocket "github.com/nebser/crypto-vote/internal/pkg/websocket"
	"github.com/pkg/errors"
)

type RegisterFn func(nodeID string) ([]string, error)

type registerPayload struct {
	NodeID string `json:"nodeId"`
}

type registerResult struct {
	Nodes []string `json:"nodes"`
}

func Register(conn *websocket.Conn, w wallet.Wallet) RegisterFn {
	return func(nodeID string) ([]string, error) {
		payload := operation{
			Message: _websocket.RegisterMessage,
			Body: registerPayload{
				NodeID: nodeID,
			},
			Sender: base64.StdEncoding.EncodeToString(w.PublicKey),
		}
		rawSignature, err := wallet.Sign(payload, w.PrivateKey)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to sign payload")
		}
		payload.Signature = base64.StdEncoding.EncodeToString(rawSignature)
		var r registerResult
		if err := call(conn, payload, &r); err != nil {
			return nil, errors.Wrapf(err, "Failed to send operation %#v", payload)
		}
		return r.Nodes, nil
	}
}
