<script setup lang="ts">
import { ref, onMounted, nextTick, computed, watch } from 'vue'
import { getUser, logout, chatRoomApi, messageApi, uploadApi, getToken } from '@/sdk/api'
import type { User, ChatRoom, CreateChatRoomRequest } from '@/sdk/types'
import { useToast } from '@/composables/useToast'
import { useRouter } from 'vue-router'
import { useWebSocket } from '@/composables/useWebSocket'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Card from '@/components/ui/Card.vue'
import CardHeader from '@/components/ui/CardHeader.vue'
import CardTitle from '@/components/ui/CardTitle.vue'
import CardContent from '@/components/ui/CardContent.vue'
import Dialog from '@/components/ui/Dialog.vue'
import Select from '@/components/ui/Select.vue'
import SelectGroup from '@/components/ui/SelectGroup.vue'
import SelectItem from '@/components/ui/SelectItem.vue'

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
  isBot: boolean
}

const currentChannel = ref('')
const channels = ref<ChatRoom[]>([])
const messages = ref<Message[]>([])
const inputMessage = ref('')
const chatContainer = ref<HTMLElement | null>(null)
const isEnteringRoom = ref(false)
const isInRoom = ref(false)

// WebSocket 连接
const getWsUrl = () => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  const token = getToken()
  if (token) {
    return `${protocol}//${host}/ws?token=${token}`
  }
  return `${protocol}//${host}/ws`
}

const { status: wsStatus, connect: connectWs, close: closeWs, send: sendWs, onMessage, onReconnect } = useWebSocket(getWsUrl())

const wsStatusText = computed(() => {
  switch (wsStatus.value) {
    case 'connecting':
      return '连接中...'
    case 'online':
      return '在线'
    case 'disconnected':
      return '已断开'
    default:
      return '未知'
  }
})

const wsStatusClass = computed(() => {
  switch (wsStatus.value) {
    case 'connecting':
      return 'text-yellow-500'
    case 'online':
      return 'text-green-500'
    case 'disconnected':
      return 'text-red-500'
    default:
      return 'text-gray-500'
  }
})

const isSendButtonDisabled = computed(() => {
  const disabled = !isInRoom.value || wsStatus.value !== 'online' || !inputMessage.value.trim()
  if (disabled) {
    console.log('[ChatRoom] Send button disabled - isInRoom:', isInRoom.value, 'wsStatus:', wsStatus.value, 'hasInput:', !!inputMessage.value.trim())
  }
  return disabled
})

const newRoomName = ref('')
const newRoomDesc = ref('')
const newRoomGroup = ref('default')
const newRoomLogo = ref('')
const showCreateDialog = ref(false)
const logoUploading = ref(false)
const logoInputRef = ref<HTMLInputElement | null>(null)

// 表单校验错误状态
const errors = ref({
  name: '',
  desc: '',
  group: ''
})

const currentUser = computed<User | null>(() => getUser())

const { error: showError, success: showSuccess } = useToast()
const router = useRouter()

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

const ROBOT_POOL: OnlineUser[] = [
  { id: -1, name: '张三', avatarUrl: `${AVATAR_BASE_URL}zhangsan`, status: 'online', isBot: true },
  { id: -2, name: '李四', avatarUrl: `${AVATAR_BASE_URL}lisi`, status: 'online', isBot: true },
  { id: -3, name: '王五', avatarUrl: `${AVATAR_BASE_URL}wangwu`, status: 'away', isBot: true },
  { id: -4, name: '赵六', avatarUrl: `${AVATAR_BASE_URL}zhaoliu`, status: 'online', isBot: true },
  { id: -5, name: '钱七', avatarUrl: `${AVATAR_BASE_URL}qianqi`, status: 'busy', isBot: true },
  { id: -6, name: '孙八', avatarUrl: `${AVATAR_BASE_URL}sunba`, status: 'online', isBot: true },
]

const onlineUsers = ref<OnlineUser[]>([])

const mockMessages: Message[] = [
  { id: 1, user: '张三', avatarUrl: `${AVATAR_BASE_URL}zhangsan`, content: '大家好！欢迎来到聊天室', timestamp: '10:00' },
  { id: 2, user: '李四', avatarUrl: `${AVATAR_BASE_URL}lisi`, content: '嗨，大家好！', timestamp: '10:01' },
  { id: 3, user: '王五', avatarUrl: `${AVATAR_BASE_URL}wangwu`, content: '今天天气真不错', timestamp: '10:02' },
  { id: 4, user: '赵六', avatarUrl: `${AVATAR_BASE_URL}zhaoliu`, content: '有人想一起讨论技术吗？', timestamp: '10:03' },
  { id: 5, user: '钱七', avatarUrl: `${AVATAR_BASE_URL}qianqi`, content: '我想学 Vue 3', timestamp: '10:04' },
]

const sendMessage = () => {
  if (!inputMessage.value.trim()) return

  const cur = channels.value.find(c => c.name === currentChannel.value)
  if (!cur) return

  sendWs({
    action: 'chat',
    payload: {
      room_id: String(cur.id),
      content: inputMessage.value,
    },
  })

  inputMessage.value = ''
}

const scrollToBottom = () => {
  if (chatContainer.value) {
    chatContainer.value.scrollTop = chatContainer.value.scrollHeight
  }
}

// 进入房间（包含 leave 操作，用于首次进入或切换房间）
const selectChannel = async (channel: ChatRoom) => {
  // 离开当前房间
  if (currentChannel.value) {
    const prev = channels.value.find(c => c.name === currentChannel.value)
    if (prev) {
      sendWs({ action: 'leave', payload: { room_id: String(prev.id) } })
    }
  }
  
  await enterRoom(channel)
}

// 进入房间核心逻辑（不包含 leave，用于重连后重新进入）
const enterRoom = async (channel: ChatRoom) => {
  console.log('[ChatRoom] enterRoom called with:', channel.name)
  currentChannel.value = channel.name
  messages.value = []
  isEnteringRoom.value = true
  isInRoom.value = false

  // 随机选 3-5 个机器人
  const pickBots = () => {
    const n = 3 + Math.floor(Math.random() * 3)
    const shuffled = [...ROBOT_POOL].sort(() => Math.random() - 0.5)
    return shuffled.slice(0, n)
  }

  // 获取 token 并发起 join
  try {
    const tokenRes = await chatRoomApi.getToken(channel.id)
    sendWs({
      action: 'join',
      payload: { room_id: String(channel.id), token: tokenRes.token },
    })
  } catch {
    // token 获取失败不影响用户列表
  }

  try {
    const result = await chatRoomApi.getUsers(channel.id)
    const realUsers: OnlineUser[] = result.users
      .filter(u => u.user_id !== myUser.value.id)
      .map(u => ({
        id: u.user_id,
        name: u.nickname,
        avatarUrl: `${AVATAR_BASE_URL}${encodeURIComponent(u.nickname)}`,
        status: 'online' as const,
        isBot: false,
      }))

    onlineUsers.value = [...realUsers, ...pickBots()]
  } catch {
    onlineUsers.value = pickBots()
  }

  // 读取房间历史消息（默认50条）
  try {
    const msgResult = await messageApi.getRoomMessages(channel.id, 1, 50)
    if (msgResult.data && msgResult.data.length > 0) {
      const historyMessages: Message[] = msgResult.data.map(m => ({
        id: m.id,
        user: m.nickname,
        avatarUrl: `${AVATAR_BASE_URL}${encodeURIComponent(m.nickname)}`,
        content: m.message,
        timestamp: new Date(m.send_time).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
      }))
      // 反转数组，使最早的消息（id 小的）显示在前面
      messages.value = [...historyMessages].reverse()
    }
  } catch (e) {
    console.error('Failed to load history messages:', e)
    // 如果加载失败，使用 mock 数据
    messages.value = [...mockMessages]
  }

  // 完成进入房间
  isEnteringRoom.value = false
  isInRoom.value = true
  console.log('[ChatRoom] enterRoom completed - isInRoom:', isInRoom.value, 'wsStatus:', wsStatus.value)
  showSuccess(`已进入「${channel.name}」聊天室，祝玩的开心！`, 2000)
  nextTick(() => scrollToBottom())
}

// 重连后重新进入房间
const reenterRoom = async (channel: ChatRoom) => {
  console.log(`[ChatRoom] Re-entering room: ${channel.name}`)
  await enterRoom(channel)
}

const handleLogout = () => {
  // 断开 WebSocket 连接
  closeWs()
  logout()
  router.push('/login')
}

const openCreateDialog = () => {
  newRoomName.value = ''
  newRoomDesc.value = ''
  newRoomGroup.value = 'default'
  newRoomLogo.value = ''
  errors.value = { name: '', desc: '', group: '' }
  showCreateDialog.value = true
}

const closeCreateDialog = () => {
  showCreateDialog.value = false
}

const handleLogoClick = () => {
  if (logoInputRef.value) {
    logoInputRef.value.click()
  }
}

// ------ 图片加载重试 ------
const logoRetryCount = ref(0)
const MAX_LOGO_RETRY = 5
const RETRY_INTERVAL = 300

const handleLogoError = () => {
  if (logoRetryCount.value >= MAX_LOGO_RETRY) {
    showError('Logo 加载失败，请重新上传')
    logoRetryCount.value = 0
    return
  }
  logoRetryCount.value++
  const baseUrl = newRoomLogo.value.split('?')[0]
  const img = new Image()
  img.onload = () => {
    newRoomLogo.value = baseUrl + `?v=${Date.now()}`
  }
  img.onerror = () => {
    setTimeout(handleLogoError, RETRY_INTERVAL)
  }
  img.src = baseUrl + `?v=${Date.now()}`
}

const handleLogoUpload = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  
  if (!file) return

  const validExtensions = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp']
  if (!validExtensions.includes(file.type)) {
    showError('请选择有效的图片文件（JPG、PNG、GIF、WebP）')
    return
  }

  logoUploading.value = true
  
  try {
    const result = await uploadApi.uploadRoomLogo(file)
    logoRetryCount.value = 0
    newRoomLogo.value = result.url
    showSuccess('Logo 上传成功')
  } catch (e) {
    showError((e as Error).message || 'Logo 上传失败')
    newRoomLogo.value = ''
  } finally {
    logoUploading.value = false
    if (target) {
      target.value = ''
    }
  }
}

const handleCreateRoom = async () => {
  // 重置错误状态
  errors.value = { name: '', desc: '', group: '' }
  
  // 表单校验
  let isValid = true
  
  if (!newRoomName.value.trim()) {
    errors.value.name = '请输入聊天室名称'
    isValid = false
  } else if (newRoomName.value.length < 2) {
    errors.value.name = '聊天室名称至少需要2个字符'
    isValid = false
  } else if (newRoomName.value.length > 50) {
    errors.value.name = '聊天室名称不能超过50个字符'
    isValid = false
  }

  if (!isValid) {
    return
  }

  try {
    const request: CreateChatRoomRequest = {
      name: newRoomName.value,
      logo: newRoomLogo.value || undefined,
      desc: newRoomDesc.value || undefined,
      owner_id: myUser.value.id,
      group: newRoomGroup.value || undefined
    }

    const newRoom = await chatRoomApi.create(request)
    
    channels.value.push(newRoom)
    showSuccess('聊天室创建成功')
    closeCreateDialog()
  } catch (e) {
    showError((e as Error).message || '创建聊天室失败')
  }
}

const loadChatRooms = async () => {
  try {
    const result = await chatRoomApi.list(1, 20)
    channels.value = result.data || []
    if (channels.value.length > 0) {
      selectChannel(channels.value[0])
    }
  } catch (e) {
    console.error('Failed to load chat rooms:', e)
    showError('加载聊天室列表失败')
  }
}

const processWsMsg = (data: { action: string; payload: any }) => {
  if (data.action === 'chat') {
    const payload = data.payload as { room_id: string; user_id: number; nickname: string; content: string }
    const cur = channels.value.find(c => c.name === currentChannel.value)
    if (!cur || String(cur.id) !== payload.room_id) return
    const newMessage: Message = {
      id: Date.now(),
      user: payload.nickname,
      avatarUrl: `${AVATAR_BASE_URL}${encodeURIComponent(payload.nickname)}`,
      content: payload.content,
      timestamp: new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
    }
    messages.value.push(newMessage)
    nextTick(() => scrollToBottom())
  } else if (data.action === 'user_join') {
    const payload = data.payload as { room_id: string; user_id: number; nickname: string }
    const cur = channels.value.find(c => c.name === currentChannel.value)
    if (!cur || String(cur.id) !== payload.room_id) return
    if (payload.user_id === myUser.value.id) return
    if (onlineUsers.value.some(u => u.id === payload.user_id && !u.isBot)) return
    onlineUsers.value.push({
      id: payload.user_id,
      name: payload.nickname,
      avatarUrl: `${AVATAR_BASE_URL}${encodeURIComponent(payload.nickname)}`,
      status: 'online',
      isBot: false,
    })
  } else if (data.action === 'user_leave') {
    const payload = data.payload as { room_id: string; user_id: number; nickname: string }
    const cur = channels.value.find(c => c.name === currentChannel.value)
    if (!cur || String(cur.id) !== payload.room_id) return
    onlineUsers.value = onlineUsers.value.filter(u => u.id !== payload.user_id || u.isBot)
  } else if (data.action === 'ping') {
    console.log('Received server ping')
  } else if (data.action === 'pong') {
    console.log('Received pong response')
  }
}

const handleWsMessage = (event: MessageEvent) => {
  const raw = event.data as string
  const lines = raw.split('\n').filter(l => l.trim())
  for (const line of lines) {
    try {
      processWsMsg(JSON.parse(line))
    } catch (e) {
      console.error('Failed to parse WebSocket message:', e)
    }
  }
}

// 监听 WebSocket 状态变化
watch(wsStatus, (newStatus, oldStatus) => {
  console.log(`[ChatRoom] WebSocket status changed: ${oldStatus} -> ${newStatus}`)
  console.log(`[ChatRoom] Current isInRoom: ${isInRoom.value}`)
  // 如果 WebSocket 断开连接，设置未进入房间状态
  if (newStatus === 'disconnected') {
    console.log('[ChatRoom] WebSocket disconnected, setting isInRoom to false')
    isInRoom.value = false
  }
})

onMounted(() => {
  console.log('[ChatRoom] onMounted called')
  
  loadChatRooms()
  
  // 设置消息处理回调
  console.log('[ChatRoom] Setting message callback')
  onMessage(handleWsMessage)
  
  // 设置重连回调
  onReconnect(() => {
    console.log('[ChatRoom] WebSocket reconnected, re-entering room')
    console.log('[ChatRoom] Current channel:', currentChannel.value)
    console.log('[ChatRoom] Channels count:', channels.value.length)
    if (currentChannel.value && channels.value.length > 0) {
      const currentRoom = channels.value.find(c => c.name === currentChannel.value)
      console.log('[ChatRoom] Found current room:', currentRoom)
      if (currentRoom) {
        // 重连成功后重新进入房间：获取新token、join、读取历史记录
        reenterRoom(currentRoom)
        showSuccess('连接已恢复，已重新进入聊天室')
      }
    }
  })
  
  // 连接 WebSocket
  console.log('[ChatRoom] Connecting WebSocket...')
  connectWs()
  
  // 输出初始状态
  console.log('[ChatRoom] Initial WebSocket status:', wsStatus.value)
})
</script>

<template>
  <div class="flex h-screen bg-background text-foreground">
    <div class="w-16 bg-muted flex flex-col items-center py-4 gap-4 border-r border-border">
      <div class="w-12 h-12 rounded-2xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center cursor-pointer hover:from-indigo-400 hover:to-purple-500 transition-all">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
        </svg>
      </div>
      <div class="w-1 h-8 bg-border rounded-full"></div>
      <div 
        v-for="ch in channels.slice(0, 4)" 
        :key="ch.id"
        @click="selectChannel(ch)"
        :class="[
          'w-12 h-12 rounded-2xl flex items-center justify-center cursor-pointer transition-all',
          currentChannel === ch.name 
            ? 'bg-indigo-600 hover:bg-indigo-500 text-white' 
            : 'bg-muted hover:bg-accent text-muted-foreground'
        ]"
      >
        <span class="text-xs font-bold">{{ ch.name.slice(0, 2) }}</span>
      </div>
      <Button
        variant="outline"
        size="icon"
        class="w-12 h-12 rounded-2xl border-dashed border-green-500/50 text-green-500 hover:bg-green-500/10 hover:border-green-500"
        @click="openCreateDialog"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="12" y1="5" x2="12" y2="19"></line>
          <line x1="5" y1="12" x2="19" y2="12"></line>
        </svg>
      </Button>
    </div>

    <div class="w-60 bg-muted flex flex-col border-r border-border h-full">
      <Card class="border-none bg-transparent flex-1 flex flex-col overflow-hidden">
        <CardHeader class="p-4 pb-2 border-b border-border flex-shrink-0">
          <CardTitle class="text-lg">{{ currentChannel }}</CardTitle>
          <p class="text-xs text-muted-foreground mt-1">聊天室</p>
        </CardHeader>
        
        <CardContent class="p-4 pt-2 flex-1 overflow-y-auto">
          <div class="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-3">
            在线用户 - {{ onlineUsers.length + 1 }}
          </div>
          <div class="space-y-2">
            <div 
              class="flex items-center gap-3 p-2 rounded-lg bg-indigo-500/10 cursor-pointer transition-colors"
            >
              <div class="relative">
                <img 
                  :src="myUser.avatarUrl" 
                  :alt="myUser.name"
                  class="w-8 h-8 rounded-full bg-muted"
                />
                <div class="absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full border-2 border-muted bg-green-500"></div>
              </div>
              <span class="text-sm font-medium text-indigo-500">{{ myUser.name }} (我)</span>
            </div>
            <div 
              v-for="user in onlineUsers" 
              :key="user.id"
              class="flex items-center gap-3 p-2 rounded-lg hover:bg-accent cursor-pointer transition-colors"
            >
              <div class="relative">
                <img 
                  :src="user.avatarUrl" 
                  :alt="user.name"
                  class="w-8 h-8 rounded-full bg-muted"
                />
                <div 
                  :class="[
                    'absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full border-2 border-muted',
                    user.status === 'online' ? 'bg-green-500' : user.status === 'away' ? 'bg-yellow-500' : 'bg-red-500'
                  ]"
                ></div>
              </div>
              <span class="text-sm">
                {{ user.name }}
                <span v-if="user.isBot" class="text-[10px] text-muted-foreground ml-1 px-1 rounded bg-muted-foreground/10">bot</span>
              </span>
            </div>
          </div>
        </CardContent>
      </Card>
      
      <CardContent class="p-3 bg-muted border-t border-border flex-shrink-0">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <img 
              :src="myUser.avatarUrl" 
              :alt="myUser.name"
              class="w-8 h-8 rounded-full bg-muted"
            />
            <div>
              <div class="text-sm font-medium">{{ myUser.name }}</div>
              <div :class="['text-xs flex items-center gap-1', wsStatusClass]">
                <div :class="[
                  'w-1.5 h-1.5 rounded-full',
                  wsStatus === 'connecting' ? 'bg-yellow-500 animate-pulse' : '',
                  wsStatus === 'online' ? 'bg-green-500' : '',
                  wsStatus === 'disconnected' ? 'bg-red-500' : '',
                ]"></div>
                {{ wsStatusText }}
              </div>
            </div>
          </div>
          <Button
            variant="destructive"
            size="sm"
            class="text-xs"
            @click="handleLogout"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mr-1" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
              <polyline points="16 17 21 12 16 7"></polyline>
              <line x1="21" y1="12" x2="9" y2="12"></line>
            </svg>
            登出
          </Button>
        </div>
      </CardContent>
    </div>

    <div class="flex-1 flex flex-col">
      <Card class="border-none bg-muted border-b border-border">
        <CardHeader class="p-4">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-indigo-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
              </svg>
              <CardTitle class="text-base">{{ currentChannel }}</CardTitle>
            </div>
            <div class="flex items-center gap-2">
              <div :class="[
                'w-2 h-2 rounded-full',
                wsStatus === 'connecting' ? 'bg-yellow-500 animate-pulse' : '',
                wsStatus === 'online' ? 'bg-green-500' : '',
                wsStatus === 'disconnected' ? 'bg-red-500' : '',
              ]"></div>
              <span :class="['text-xs font-medium', wsStatusClass]">{{ wsStatusText }}</span>
              <span class="text-xs text-gray-400">({{ wsStatus }})</span>
            </div>
          </div>
        </CardHeader>
      </Card>

      <div 
        ref="chatContainer"
        class="flex-1 overflow-y-auto p-4 space-y-4"
      >
        <!-- Loading 状态 -->
        <div v-if="isEnteringRoom" class="flex flex-col items-center justify-center py-12">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-10 h-10 text-indigo-500 animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10" stroke-linecap="round"/>
          </svg>
          <span class="mt-4 text-sm text-muted-foreground">正在进入房间...</span>
        </div>
        
        <div v-else>
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
            class="w-10 h-10 rounded-full flex-shrink-0 bg-muted"
          />
          <div :class="['flex flex-col gap-1 max-w-md', msg.user === myUser.name ? 'items-end' : 'items-start']">
            <div :class="['flex items-baseline gap-2', msg.user === myUser.name ? 'flex-row-reverse' : '']">
              <span :class="['font-semibold text-sm', msg.user === myUser.name ? 'text-indigo-500' : 'text-foreground']">{{ msg.user }}</span>
              <span class="text-xs text-muted-foreground">{{ msg.timestamp }}</span>
            </div>
            <div 
              :class="[
                'px-4 py-2 rounded-2xl overflow-hidden',
                msg.user === myUser.name 
                  ? 'bg-indigo-500 text-white rounded-br-md' 
                  : 'bg-muted text-foreground rounded-bl-md'
              ]"
            >
              <p class="text-sm break-all overflow-hidden">{{ msg.content }}</p>
            </div>
          </div>
        </div>
        </div>
      </div>

      <Card class="border-none bg-muted border-t border-border">
        <CardContent class="p-4">
          <div class="flex gap-3">
            <Input
              v-model="inputMessage"
              @keyup.enter="sendMessage"
              placeholder="发送消息到 {{ currentChannel }}..."
              class="flex-1 bg-background"
            />
            <Button @click="sendMessage" :disabled="isSendButtonDisabled">
              发送
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>

    <Dialog :open="showCreateDialog" @close="closeCreateDialog">
      <div class="p-4 space-y-4">
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold">创建聊天室</h3>
          <button 
            @click="closeCreateDialog"
            class="text-muted-foreground hover:text-foreground transition-colors"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>
        
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-muted-foreground mb-1.5">聊天室名称 *</label>
            <Input
              v-model="newRoomName"
              placeholder="请输入聊天室名称"
              :class="errors.name ? 'border-red-500 focus:ring-red-500 focus:border-red-500' : ''"
            />
            <p v-if="errors.name" class="mt-1 text-xs text-red-500">{{ errors.name }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-muted-foreground mb-1.5">聊天室描述</label>
            <textarea
              v-model="newRoomDesc"
              placeholder="请输入聊天室描述（可选）"
              rows="3"
              class="w-full bg-background border border-input rounded-lg px-3 py-2 text-sm resize-none focus:outline-none focus:ring-2 focus:ring-ring"
            ></textarea>
          </div>
          <div>
            <label class="block text-sm font-medium text-muted-foreground mb-1.5">分组</label>
            <Select v-model="newRoomGroup" class="w-full">
              <SelectGroup>
                <SelectItem value="default">默认分组</SelectItem>
                <SelectItem value="work">工作</SelectItem>
                <SelectItem value="social">社交</SelectItem>
                <SelectItem value="gaming">游戏</SelectItem>
              </SelectGroup>
            </Select>
          </div>
          <div>
            <label class="block text-sm font-medium text-muted-foreground mb-1.5">Logo</label>
            <input
              ref="logoInputRef"
              type="file"
              accept="image/jpeg,image/jpg,image/png,image/gif,image/webp"
              class="hidden"
              @change="handleLogoUpload"
            />
            <div
              class="w-24 h-24 rounded-xl border-2 border-dashed border-input bg-muted cursor-pointer hover:border-accent hover:bg-accent/10 transition-all relative overflow-hidden"
              @click="handleLogoClick"
            >
              <!-- 上传中显示加载动画 -->
              <div v-if="logoUploading" class="absolute inset-0 bg-black/30 flex items-center justify-center z-20">
                <svg xmlns="http://www.w3.org/2000/svg" class="w-8 h-8 text-white animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10" stroke-linecap="round"/>
                </svg>
              </div>
              
              <!-- 显示图片 -->
              <img
                v-if="newRoomLogo"
                :src="newRoomLogo"
                alt="Logo"
                class="w-full h-full object-cover"
                @error="handleLogoError"
              />
              
              <!-- 默认上传提示 -->
              <div v-if="!newRoomLogo" class="w-full h-full flex flex-col items-center justify-center text-muted-foreground">
                <svg xmlns="http://www.w3.org/2000/svg" class="w-8 h-8 mb-1" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                  <polyline points="17 8 12 3 7 8"></polyline>
                  <line x1="12" y1="3" x2="12" y2="15"></line>
                </svg>
                <span class="text-xs">点击上传</span>
              </div>
            </div>
          </div>
        </div>
        
        <div class="flex gap-3 pt-4">
          <Button variant="outline" class="flex-1" @click="closeCreateDialog">
            取消
          </Button>
          <Button class="flex-1" @click="handleCreateRoom">
            创建
          </Button>
        </div>
      </div>
    </Dialog>
  </div>
</template>
