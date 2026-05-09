# 聊天室应用

基于 Go + Vue 3 的实时聊天室应用，支持 WebSocket 通信和消息持久化。

## 技术栈

### 后端
- **语言**: Go 1.25+
- **框架**: Chi Router
- **数据库**: SQLite 3
- **WebSocket**: gorilla/websocket
- **日志**: Zerolog

### 前端
- **框架**: Vue 3 + TypeScript
- **构建工具**: Vite 8
- **UI 组件**: shadcn-vue + Radix Vue + reka-ui
- **图标**: Lucide Vue
- **样式**: Tailwind CSS 4
- **路由**: Vue Router 5
- **HTTP 客户端**: Axios
- **项目管理**: pnpm

## 快速开始

### 后端启动

```bash
cd backend
make run

```

后端服务默认运行在 `http://localhost:8080`

### 前端启动

```bash
cd frontend
pnpm install
pnpm run dev
```

前端开发服务器默认运行在 `http://localhost:5173`

## 项目结构

```
├── backend/           # Go 后端服务
│   ├── cmd/          # 入口文件
│   ├── config/       # 配置管理
│   ├── dto/          # 数据传输对象
│   ├── handler/      # HTTP 处理器
│   ├── middleware/   # 中间件
│   ├── models/       # 数据库模型
│   ├── repository/   # 数据访问层
│   ├── router/       # 路由配置
│   ├── service/      # 业务逻辑层
│   ├── utils/        # 工具函数
│   └── websocket/    # WebSocket 管理
├── frontend/         # Vue 前端
│   ├── src/
│   │   ├── components/ # UI 组件
│   │   ├── composables/ # 组合式函数
│   │   ├── pages/      # 页面组件
│   │   └── sdk/        # API 封装
└── doc/              # 文档
```

## API 接口

### 用户接口
| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/v1/users/register` | 用户注册 |
| POST | `/api/v1/users/login` | 用户登录 |

### 聊天室接口
| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/v1/chat-rooms` | 获取房间列表 |
| POST | `/api/v1/chat-rooms` | 创建房间 |
| GET | `/api/v1/chat-rooms/:id/users` | 获取房间在线用户 |
| GET | `/api/v1/chat-rooms/:id/messages` | 获取历史消息 |
| POST | `/api/v1/chat-rooms/:id/token` | 获取房间 token |

### WebSocket 消息格式

**客户端发送**：
```json
{
  "action": "join|leave|chat",
  "payload": {
    "room_id": "1",
    "token": "xxx",
    "content": "hello"
  }
}
```

**服务端响应**：
```json
{
  "action": "join|leave|chat|online_users",
  "payload": {
    "user_id": 1,
    "nickname": "user",
    "room_id": "1",
    "message": "hello",
    "send_time": "2024-01-01T00:00:00Z"
  }
}
```

## 功能特性

- ✅ 用户注册/登录
- ✅ 创建/加入聊天室
- ✅ 实时消息推送
- ✅ 历史消息查询
- ✅ 在线用户列表
- ✅ WebSocket 重连机制
- ✅ 消息持久化存储

## 配置说明

后端配置通过环境变量或 `.env` 文件设置：

```env
PORT=8080
DB_PATH=./chat.db
JWT_SECRET=your-secret-key
JWT_EXPIRE=86400
```

## 许可证

MIT License