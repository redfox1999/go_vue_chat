import WebSocket from 'ws';
import axios from 'axios';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

class ChatBot {
  constructor() {
    this.config = this.loadConfig();
    this.token = null;
    this.currentRoom = null;
    this.ws = null;
    this.isRunning = false;
    this.messageQueue = [];
    this.lastMessageTime = 0;
    this.roomStartTime = 0;
  }

  loadConfig() {
    const configPath = path.join(__dirname, 'config.json');
    const configData = fs.readFileSync(configPath, 'utf-8');
    return JSON.parse(configData);
  }

  log(message, level = 'info') {
    const timestamp = new Date().toISOString();
    const levels = { info: 'INFO', warn: 'WARN', error: 'ERROR' };
    console.log(`[${timestamp}] [${levels[level] || 'INFO'}] ${message}`);
  }

  async login() {
    try {
      this.log('正在登录...');
      const response = await axios.post(`${this.config.apiBaseUrl}/users/login`, {
        username: this.config.credentials.username,
        password: this.config.credentials.password
      });
      
      if (!response.data || !response.data.token) {
        this.log(`登录失败: 响应数据格式错误 - ${JSON.stringify(response.data)}`, 'error');
        return false;
      }
      
      this.token = response.data.token;
      this.log(`登录成功，用户: ${this.config.credentials.username}`);
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
      this.log(`获取房间 ${roomId} 的 token...`);
      const response = await axios.get(`${this.config.apiBaseUrl}/chat-rooms/${roomId}/token`, {
        headers: { Authorization: `Bearer ${this.token}` }
      });
      if (!response.data || !response.data.token) {
        this.log(`获取房间 token 失败: 响应数据格式错误`, 'error');
        return null;
      }
      this.log(`获取房间 ${roomId} 的 token 成功`);
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
      // 使用用户的 JWT token 来连接，以便后端能识别用户身份
      const wsUrl = `${this.config.wsBaseUrl}?token=${encodeURIComponent(this.token)}`;
      this.log(`WebSocket URL: ${wsUrl}`);
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
          // 处理可能包含多条消息的情况
          const messages = messageStr.split('\n').filter(m => m.trim());
          for (const msg of messages) {
            if (msg.trim()) {
              this.handleMessage(JSON.parse(msg.trim()));
            }
          }
        } catch (error) {
          this.log(`消息解析失败: ${error.message}`, 'warn');
          this.log(`原始消息: ${data.toString()}`, 'warn');
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
    // 显示完整的消息内容用于调试
    this.log(`收到消息: ${JSON.stringify(data)}`);
    
    if (!data.action) {
      this.log(`无效消息: 缺少 action 字段`, 'warn');
      return;
    }
    
    this.log(`消息 action: ${data.action}`);
    
    if (data.action === 'chat') {
      const payload = data.payload;
      if (!payload) {
        this.log(`聊天消息缺少 payload`, 'warn');
        return;
      }
      // 后端发送的字段名是 user_id 和 content
      const userId = payload.user_id || payload.user || '未知ID';
      const nickname = payload.nickname || `用户${userId}`;
      const message = payload.content || payload.message || '空消息';
      this.log(`收到聊天消息: 用户ID=${userId}, 昵称=${nickname}, 内容=${message}`);
      
      // 随机决定是否回复（回复1-2条）
      if (message && message !== '空消息') {
        const shouldReply = Math.random() < 0.9; // 70%概率回复
        if (shouldReply) {
          this.scheduleReply();
        } else {
          this.log('收到消息，本次不回复');
        }
      }
    } else if (data.action === 'online_users') {
      const payload = data.payload;
      const userCount = payload?.users?.length || 0;
      this.log(`在线用户更新: ${userCount} 人`);
    } else if (data.action === 'join_ok') {
      this.log(`加入房间成功: ${data.payload?.room_id}`);
    } else {
      this.log(`未知消息类型: ${data.action}`, 'warn');
    }
  }

  scheduleReply() {
    const delay = Math.random() * (this.config.behavior.replyDelay.max - this.config.behavior.replyDelay.min) + this.config.behavior.replyDelay.min;
    const replyCount = Math.floor(Math.random() * 2) + 1; // 回复1-2条
    this.log(`计划在 ${Math.round(delay / 1000)} 秒后回复 ${replyCount} 条消息`);
    
    setTimeout(() => {
      if (!this.isRunning || !this.ws || this.ws.readyState !== WebSocket.OPEN) return;
      
      const messages = this.config.behavior.replyMessages;
      let messageDelay = 0;
      
      for (let i = 0; i < replyCount; i++) {
        setTimeout(() => {
          if (this.isRunning && this.ws && this.ws.readyState === WebSocket.OPEN) {
            const message = messages[Math.floor(Math.random() * messages.length)];
            this.sendMessage(message);
          }
        }, messageDelay);
        
        messageDelay += 1500; // 每条消息间隔1.5秒
      }
    }, delay);
  }

  sendMessage(message) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        action: 'chat',
        payload: {
          room_id: String(this.currentRoom),
          content: message
        }
      }));
      this.log(`发送消息: ${message}`);
      this.lastMessageTime = Date.now();
    }
  }

  async sendInitialMessages() {
    const count = Math.floor(Math.random() * 2) + 1; // 1-2条消息
    this.log(`进入房间后发送 ${count} 条初始消息`);
    
    for (let i = 0; i < count; i++) {
      if (!this.isRunning || !this.ws || this.ws.readyState !== WebSocket.OPEN) break;
      
      const messages = this.config.behavior.randomMessages;
      const message = messages[Math.floor(Math.random() * messages.length)];
      this.sendMessage(message);
      
      if (i < count - 1) {
        await new Promise(resolve => setTimeout(resolve, 2000)); // 消息间隔2秒
      }
    }
    
    this.log('初始消息发送完毕，等待其他用户消息...');
  }

  async stayInRoom() {
    const stayTime = Math.random() * (this.config.behavior.maxStayTime - this.config.behavior.minStayTime) + this.config.behavior.minStayTime;
    this.log(`将在房间 ${this.currentRoom} 停留 ${Math.round(stayTime)} 分钟`);
    
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
    this.log('聊天室机器人启动');
    
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
        await this.sendInitialMessages(); // 发送1-2条初始消息后停止
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
    this.log('正在停止机器人...');
    this.isRunning = false;
    this.leaveRoom();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.log('机器人已停止');
  }
}

const bot = new ChatBot();

process.on('SIGINT', () => {
  bot.stop();
  process.exit(0);
});

bot.run().catch(error => {
  bot.log(`机器人运行出错: ${error.message}`, 'error');
  process.exit(1);
});
