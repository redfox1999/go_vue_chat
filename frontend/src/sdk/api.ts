import type {
  User,
  CreateUserRequest,
  UpdateUserRequest,
  LoginRequest,
  LoginResponse,
  ChatRoom,
  CreateChatRoomRequest,
  UpdateChatRoomRequest,
  Message,
  CreateMessageRequest,
  PageResponse
} from './types'

const BASE_URL = import.meta.env.VITE_API_URL || '/api/v1'

const tokenStorageKey = 'chat_token'
const userStorageKey = 'chat_user'

export function getToken(): string | null {
  return localStorage.getItem(tokenStorageKey)
}

export function setToken(token: string): void {
  localStorage.setItem(tokenStorageKey, token)
}

export function clearToken(): void {
  localStorage.removeItem(tokenStorageKey)
}

export function getUser(): User | null {
  const userStr = localStorage.getItem(userStorageKey)
  return userStr ? JSON.parse(userStr) : null
}

export function setUser(user: User): void {
  localStorage.setItem(userStorageKey, JSON.stringify(user))
}

export function clearUser(): void {
  localStorage.removeItem(userStorageKey)
}

async function request<T>(
  url: string,
  options: RequestInit = {}
): Promise<T> {
  const token = getToken()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>)
  }

  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const response = await fetch(`${BASE_URL}${url}`, {
    ...options,
    headers
  })

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: 'Unknown error' }))
    throw new Error(errorData.error || `HTTP error! status: ${response.status}`)
  }

  if (response.status === 204) {
    return null as T
  }

  return response.json()
}

export const userApi = {
  async register(data: CreateUserRequest): Promise<User> {
    return request<User>('/users', {
      method: 'POST',
      body: JSON.stringify(data)
    })
  },

  async login(data: LoginRequest): Promise<LoginResponse> {
    const result = await request<LoginResponse>('/users/login', {
      method: 'POST',
      body: JSON.stringify(data)
    })
    if (result.token) {
      setToken(result.token)
      setUser(result.user)
    }
    return result
  },

  async getById(id: number): Promise<User> {
    return request<User>(`/users/get?id=${id}`)
  },

  async getCurrent(): Promise<User> {
    return request<User>('/users/me')
  },

  async update(id: number, data: UpdateUserRequest): Promise<User> {
    return request<User>(`/users/get?id=${id}`, {
      method: 'PUT',
      body: JSON.stringify(data)
    })
  },

  async delete(id: number): Promise<void> {
    return request<void>(`/users/get?id=${id}`, {
      method: 'DELETE'
    })
  },

  async list(page: number = 1, pageSize: number = 10): Promise<PageResponse<User>> {
    return request<PageResponse<User>>(`/users?page=${page}&page_size=${pageSize}`)
  }
}

export const chatRoomApi = {
  async create(data: CreateChatRoomRequest): Promise<ChatRoom> {
    return request<ChatRoom>('/chat-rooms', {
      method: 'POST',
      body: JSON.stringify(data)
    })
  },

  async getById(id: number): Promise<ChatRoom> {
    return request<ChatRoom>(`/chat-rooms/${id}`)
  },

  async update(id: number, data: UpdateChatRoomRequest): Promise<ChatRoom> {
    return request<ChatRoom>(`/chat-rooms/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data)
    })
  },

  async delete(id: number): Promise<void> {
    return request<void>(`/chat-rooms/${id}`, {
      method: 'DELETE'
    })
  },

  async list(page: number = 1, pageSize: number = 10): Promise<PageResponse<ChatRoom>> {
    return request<PageResponse<ChatRoom>>(`/chat-rooms?page=${page}&page_size=${pageSize}`)
  },

  async listByGroup(group: string, page: number = 1, pageSize: number = 10): Promise<PageResponse<ChatRoom>> {
    return request<PageResponse<ChatRoom>>(`/chat-rooms/group/${group}?page=${page}&page_size=${pageSize}`)
  },

  async listByOwner(ownerId: number, page: number = 1, pageSize: number = 10): Promise<PageResponse<ChatRoom>> {
    return request<PageResponse<ChatRoom>>(`/chat-rooms/owner/${ownerId}?page=${page}&page_size=${pageSize}`)
  }
}

export const messageApi = {
  async create(data: CreateMessageRequest): Promise<Message> {
    return request<Message>('/messages', {
      method: 'POST',
      body: JSON.stringify(data)
    })
  },

  async getById(id: number): Promise<Message> {
    return request<Message>(`/messages/${id}`)
  },

  async delete(id: number): Promise<void> {
    return request<void>(`/messages/${id}`, {
      method: 'DELETE'
    })
  },

  async listByRoom(roomId: number, page: number = 1, pageSize: number = 50): Promise<PageResponse<Message>> {
    return request<PageResponse<Message>>(`/messages/room/${roomId}?page=${page}&page_size=${pageSize}`)
  },

  async listBySender(senderId: number, page: number = 1, pageSize: number = 50): Promise<PageResponse<Message>> {
    return request<PageResponse<Message>>(`/messages/sender/${senderId}?page=${page}&page_size=${pageSize}`)
  }
}
