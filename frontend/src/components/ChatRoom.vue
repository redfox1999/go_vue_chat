<script setup lang="ts">
import { ref, onMounted, nextTick, computed } from 'vue'
import { getUser } from '@/sdk/api'
import type { User } from '@/sdk/types'

const AVATAR_BASE_URL = 'https://api.dicebear.com/9.x/avataaars/svg?seed='

interface Message {
  id: number
  user: string
  avatarUrl: string
  content: string
  timestamp: string
}

interface OnlineUser {
  id: number
  name: string
  avatarUrl: string
  status: 'online' | 'away' | 'busy'
}

const currentChannel = ref('综合聊天')
const channels = ['综合聊天', '技术讨论', '休闲娱乐', '公告']
const messages = ref<Message[]>([])
const inputMessage = ref('')
const chatContainer = ref<HTMLElement | null>(null)

const currentUser = computed<User | null>(() => getUser())

const myUser = computed(() => {
  const user = currentUser.value
  if (user) {
    const name = user.nickname || user.username || 'me'
    return {
      name: name,
      avatarUrl: `${AVATAR_BASE_URL}${encodeURIComponent(name)}`,
      id: user.id
    }
  }
  return {
    name: '我',
    avatarUrl: `${AVATAR_BASE_URL}me`,
    id: 0
  }
})

const onlineUsers = ref<OnlineUser[]>([
  { id: 1, name: '张三', avatarUrl: `${AVATAR_BASE_URL}zhangsan`, status: 'online' },
  { id: 2, name: '李四', avatarUrl: `${AVATAR_BASE_URL}lisi`, status: 'online' },
  { id: 3, name: '王五', avatarUrl: `${AVATAR_BASE_URL}wangwu`, status: 'away' },
  { id: 4, name: '赵六', avatarUrl: `${AVATAR_BASE_URL}zhaoliu`, status: 'online' },
  { id: 5, name: '钱七', avatarUrl: `${AVATAR_BASE_URL}qianqi`, status: 'busy' },
  { id: 6, name: '孙八', avatarUrl: `${AVATAR_BASE_URL}sunba`, status: 'online' },
])

const mockMessages: Message[] = [
  { id: 1, user: '张三', avatarUrl: `${AVATAR_BASE_URL}zhangsan`, content: '大家好！欢迎来到聊天室', timestamp: '10:00' },
  { id: 2, user: '李四', avatarUrl: `${AVATAR_BASE_URL}lisi`, content: '嗨，大家好！', timestamp: '10:01' },
  { id: 3, user: '王五', avatarUrl: `${AVATAR_BASE_URL}wangwu`, content: '今天天气真不错', timestamp: '10:02' },
  { id: 4, user: '赵六', avatarUrl: `${AVATAR_BASE_URL}zhaoliu`, content: '有人想一起讨论技术吗？', timestamp: '10:03' },
  { id: 5, user: '钱七', avatarUrl: `${AVATAR_BASE_URL}qianqi`, content: '我想学 Vue 3', timestamp: '10:04' },
]

const sendMessage = () => {
  if (!inputMessage.value.trim()) return
  
  const newMessage: Message = {
    id: Date.now(),
    user: myUser.value.name,
    avatarUrl: myUser.value.avatarUrl,
    content: inputMessage.value,
    timestamp: new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
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
      avatarUrl: randomUser.avatarUrl,
      content: replies[Math.floor(Math.random() * replies.length)],
      timestamp: new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
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
          在线用户 - {{ onlineUsers.length + 1 }}
        </div>
        <div class="space-y-2">
          <div 
            class="flex items-center gap-3 p-2 rounded-lg bg-indigo-900/30 cursor-pointer transition-colors"
          >
            <div class="relative">
              <img 
                :src="myUser.avatarUrl" 
                :alt="myUser.name"
                class="w-8 h-8 rounded-full bg-gray-700"
              />
              <div class="absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full border-2 border-[#1c1c1c] bg-green-500"></div>
            </div>
            <span class="text-sm font-medium text-indigo-400">{{ myUser.name }} (我)</span>
          </div>
          <div 
            v-for="user in onlineUsers" 
            :key="user.id"
            class="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-800 cursor-pointer transition-colors"
          >
            <div class="relative">
              <img 
                :src="user.avatarUrl" 
                :alt="user.name"
                class="w-8 h-8 rounded-full bg-gray-700"
              />
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
          <img 
            :src="myUser.avatarUrl" 
            :alt="myUser.name"
            class="w-8 h-8 rounded-full bg-gray-700"
          />
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
        class="flex-1 overflow-y-auto px-6 py-4 space-y-4"
      >
        <div
          v-for="msg in messages"
          :key="msg.id"
          :class="[
            'flex gap-4',
            msg.user === myUser.name ? 'flex-row-reverse' : ''
          ]"
        >
          <img 
            :src="msg.avatarUrl" 
            :alt="msg.user"
            class="w-10 h-10 rounded-full flex-shrink-0 bg-gray-700"
          />
          <div :class="['flex flex-col gap-1 max-w-md', msg.user === myUser.name ? 'items-end' : 'items-start']">
            <div :class="['flex items-baseline gap-2', msg.user === myUser.name ? 'flex-row-reverse' : '']">
              <span :class="['font-semibold text-sm', msg.user === myUser.name ? 'text-indigo-400' : 'text-gray-300']">{{ msg.user }}</span>
              <span class="text-xs text-gray-500">{{ msg.timestamp }}</span>
            </div>
            <div 
              :class="[
                'px-4 py-2 rounded-2xl',
                msg.user === myUser.name 
                  ? 'bg-indigo-600 text-white rounded-br-md' 
                  : 'bg-gray-800 text-gray-200 rounded-bl-md'
              ]"
            >
              <p class="text-sm break-words">{{ msg.content }}</p>
            </div>
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
