// MessagePack 编解码封装
import * as msgpack from 'https://cdn.jsdelivr.net/npm/msgpack-lite@0.1.26/dist/msgpack.min.js';

export const codec = {
    encode(data) {
        return msgpack.encode(data);
    },

    decode(buffer) {
        return msgpack.decode(new Uint8Array(buffer));
    }
};

// 编码WebSocket消息
export function encodeMessage(type, seq, payload) {
    return msgpack.encode({
        type: type,
        seq: seq,
        payload: codec.encode(payload)
    });
}

// 解码WebSocket消息
export function decodeMessage(buffer) {
    const data = msgpack.decode(new Uint8Array(buffer));
    return {
        type: data.type,
        seq: data.seq,
        payload: codec.decode(data.payload)
    };
}
