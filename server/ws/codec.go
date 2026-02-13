package ws

import (
	"bytes"
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	"zmessage/server/pkg/protocol"
)

// Encoder MessagePack编码器
type Encoder struct {
	encoder *msgpack.Encoder
	buf     *bytes.Buffer
}

// NewEncoder 创建编码器
func NewEncoder() *Encoder {
	buf := &bytes.Buffer{}
	return &Encoder{
		encoder: msgpack.NewEncoder(buf),
		buf:     buf,
	}
}

// Encode 编码完整的WebSocket消息
func (e *Encoder) Encode(msg *protocol.WSMessage) ([]byte, error) {
	e.buf.Reset()
	if err := e.encoder.Encode(msg); err != nil {
		return nil, fmt.Errorf("encode message: %w", err)
	}
	return e.buf.Bytes(), nil
}

// EncodePayload 编码负载数据
func (e *Encoder) EncodePayload(payload interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := msgpack.NewEncoder(&buf).Encode(payload); err != nil {
		return nil, fmt.Errorf("encode payload: %w", err)
	}
	return buf.Bytes(), nil
}

// Decoder MessagePack解码器
type Decoder struct{}

// NewDecoder 创建解码器
func NewDecoder() *Decoder {
	return &Decoder{}
}

// Decode 解码WebSocket消息
func (d *Decoder) Decode(data []byte) (*protocol.WSMessage, error) {
	var msg protocol.WSMessage
	if err := msgpack.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("decode message: %w", err)
	}
	return &msg, nil
}

// DecodePayload 解码负载数据
func (d *Decoder) DecodePayload(data []byte, dest interface{}) error {
	if err := msgpack.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}
	return nil
}
