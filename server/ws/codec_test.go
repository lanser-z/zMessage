package ws

import (
	"testing"

	"zmessage/server/pkg/protocol"
)

func TestEncoder_Decode(t *testing.T) {
	enc := NewEncoder()
	dec := NewDecoder()

	tests := []struct {
		name    string
		message *protocol.WSMessage
	}{
		{
			name: "encode auth message",
			message: &protocol.WSMessage{
				Type: protocol.MsgAuth,
				Seq:  1,
				Payload: []byte("test"),
			},
		},
		{
			name: "encode chat message",
			message: &protocol.WSMessage{
				Type:    protocol.MsgChat,
				Seq:     2,
				Payload: []byte("hello"),
			},
		},
		{
			name: "encode pong message",
			message: &protocol.WSMessage{
				Type: protocol.MsgPong,
				Seq:  0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 编码
			data, err := enc.Encode(tt.message)
			if err != nil {
				t.Fatalf("Encode failed: %v", err)
			}
			if len(data) == 0 {
				t.Fatal("Encoded data is empty")
			}

			// 解码
			decoded, err := dec.Decode(data)
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			// 验证
			if decoded.Type != tt.message.Type {
				t.Errorf("Type mismatch: got %d, want %d", decoded.Type, tt.message.Type)
			}
			if decoded.Seq != tt.message.Seq {
				t.Errorf("Seq mismatch: got %d, want %d", decoded.Seq, tt.message.Seq)
			}
		})
	}
}

func TestEncoder_EncodePayload(t *testing.T) {
	enc := NewEncoder()
	dec := NewDecoder()

	tests := []struct {
		name    string
		payload interface{}
	}{
		{
			name: "encode auth payload",
			payload: &protocol.AuthPayload{
				Token: "test_token_123",
			},
		},
		{
			name: "encode error payload",
			payload: &protocol.ErrorPayload{
				Code:    "invalid_token",
				Message: "token is expired",
			},
		},
		{
			name: "encode chat push payload",
			payload: &protocol.ChatPushPayload{
				MessageID: 123,
				From:       1,
				To:         2,
				Type:       "text",
				Content:    "hello world",
				CreatedAt:  1234567890,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := enc.EncodePayload(tt.payload)
			if err != nil {
				t.Fatalf("EncodePayload failed: %v", err)
			}
			if len(data) == 0 {
				t.Fatal("Encoded data is empty")
			}

			// 解码验证
			var dest interface{}
			switch tt.payload.(type) {
			case *protocol.AuthPayload:
				dest = &protocol.AuthPayload{}
			case *protocol.ErrorPayload:
				dest = &protocol.ErrorPayload{}
			case *protocol.ChatPushPayload:
				dest = &protocol.ChatPushPayload{}
			}

			err = dec.DecodePayload(data, dest)
			if err != nil {
				t.Fatalf("DecodePayload failed: %v", err)
			}
		})
	}
}

func TestDecoder_Decode(t *testing.T) {
	enc := NewEncoder()
	dec := NewDecoder()

	original := &protocol.WSMessage{
		Type:    protocol.MsgAuth,
		Seq:     123,
		Payload: []byte("payload_data"),
	}

	// 先编码
	data, err := enc.Encode(original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// 再解码
	decoded, err := dec.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// 验证
	if decoded.Type != original.Type {
		t.Errorf("Type = %v, want %v", decoded.Type, original.Type)
	}
	if decoded.Seq != original.Seq {
		t.Errorf("Seq = %v, want %v", decoded.Seq, original.Seq)
	}
}

func TestDecoder_DecodePayload(t *testing.T) {
	enc := NewEncoder()
	dec := NewDecoder()

	original := &protocol.AuthPayload{
		Token: "test_token_abc123",
	}

	// 先编码
	data, err := enc.EncodePayload(original)
	if err != nil {
		t.Fatalf("EncodePayload failed: %v", err)
	}

	// 再解码
	var decoded protocol.AuthPayload
	err = dec.DecodePayload(data, &decoded)
	if err != nil {
		t.Fatalf("DecodePayload failed: %v", err)
	}

	// 验证
	if decoded.Token != original.Token {
		t.Errorf("Token = %v, want %v", decoded.Token, original.Token)
	}
}
