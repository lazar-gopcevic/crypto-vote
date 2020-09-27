package websocket

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/nebser/crypto-vote/internal/pkg/transaction"
	"github.com/nebser/crypto-vote/internal/pkg/wallet"
	"github.com/pkg/errors"
)

type Message int

const (
	GetBlockchainHeightMessage Message = iota + 1
	CloseConnectionMessage
	GetMissingBlocksMessage
	GetBlockMessage
	RegisterMessage
	ErrorMessage
	ResponseMessage
	TransactionReceivedMessage
	NoActionMessage
	ForgeBlockMessage
	BlockForgedMessage
	DisconnectMessage
)

func (m Message) String() string {
	switch m {
	case GetBlockchainHeightMessage:
		return "get-blockchain-height"
	case CloseConnectionMessage:
		return "close-connection"
	case GetMissingBlocksMessage:
		return "get-missing-blocks"
	case GetBlockMessage:
		return "get-block"
	case RegisterMessage:
		return "register"
	case ErrorMessage:
		return "error"
	case ResponseMessage:
		return "response"
	case TransactionReceivedMessage:
		return "transaction-received"
	case NoActionMessage:
		return "no-action"
	case ForgeBlockMessage:
		return "forge-block"
	case BlockForgedMessage:
		return "block-forged"
	case DisconnectMessage:
		return "disconnect"
	default:
		return fmt.Sprintf("Unknown message %d", m)
	}
}

type ForgeBlockBody struct {
	Height int `json:"height"`
}

type BlockForgedBody struct {
	Height int         `json:"height"`
	Block  interface{} `json:"block"`
}

type SaveTransactionBody struct {
	Transaction transaction.Transaction `json:"transaction"`
}

type Ping struct {
	Message   Message         `json:"message"`
	Body      json.RawMessage `json:"body"`
	Signature string          `json:"signature,omitempty"`
	Sender    string          `json:"sender,omitempty"`
}

type signablePing struct {
	Body    json.RawMessage `json:"body"`
	Sender  string          `json:"sender,omitempty"`
	Message Message         `json:"message,omitempty"`
}

func (p Ping) Signable() ([]byte, error) {
	s := signablePing{
		Body:    p.Body,
		Message: p.Message,
		Sender:  p.Sender,
	}
	return json.Marshal(s)
}

func (p Ping) Verified() bool {
	senderPKey, err := base64.StdEncoding.DecodeString(p.Sender)
	if err != nil {
		return false
	}
	signature, err := base64.RawStdEncoding.DecodeString(p.Signature)
	if err != nil {
		return false
	}
	return wallet.Verify(p, signature, senderPKey)
}

type Pong struct {
	Message   Message     `json:"message"`
	Body      interface{} `json:"body"`
	Signature string      `json:"signature,omitempty"`
	Sender    string      `json:"sender,omitempty"`
}

type signablePong struct {
	Body    interface{} `json:"body"`
	Sender  string      `json:"sender,omitempty"`
	Message Message     `json:"message"`
}

func (p Pong) Signable() ([]byte, error) {
	s := signablePong{
		Body:    p.Body,
		Message: p.Message,
		Sender:  p.Sender,
	}
	return json.Marshal(s)
}

func (p Pong) Signed(signer wallet.Signer) (Pong, error) {
	p.Sender = signer.Verifier()
	signature, err := signer.Sign(p)
	if err != nil {
		return p, errors.Wrapf(err, "Failed to sign pong %#v", p)
	}
	return Pong{
		Body:      p.Body,
		Message:   p.Message,
		Sender:    p.Sender,
		Signature: signature,
	}, nil
}

func NewErrorPong(e Error) *Pong {
	return &Pong{
		Message: ErrorMessage,
		Body:    e,
	}
}

func NewResponsePong(body interface{}) *Pong {
	return &Pong{
		Message: ResponseMessage,
		Body:    body,
	}
}

func NewNoActionPong() *Pong {
	return &Pong{Message: NoActionMessage}
}

func NewDisconnectPong() *Pong {
	return &Pong{Message: DisconnectMessage}
}
