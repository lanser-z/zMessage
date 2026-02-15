// SSE 版本：使用 JSON 替代 msgpack
// 由于 SSE 原生支持 JSON，无需额外编解码

// 编码消息（用于 HTTP POST）
export function encodeMessage(type, payload) {
    return JSON.stringify({
        type: type,
        payload: payload
    });
}

// 解码 SSE 消息
export function decodeSSEEvent(event) {
    try {
        return JSON.parse(event.data);
    } catch (error) {
        console.error('Failed to parse SSE event:', error);
        return null;
    }
}
