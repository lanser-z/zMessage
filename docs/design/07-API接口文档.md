# API 接口文档

## 基础信息

- **Base URL**: `https://localhost:8443` (开发环境使用自签名证书)
- **WebSocket URL**: `wss://localhost:8443/ws`
- **认证方式**: Bearer Token (JWT)
- **序列化**: JSON (REST API) / MessagePack (WebSocket)

## 认证接口

### POST /api/auth/register
用户注册

**请求头:**
```
Content-Type: application/json
```

**请求体:**
```json
{
  "username": "alice",
  "password": "password123",
  "nickname": "Alice"
}
```

**参数说明:**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名, 3-20字符 |
| password | string | 是 | 密码, 6-32字符 |
| nickname | string | 否 | 昵称, 默认与用户名相同 |

**响应 (200):**
```json
{
  "user": {
    "id": 1,
    "username": "alice",
    "nickname": "Alice",
    "avatar": null,
    "created_at": 1707600000
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**错误响应:**
- `400 Bad Request`: 参数错误
```json
{
  "error": "USER_INVALID_USERNAME",
  "message": "用户名格式无效"
}
```

- `409 Conflict`: 用户名已存在
```json
{
  "error": "USER_ALREADY_EXISTS",
  "message": "用户名已存在"
}
```

---

### POST /api/auth/login
用户登录

**请求头:**
```
Content-Type: application/json
```

**请求体:**
```json
{
  "username": "alice",
  "password": "password123"
}
```

**响应 (200):**
```json
{
  "user": {
    "id": 1,
    "username": "alice",
    "nickname": "Alice",
    "avatar": null,
    "created_at": 1707600000
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**错误响应:**
- `400 Bad Request`: 参数错误
- `401 Unauthorized`: 用户名或密码错误

---

## 用户接口

### GET /api/users/me
获取当前用户信息

**请求头:**
```
Authorization: Bearer <token>
```

**响应 (200):**
```json
{
  "id": 1,
  "username": "alice",
  "nickname": "Alice",
  "avatar": "/api/media/42",
  "created_at": 1707600000
}
```

---

### GET /api/users
获取用户列表

**请求头:**
```
Authorization: Bearer <token>
```

**查询参数:**
| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| search | string | 否 | - | 搜索关键词 |
| limit | int | 否 | 50 | 返回数量, 最大50 |

**响应 (200):**
```json
{
  "users": [
    {
      "id": 1,
      "username": "alice",
      "nickname": "Alice",
      "avatar": "/api/media/42",
      "online": true,
      "last_seen": 1707600000
    }
  ]
}
```

---

### GET /api/users/:id
获取指定用户信息

**请求头:**
```
Authorization: Bearer <token>
```

**响应 (200):**
```json
{
  "id": 2,
  "username": "bob",
  "nickname": "Bob",
  "avatar": null,
  "online": false,
  "last_seen": 1707590000
}
```

---

### PUT /api/users/me
更新当前用户信息

**请求头:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体:**
```json
{
  "nickname": "New Nickname",
  "avatar_id": 42
}
```

**响应 (200):**
```json
{
  "id": 1,
  "username": "alice",
  "nickname": "New Nickname",
  "avatar": "/api/media/42",
  "created_at": 1707600000
}
```

---

## 会话接口

### GET /api/conversations
获取会话列表

**请求头:**
```
Authorization: Bearer <token>
```

**查询参数:**
| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| limit | int | 否 | 20 | 每页数量 |

**响应 (200):**
```json
{
  "conversations": [
    {
      "id": 1,
      "participant": {
        "id": 2,
        "username": "bob",
        "nickname": "Bob",
        "avatar": null
      },
      "last_message": {
        "id": 123,
        "type": "text",
        "content": "Hello!",
        "sender_id": 2,
        "created_at": 1707600000
      },
      "unread_count": 3,
      "updated_at": 1707600000
    }
  ],
  "total": 10
}
```

---

### GET /api/conversations/:id
获取会话详情

**请求头:**
```
Authorization: Bearer <token>
```

**响应 (200):**
```json
{
  "id": 1,
  "participant": {
    "id": 2,
    "username": "bob",
    "nickname": "Bob",
    "avatar": null
  },
  "created_at": 1707500000,
  "updated_at": 1707600000
}
```

---

### GET /api/conversations/with/:user_id
获取与指定用户的会话

**请求头:**
```
Authorization: Bearer <token>
```

**说明:** 如果会话不存在则自动创建

**响应 (200):**
```json
{
  "id": 1,
  "participant": {
    "id": 2,
    "username": "bob",
    "nickname": "Bob",
    "avatar": null
  },
  "created_at": 1707500000,
  "updated_at": 1707600000
}
```

---

### GET /api/conversations/:id/messages
获取会话消息历史

**请求头:**
```
Authorization: Bearer <token>
```

**查询参数:**
| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| before_id | int | 否 | - | 获取此ID之前的消息 |
| limit | int | 否 | 50 | 返回数量, 最大100 |

**响应 (200):**
```json
{
  "messages": [
    {
      "id": 120,
      "conversation_id": 1,
      "sender_id": 2,
      "receiver_id": 1,
      "type": "text",
      "content": "Hi there!",
      "status": "read",
      "created_at": 1707599000
    }
  ],
  "has_more": true
}
```

---

### POST /api/conversations/:id/read
标记会话消息为已读

**请求头:**
```
Authorization: Bearer <token>
```

**响应 (200):**
```json
{
  "success": true
}
```

---

## 媒体接口

### POST /api/media/upload
上传媒体文件

**请求头:**
```
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

**请求体 (multipart/form-data):**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | File | 是 | 媒体文件 |
| type | string | 是 | 文件类型: image 或 voice |

**文件限制:**
| 类型 | 最大大小 | 允许格式 |
|------|----------|----------|
| image | 5MB | jpg, jpeg, png, gif, webp |
| voice | 10MB | mp3, m4a, ogg, wav |

**响应 (200):**
```json
{
  "id": 42,
  "type": "image",
  "original_url": "/api/media/42",
  "thumbnail_url": "/api/media/42/thumb",
  "size": 1024000,
  "mime_type": "image/jpeg",
  "width": 1920,
  "height": 1080,
  "created_at": 1707600000
}
```

**错误响应:**
- `400 Bad Request`: 文件类型不支持或参数错误
- `413 Payload Too Large`: 文件过大

---

### GET /api/media/:id
获取媒体文件

**请求头:**
```
Authorization: Bearer <token> (可选)
```

**响应 (200):**
```
Content-Type: <mime_type>
Content-Length: <size>

<binary data>
```

---

### GET /api/media/:id/thumb
获取缩略图

**请求头:**
```
Authorization: Bearer <token> (可选)
```

**说明:** 仅图片类型支持缩略图

**响应 (200):**
```
Content-Type: image/jpeg
Content-Length: <size>

<thumbnail binary data>
```

---

### DELETE /api/media/:id
删除媒体文件

**请求头:**
```
Authorization: Bearer <token>
```

**响应 (200):**
```json
{
  "success": true
}
```

---

## WebSocket 协议

### 连接
```
URL: wss://localhost:8443/ws
```

### 消息格式
所有消息使用 MessagePack 编码:

```javascript
{
  "type": int16,    // 消息类型
  "seq": int64,     // 序列号 (客户端生成)
  "payload": binary // 消息负载 (MessagePack编码)
}
```

### 消息类型

#### 客户端 → 服务端

| 类型 | 值 | 说明 |
|------|-----|------|
| MsgAuth | 1 | 认证 |
| MsgChat | 2 | 聊天消息 |
| MsgAck | 3 | 消息确认 |
| MsgSyncReq | 4 | 同步请求 |
| MsgPresence | 5 | 在线状态 |
| MsgPing | 6 | 心跳 |

#### 服务端 → 客户端

| 类型 | 值 | 说明 |
|------|-----|------|
| MsgAuthRsp | 101 | 认证响应 |
| MsgChatPush | 102 | 消息推送 |
| MsgSyncRsp | 103 | 同步响应 |
| MsgPresencePush | 104 | 在线状态推送 |
| MsgPong | 105 | 心跳响应 |
| MsgError | 106 | 错误通知 |

### 认证 (MsgAuth)

**发送:**
```javascript
{
  "type": 1,
  "seq": 1,
  "payload": {
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

**响应 (MsgAuthRsp):**
```javascript
{
  "type": 101,
  "seq": 1,
  "payload": {
    "success": true,
    "user_id": 1
  }
}
```

### 发送消息 (MsgChat)

**发送:**
```javascript
{
  "type": 2,
  "seq": 2,
  "payload": {
    "to": 2,
    "type": "text",
    "content": "Hello!"
  }
}
```

**参数说明:**
| 参数 | 类型 | 说明 |
|------|------|------|
| to | int64 | 接收者用户ID |
| type | string | 消息类型: text/voice/image |
| content | string | 消息内容或媒体ID |

### 消息确认 (MsgAck)

**发送:**
```javascript
{
  "type": 3,
  "seq": 3,
  "payload": {
    "message_id": 123,
    "status": "delivered"
  }
}
```

**参数说明:**
| 参数 | 类型 | 说明 |
|------|------|------|
| message_id | int64 | 消息ID |
| status | string | 状态: delivered/read |

### 同步请求 (MsgSyncReq)

**发送:**
```javascript
{
  "type": 4,
  "seq": 4,
  "payload": {
    "last_message_id": 120,
    "last_sync_time": 0
  }
}
```

**响应 (MsgSyncRsp):**
```javascript
{
  "type": 103,
  "seq": 4,
  "payload": {
    "messages": [
      {
        "id": 121,
        "conversation_id": 1,
        "sender_id": 2,
        "receiver_id": 1,
        "type": "text",
        "content": "Hello!",
        "status": "sent",
        "created_at": 1707600000
      }
    ],
    "has_more": false
  }
}
```

### 在线状态 (MsgPresence)

**发送:**
```javascript
{
  "type": 5,
  "seq": 5,
  "payload": {
    "status": "online"
  }
}
```

**状态值:** online / offline / away

### 心跳 (MsgPing)

**发送:**
```javascript
{
  "type": 6,
  "seq": 6,
  "payload": {}
}
```

**响应 (MsgPong):**
```javascript
{
  "type": 105,
  "seq": 6,
  "payload": {}
}
```

### 消息推送 (MsgChatPush)

**接收:**
```javascript
{
  "type": 102,
  "seq": 0,
  "payload": {
    "id": 121,
    "conversation_id": 1,
    "sender_id": 2,
    "receiver_id": 1,
    "type": "text",
    "content": "Hello!",
    "status": "sent",
    "created_at": 1707600000
  }
}
```

### 在线状态推送 (MsgPresencePush)

**接收:**
```javascript
{
  "type": 104,
  "seq": 0,
  "payload": {
    "user_id": 2,
    "status": "online"
  }
}
```

### 错误通知 (MsgError)

**接收:**
```javascript
{
  "type": 106,
  "seq": 0,
  "payload": {
    "code": "WS_AUTH_FAILED",
    "message": "认证失败"
  }
}
```

## 错误码

### 认证错误
| 错误码 | HTTP | 说明 |
|--------|------|------|
| USER_INVALID_USERNAME | 400 | 用户名格式无效 |
| USER_INVALID_PASSWORD | 400 | 密码格式无效 |
| USER_ALREADY_EXISTS | 409 | 用户名已存在 |
| USER_NOT_FOUND | 404 | 用户不存在 |
| USER_INVALID_PASSWORD | 401 | 密码错误 |
| USER_INVALID_TOKEN | 401 | Token无效或过期 |
| USER_UNAUTHORIZED | 401 | 未认证 |

### 消息错误
| 错误码 | HTTP | 说明 |
|--------|------|------|
| MSG_INVALID_TYPE | 400 | 无效的消息类型 |
| MSG_INVALID_CONTENT | 400 | 无效的消息内容 |
| MSG_CONVERSATION_NOT_FOUND | 404 | 会话不存在 |
| MSG_USER_NOT_FOUND | 404 | 用户不存在 |
| MSG_SEND_TO_SELF | 400 | 不能给自己发消息 |
| MSG_NOT_FOUND | 404 | 消息不存在 |
| MSG_ACCESS_DENIED | 403 | 无权访问 |

### 媒体错误
| 错误码 | HTTP | 说明 |
|--------|------|------|
| MEDIA_INVALID_TYPE | 400 | 不支持的文件类型 |
| MEDIA_INVALID_SIZE | 413 | 文件大小超出限制 |
| MEDIA_INVALID_FORMAT | 400 | 文件格式无效 |
| MEDIA_UPLOAD_FAILED | 500 | 上传失败 |
| MEDIA_NOT_FOUND | 404 | 文件不存在 |
| MEDIA_ACCESS_DENIED | 403 | 无权访问 |

### WebSocket错误
| 错误码 | 说明 |
|--------|------|
| WS_INVALID_MESSAGE | 无效的消息格式 |
| WS_NOT_AUTHENTICATED | 未认证 |
| WS_AUTH_FAILED | 认证失败 |
| WS_INVALID_PAYLOAD | 无效的消息负载 |
| WS_USER_NOT_FOUND | 用户不存在 |
| WS_RATE_LIMITED | 频率限制 |
| WS_INTERNAL_ERROR | 内部错误 |

## 通用错误响应格式

```json
{
  "error": "ERROR_CODE",
  "message": "错误描述信息"
}
```
