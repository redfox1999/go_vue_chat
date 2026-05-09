import WebSocket from 'ws';
import axios from 'axios';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

function loadConfig() {
  const configPath = path.join(__dirname, 'config.json');
  return JSON.parse(fs.readFileSync(configPath, 'utf-8'));
}

const botIds = new Set();

class ChatBot {
  constructor(username, password, config) {
    this.config = config;
    this.username = username;
    this.password = password;
    this.token = null;
    this.userId = null;
    this.currentRoom = null;
    this.ws = null;
    this.isRunning = false;
    this.messageQueue = [];
    this.lastMessageTime = 0;
    this.lastReplyTime = 0;
    this.replyInRoom = 0;
    this.maxReplies = 5;
    this.MIN_REPLY_INTERVAL = 10000; // 两次回复至少间隔10秒
    this.roomStartTime = 0;
  }

  log(message, level = 'info') {
    const timestamp = new Date().toISOString();
    const levels = { info: 'INFO', warn: 'WARN', error: 'ERROR' };
    const prefix = this.username ? `[${this.username}]` : '';
    console.log(`${prefix}[${timestamp}] [${levels[level] || 'INFO'}] ${message}`);
  }

  async login() {
    try {
      this.log('正在登录...');
      const response = await axios.post(`${this.config.apiBaseUrl}/users/login`, {
        username: this.username,
        password: this.password
      });
      
      if (!response.data || !response.data.token) {
        this.log(`登录失败: 响应数据格式错误 - ${JSON.stringify(response.data)}`, 'error');
        return false;
      }
      
      this.token = response.data.token;
      this.userId = response.data.user?.id;
      if (this.userId) botIds.add(this.userId);
      this.log(`登录成功，ID=${this.userId}`);
      return true;
    } catch (error) {
      const errorMsg = error.response?.data?.error || error.response?.data?.message || error.message;
      this.log(`登录失败: ${errorMsg}`, 'error');
      return false;
    }
  }

  async getChatRooms() {
    try {
      this.log('获取聊天室列表...');
      const response = await axios.get(`${this.config.apiBaseUrl}/chat-rooms`, {
        headers: { Authorization: `Bearer ${this.token}` }
      });
      const rooms = response.data.data;
      this.log(`获取到 ${rooms.length} 个聊天室`);
      return rooms;
    } catch (error) {
      this.log(`获取聊天室列表失败: ${error.message}`, 'error');
      return [];
    }
  }

  async getRoomToken(roomId) {
    try {
      const response = await axios.get(`${this.config.apiBaseUrl}/chat-rooms/${roomId}/token`, {
        headers: { Authorization: `Bearer ${this.token}` }
      });
      if (!response.data || !response.data.token) {
        this.log(`获取房间 token 失败: 响应数据格式错误`, 'error');
        return null;
      }
      return response.data.token;
    } catch (error) {
      const errorMsg = error.response?.data?.error || error.response?.statusText || error.message;
      this.log(`获取房间 token 失败: ${errorMsg}`, 'error');
      return null;
    }
  }

  connectWebSocket(roomId, token) {
    return new Promise((resolve, reject) => {
      this.log(`正在连接 WebSocket 到房间 ${roomId}...`);
      const wsUrl = `${this.config.wsBaseUrl}?token=${encodeURIComponent(this.token)}`;
      this.ws = new WebSocket(wsUrl);

      this.ws.on('open', () => {
        this.log(`WebSocket 连接成功，房间: ${roomId}`);
        resolve();
      });

      this.ws.on('error', (error) => {
        this.log(`WebSocket 错误: ${error.message}`, 'error');
        reject(error);
      });

      this.ws.on('close', () => {
        this.log(`WebSocket 连接已关闭`);
        this.ws = null;
      });

      this.ws.on('message', (data) => {
        try {
          const messageStr = data.toString().trim();
          const messages = messageStr.split('\n').filter(m => m.trim());
          for (const msg of messages) {
            if (msg.trim()) {
              this.handleMessage(JSON.parse(msg.trim()));
            }
          }
        } catch (error) {
          this.log(`消息解析失败: ${error.message}`, 'warn');
        }
      });
    });
  }

  joinRoom(roomId, token) {
    this.log(`正在加入房间 ${roomId}...`);
    this.ws.send(JSON.stringify({
      action: 'join',
      payload: { room_id: String(roomId), token }
    }));
    this.currentRoom = roomId;
    this.replyInRoom = 0;
    this.lastReplyTime = 0;
    this.roomStartTime = Date.now();
  }

  leaveRoom() {
    if (this.currentRoom && this.ws) {
      this.log(`正在离开房间 ${this.currentRoom}...`);
      this.ws.send(JSON.stringify({
        action: 'leave',
        payload: { room_id: String(this.currentRoom) }
      }));
      this.currentRoom = null;
    }
  }

  handleMessage(data) {
    if (!data.action) return;
    
    if (data.action === 'chat') {
      const payload = data.payload;
      if (!payload) return;
      // 忽略所有机器人发的消息（包括自己）
      if (payload.user_id && botIds.has(Number(payload.user_id))) return;

      const message = payload.content || '';
      if (message) {
        // 冷却检查 + 房间回复次数上限
        const now = Date.now();
        if (this.replyInRoom >= this.maxReplies) {
          this.log(`回复已达上限 (${this.maxReplies})，不再回复`);
          return;
        }
        if (now - this.lastReplyTime < this.MIN_REPLY_INTERVAL) {
          this.log(`冷却中 (距上次 ${Math.round((now - this.lastReplyTime) / 1000)}秒)，跳过回复`);
          return;
        }
        const shouldReply = Math.random() < 0.7;
        if (shouldReply) {
          this.scheduleReply();
        }
      }
    }
  }

  scheduleReply() {
    const delay = Math.random() * (this.config.behavior.replyDelay.max - this.config.behavior.replyDelay.min) + this.config.behavior.replyDelay.min;
    const replyCount = Math.min(
      Math.floor(Math.random() * 2) + 1,
      this.maxReplies - this.replyInRoom
    );
    if (replyCount <= 0) return;

    this.log(`计划在 ${Math.round(delay / 1000)}s 后回复 ${replyCount} 条`);

    setTimeout(() => {
      if (!this.isRunning || !this.ws || this.ws.readyState !== WebSocket.OPEN) return;
      
      const messages = this.config.behavior.replyMessages;
      let messageDelay = 0;
      
      for (let i = 0; i < replyCount; i++) {
        setTimeout(() => {
          if (this.isRunning && this.ws && this.ws.readyState === WebSocket.OPEN) {
            const message = messages[Math.floor(Math.random() * messages.length)];
            this.sendMessage(message, '回复');
          }
        }, messageDelay);
        
        messageDelay += 1500;
      }
    }, delay);
  }

  sendMessage(message, reason = '') {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        action: 'chat',
        payload: {
          room_id: String(this.currentRoom),
          content: message
        }
      }));
      this.replyInRoom++;
      this.lastReplyTime = Date.now();
      this.lastMessageTime = Date.now();
      const reasonStr = reason ? ` [${reason}]` : '';
      this.log(`发送消息${reasonStr}: ${message} (${this.replyInRoom}/${this.maxReplies})`);
    }
  }

  async sendInitialMessages() {
    const count = Math.floor(Math.random() * 2) + 1;
    this.log(`发送 ${count} 条初始消息`);
    
    for (let i = 0; i < count; i++) {
      if (!this.isRunning || !this.ws || this.ws.readyState !== WebSocket.OPEN) break;
      
      const messages = this.config.behavior.randomMessages;
      const message = messages[Math.floor(Math.random() * messages.length)];
      this.sendMessage(message, '进场');
      
      if (i < count - 1) {
        await new Promise(resolve => setTimeout(resolve, 2000));
      }
    }
    
    this.log('初始消息发送完毕');
  }

  async stayInRoom() {
    const stayTime = Math.random() * (this.config.behavior.maxStayTime - this.config.behavior.minStayTime) + this.config.behavior.minStayTime;
    this.log(`将在房间停留 ${Math.round(stayTime)} 分钟`);
    
    return new Promise(resolve => {
      setTimeout(resolve, stayTime * 60 * 1000);
    });
  }

  async enterRoom(room) {
    try {
      this.log(`进入房间: ${room.name} (ID: ${room.id})`);
      const token = await this.getRoomToken(room.id);
      if (!token) {
        this.log('获取房间 token 失败，跳过此房间', 'warn');
        return false;
      }
      
      await this.connectWebSocket(room.id, token);
      this.joinRoom(room.id, token);
      
      return true;
    } catch (error) {
      this.log(`进入房间失败: ${error.message}`, 'error');
      return false;
    }
  }

  async run() {
    this.isRunning = true;
    this.log('机器人启动');
    
    if (!await this.login()) {
      this.log('登录失败，退出', 'error');
      return;
    }

    while (this.isRunning) {
      const rooms = await this.getChatRooms();
      if (rooms.length === 0) {
        this.log('没有可用的聊天室，等待 10 秒后重试...', 'warn');
        await new Promise(resolve => setTimeout(resolve, 10000));
        continue;
      }

      const room = rooms[Math.floor(Math.random() * rooms.length)];
      this.log(`选择房间: ${room.name} (ID: ${room.id})`);

      if (await this.enterRoom(room)) {
        await this.sendInitialMessages();
        await this.stayInRoom();
        
        this.leaveRoom();
        if (this.ws) {
          this.ws.close();
          this.ws = null;
        }
        
        this.log('离开房间，准备进入下一个房间');
        await new Promise(resolve => setTimeout(resolve, 3000));
      } else {
        this.log('进入房间失败，尝试下一个房间', 'warn');
        await new Promise(resolve => setTimeout(resolve, 5000));
      }
    }
  }

  stop() {
    this.log('正在停止...');
    this.isRunning = false;
    this.leaveRoom();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}

// ──── 启动多个机器人 ────
const config = loadConfig();
const botList = (config.bots || []).map(
  botCfg => new ChatBot(botCfg.username, botCfg.password, config)
);

if (botList.length === 0) {
  console.error('config.json 中没有配置机器人 (bots 字段为空)');
  process.exit(1);
}

console.log(`共启动 ${botList.length} 个机器人: ${botList.map(b => b.username).join(', ')}`);

process.on('SIGINT', () => {
  botList.forEach(bot => bot.stop());
  process.exit(0);
});

Promise.allSettled(botList.map(bot => bot.run())).then(() => {
  console.log('所有机器人已退出');
});
