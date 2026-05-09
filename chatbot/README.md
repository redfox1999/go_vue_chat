# 聊天室机器人

一个基于 Node.js 的聊天室机器人，可以自动登录、进入聊天室、发送消息和回复其他用户。

## 功能特性

- ✅ 自动登录系统
- ✅ 自动获取聊天室列表
- ✅ 随机选择聊天室进入
- ✅ 定期发送随机消息
- ✅ 回复其他用户的消息
- ✅ 自动切换聊天室
- ✅ 详细的控制台日志输出

## 安装

```bash
cd chatbot
npm install
```

## 配置

编辑 `config.json` 文件：

```json
{
  "apiBaseUrl": "http://localhost:8080/api/v1",
  "wsBaseUrl": "ws://localhost:8080/ws",
  "credentials": {
    "username": "your_username",
    "password": "your_password"
  },
  "behavior": {
    "randomMessages": [
      "大家好！",
      "今天天气不错"
    ],
    "replyMessages": [
      "你说得对",
      "确实如此"
    ],
    "minStayTime": 3,
    "maxStayTime": 8,
    "minRandomInterval": 30,
    "maxRandomInterval": 60,
    "replyDelay": {
      "min": 1000,
      "max": 3000
    }
  }
}
```

### 配置说明

| 配置项 | 说明 |
|--------|------|
| `apiBaseUrl` | 后端 API 地址 |
| `wsBaseUrl` | WebSocket 地址 |
| `credentials.username` | 登录用户名 |
| `credentials.password` | 登录密码 |
| `randomMessages` | 随机发送的消息列表 |
| `replyMessages` | 回复消息列表 |
| `minStayTime` | 在房间最少停留时间（分钟） |
| `maxStayTime` | 在房间最多停留时间（分钟） |
| `minRandomInterval` | 随机消息最小间隔（秒） |
| `maxRandomInterval` | 随机消息最大间隔（秒） |
| `replyDelay.min` | 回复消息最小延迟（毫秒） |
| `replyDelay.max` | 回复消息最大延迟（毫秒） |

## 运行

```bash
npm start
```

## 日志输出示例

```
[2024-01-01T00:00:00.000Z] [INFO] 聊天室机器人启动
[2024-01-01T00:00:01.000Z] [INFO] 正在登录...
[2024-01-01T00:00:01.500Z] [INFO] 登录成功，用户: bot_user
[2024-01-01T00:00:01.600Z] [INFO] 获取聊天室列表...
[2024-01-01T00:00:01.800Z] [INFO] 获取到 3 个聊天室
[2024-01-01T00:00:01.900Z] [INFO] 选择房间: 聊天室1 (ID: 1)
[2024-01-01T00:00:02.000Z] [INFO] 获取房间 1 的 token...
[2024-01-01T00:00:02.200Z] [INFO] 获取房间 1 的 token 成功
[2024-01-01T00:00:02.300Z] [INFO] 正在连接 WebSocket 到房间 1...
[2024-01-01T00:00:02.500Z] [INFO] WebSocket 连接成功，房间: 1
[2024-01-01T00:00:02.600Z] [INFO] 正在加入房间 1...
[2024-01-01T00:00:02.700Z] [INFO] 将在房间 1 停留 5 分钟
[2024-01-01T00:00:30.000Z] [INFO] 下次随机消息将在 45 秒后发送
[2024-01-01T00:01:15.000Z] [INFO] 发送消息: 大家好！
[2024-01-01T00:02:00.000Z] [INFO] 收到消息: action=chat
[2024-01-01T00:02:00.100Z] [INFO] 收到聊天消息: 用户=张三, 内容=你好
[2024-01-01T00:02:00.200Z] [INFO] 计划在 2 秒后回复
[2024-01-01T00:02:02.300Z] [INFO] 发送消息: 你说得对
```

## 停止

按 `Ctrl+C` 停止机器人。

## 注意事项

1. 确保后端服务正在运行
2. 确保配置的用户名和密码正确
3. 确保有至少一个可用的聊天室
4. 机器人会无限循环，直到手动停止
