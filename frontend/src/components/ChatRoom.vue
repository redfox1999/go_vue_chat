<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'

interface Message {
  id: number
  user: string
  avatar: string
  content: string
  timestamp: string
  color: string
}

interface OnlineUser {
  id: number
  name: string
  avatar: string
  status: 'online' | 'away' | 'busy'
  color: string
}

const currentChannel = ref('综合聊天')
const channels = ['综合聊天', '技术讨论', '休闲娱乐', '公告']
const messages = ref<Message[]>([])
const inputMessage = ref('')
const chatContainer = ref<HTMLElement | null>(null)

const onlineUsers = ref<OnlineUser[]>([
  { id: 1, name: '张三', avatar: '张', status: 'online', color: 'bg-blue-500' },
  { id: 2, name: '李四', avatar: '李', status: 'online', color: 'bg-green-500' },
  { id: 3, name: '王五', avatar: '王', status: 'away', color: 'bg-yellow-500' },
  { id: 4, name: '赵六', avatar: '赵', status: 'online', color: 'bg-purple-500' },
  { id: 5, name: '钱七', avatar: '钱', status: 'busy', color: 'bg-orange-500' },
  { id: 6, name: '孙八', avatar: '孙', status: 'online', color: 'bg-pink-500' },
])

const myUser = {
  name: '我',
  avatar: '我',
  color: 'bg-indigo-600',
}

const mockMessages: Message[] = [
  { id: 1, user: '张三', avatar: '张', content: '大家好！欢迎来到聊天室', timestamp: '10:00', color: 'bg-blue-500' },
  { id: 2, user: '李四', avatar: '李', content: '嗨，大家好！', timestamp: '10:01', color: 'bg-green-500' },
  { id: 3, user: '王五', avatar: '王', content: '今天天气真不错', timestamp: '10:02', color: 'bg-yellow-500' },
  { id: 4, user: '赵六', avatar: '赵', content: '有人想一起讨论技术吗？', timestamp: '10:03', color: 'bg-purple-500' },
  { id: 5, user: '钱七', avatar: '钱', content: '我想学 Vue 3', timestamp: '10:04', color: 'bg-orange-500' },
]

const sendMessage = () => {
  if (!inputMessage.value.trim()) return
  
  const newMessage: Message = {
    id: Date.now(),
    user: myUser.name,
    avatar: myUser.avatar,
    content: inputMessage.value,
    timestamp: new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
    color: myUser.color,
  }
  
  messages.value.push(newMessage)
  inputMessage.value = ''
  
  nextTick(() => scrollToBottom())
  
  const replies = [
    '收到！',
    '好的',
    '👍',
    '哈哈',
    '有意思',
  ]
  
  setTimeout(() => {
    const randomUser = onlineUsers.value[Math.floor(Math.random() * onlineUsers.value.length)]
    const replyMessage: Message = {
      id: Date.now() + 1,
      user: randomUser.name,
      avatar: randomUser.avatar,
      content: replies[Math.floor(Math.random() * replies.length)],
      timestamp: new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
      color: randomUser.color,
    }
    messages.value.push(replyMessage)
    nextTick(() => scrollToBottom())
  }, 800 + Math.random() * 1200)
}

const scrollToBottom = () => {
  if (chatContainer.value) {
    chatContainer.value.scrollTop = chatContainer.value.scrollHeight
  }
}

const selectChannel = (channel: string) => {
  currentChannel.value = channel
  messages.value = []
  setTimeout(() => {
    messages.value = [...mockMessages]
    nextTick(() => scrollToBottom())
  }, 100)
}

onMounted(() => {
  messages.value = [...mockMessages]
  setTimeout(scrollToBottom, 100)
})
</script>

<template>
  <div class="flex h-screen bg-[#0f0f0f] text-white">
    <div class="w-16 bg-[#1c1c1c] flex flex-col items-center py-4 gap-4 border-r border-gray-800">
      <div class="w-12 h-12 rounded-2xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center cursor-pointer hover:from-indigo-400 hover:to-purple-500 transition-all">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
        </svg>
      </div>
      <div class="w-1 h-8 bg-gray-700 rounded-full"></div>
      <div 
        v-for="ch in channels.slice(0, 4)" 
        :key="ch"
        @click="selectChannel(ch)"
        :class="[
          'w-12 h-12 rounded-2xl flex items-center justify-center cursor-pointer transition-all',
          currentChannel === ch 
            ? 'bg-indigo-600 hover:bg-indigo-500' 
            : 'bg-gray-700 hover:bg-gray-600'
        ]"
      >
        <span class="text-xs font-bold">{{ ch.slice(0, 2) }}</span>
      </div>
    </div>

    <div class="w-60 bg-[#1c1c1c] border-r border-gray-800 flex flex-col">
      <div class="p-4 border-b border-gray-800">
        <h2 class="font-bold text-lg">{{ currentChannel }}</h2>
        <p class="text-xs text-gray-500 mt-1">聊天室</p>
      </div>
      
      <div class="flex-1 overflow-y-auto p-4">
        <div class="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-3">
          在线用户 - {{ onlineUsers.length }}
        </div>
        <div class="space-y-2">
          <div 
            v-for="user in onlineUsers" 
            :key="user.id"
            class="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-800 cursor-pointer transition-colors"
          >
            <div class="relative">
              <div :class="['w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold', user.color]">
                {{ user.avatar }}
              </div>
              <div 
                :class="[
                  'absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full border-2 border-[#1c1c1c]',
                  user.status === 'online' ? 'bg-green-500' : user.status === 'away' ? 'bg-yellow-500' : 'bg-red-500'
                ]"
              ></div>
            </div>
            <span class="text-sm">{{ user.name }}</span>
          </div>
        </div>
      </div>

      <div class="p-3 bg-[#252525] border-t border-gray-800">
        <div class="flex items-center gap-3">
          <div :class="['w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold', myUser.color]">
            {{ myUser.avatar }}
          </div>
          <div>
            <div class="text-sm font-medium">{{ myUser.name }}</div>
            <div class="text-xs text-green-500">在线</div>
          </div>
        </div>
      </div>
    </div>

    <div class="flex-1 flex flex-col">
      <div class="bg-[#1c1c1c] border-b border-gray-800 px-6 py-4">
        <div class="flex items-center gap-2">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-indigo-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
          </svg>
          <h3 class="font-semibold">{{ currentChannel }}</h3>
        </div>
      </div>

      <div 
        ref="chatContainer"
        class="flex-1 overflow-y-auto px-6 py-4 space-y-1"
      >
        <div
          v-for="msg in messages"
          :key="msg.id"
          class="flex gap-4 p-2 rounded-lg hover:bg-gray-900/50 group"
        >
          <div :class="['w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0 font-bold', msg.color]">
            {{ msg.avatar }}
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-baseline gap-2 mb-1">
              <span class="font-semibold text-sm">{{ msg.user }}</span>
              <span class="text-xs text-gray-500">{{ msg.timestamp }}</span>
            </div>
            <p class="text-gray-200 text-sm break-words">{{ msg.content }}</p>
          </div>
        </div>
      </div>

      <div class="p-4 bg-[#1c1c1c] border-t border-gray-800">
        <div class="flex gap-3">
          <input
            v-model="inputMessage"
            @keyup.enter="sendMessage"
            type="text"
            placeholder="发送消息到 {{ currentChannel }}..."
            class="flex-1 bg-[#2a2a2a] border border-gray-700 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500/50 transition-all"
          />
          <button
            @click="sendMessage"
            class="bg-indigo-600 hover:bg-indigo-500 text-white px-6 py-3 rounded-lg font-medium transition-colors"
          >
            发送
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
