// WebSocket协议消息类型定义
export const MessageType = {
    // 客户端 → 服务端
    MsgAuth: 1,
    MsgChat: 2,
    MsgAck: 3,
    MsgSyncReq: 4,
    MsgPresence: 5,
    MsgPing: 6,

    // 服务端 → 客户端
    MsgAuthRsp: 101,
    MsgChatPush: 102,
    MsgSyncRsp: 103,
    MsgPresencePush: 104,
    MsgPong: 105,
    MsgError: 106
};

// 消息类型名称映射
export const MessageTypeName = {
    1: 'MsgAuth',
    2: 'MsgChat',
    3: 'MsgAck',
    4: 'MsgSyncReq',
    5: 'MsgPresence',
    6: 'MsgPing',
    101: 'MsgAuthRsp',
    102: 'MsgChatPush',
    103: 'MsgSyncRsp',
    104: 'MsgPresencePush',
    105: 'MsgPong',
    106: 'MsgError'
};

// 根据名称获取消息类型
export function getMessageType(name) {
    return MessageType[name] || 0;
}
