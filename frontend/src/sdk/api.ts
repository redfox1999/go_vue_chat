import axios, { type AxiosInstance, type AxiosResponse, type AxiosError } from 'axios'
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
  PageResponse,
  RoomUser
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

export function logout(): void {
  clearToken()
  clearUser()
}

export function isLoggedIn(): boolean {
  return getToken() !== null
}

const axiosInstance: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  headers: {
    'Content-Type': 'application/json'
  },
  timeout: 10000
})

axiosInstance.interceptors.request.use(
  (config) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

axiosInstance.interceptors.response.use(
  (response: AxiosResponse) => {
    return response
  },
  (error: AxiosError) => {
    const data = error.response?.data as Record<string, unknown> | string | undefined
    let errorMessage = '请求失败'

    if (data && typeof data === 'object' && 'error' in data) {
      errorMessage = String(data.error)
    } else if (data && typeof data === 'object' && 'message' in data) {
      errorMessage = String(data.message)
    } else if (data && typeof data === 'object' && 'detail' in data) {
      errorMessage = String(data.detail)
    } else if (typeof data === 'string' && data.length > 0) {
      errorMessage = data
    } else if (error.message) {
      errorMessage = error.message
    }

    return Promise.reject(new Error(errorMessage))
  }
)

async function request<T>(
  url: string,
  options: {
    method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
    data?: unknown
    headers?: Record<string, string>
    params?: Record<string, unknown>
  } = {}
): Promise<T> {
  const { method = 'GET', data, headers, params } = options
  
  const response = await axiosInstance.request<T>({
    url,
    method,
    data,
    headers,
    params
  })
  
  return response.data
}

export const userApi = {
  async register(data: CreateUserRequest): Promise<User> {
    return request<User>('/users/register', {
      method: 'POST',
      data
    })
  },

  async login(data: LoginRequest): Promise<LoginResponse> {
    const result = await request<LoginResponse>('/users/login', {
      method: 'POST',
      data
    })
    if (result.token) {
      setToken(result.token)
      setUser(result.user)
    }
    return result
  },

  async getById(id: number): Promise<User> {
    return request<User>('/users/get', {
      params: { id }
    })
  },

  async getCurrent(): Promise<User> {
    return request<User>('/users/me')
  },

  async update(id: number, data: UpdateUserRequest): Promise<User> {
    return request<User>('/users', {
      method: 'PUT',
      params: { id },
      data
    })
  },

  async delete(id: number): Promise<void> {
    return request<void>('/users', {
      method: 'DELETE',
      params: { id }
    })
  },

  async list(page: number = 1, pageSize: number = 10): Promise<PageResponse<User>> {
    return request<PageResponse<User>>('/users', {
      params: { page, page_size: pageSize }
    })
  }
}

export const chatRoomApi = {
  async create(data: CreateChatRoomRequest): Promise<ChatRoom> {
    return request<ChatRoom>('/chat-rooms', {
      method: 'POST',
      data
    })
  },

  async getById(id: number): Promise<ChatRoom> {
    return request<ChatRoom>(`/chat-rooms/${id}`)
  },

  async update(id: number, data: UpdateChatRoomRequest): Promise<ChatRoom> {
    return request<ChatRoom>(`/chat-rooms/${id}`, {
      method: 'PUT',
      data
    })
  },

  async delete(id: number): Promise<void> {
    return request<void>(`/chat-rooms/${id}`, {
      method: 'DELETE'
    })
  },

  async list(page: number = 1, pageSize: number = 10): Promise<PageResponse<ChatRoom>> {
    return request<PageResponse<ChatRoom>>('/chat-rooms', {
      params: { page, page_size: pageSize }
    })
  },

  async listByGroup(group: string, page: number = 1, pageSize: number = 10): Promise<PageResponse<ChatRoom>> {
    return request<PageResponse<ChatRoom>>(`/chat-rooms/group/${group}`, {
      params: { page, page_size: pageSize }
    })
  },

  async listByOwner(ownerId: number, page: number = 1, pageSize: number = 10): Promise<PageResponse<ChatRoom>> {
    return request<PageResponse<ChatRoom>>(`/chat-rooms/owner/${ownerId}`, {
      params: { page, page_size: pageSize }
    })
  },

  async getUsers(roomId: number): Promise<{ room_id: string; users: RoomUser[] }> {
    return request<{ room_id: string; users: RoomUser[] }>(`/chat-rooms/${roomId}/users`)
  },

  async getToken(roomId: number): Promise<{ room_id: string; token: string }> {
    return request<{ room_id: string; token: string }>(`/chat-rooms/${roomId}/token`)
  }
}

export const messageApi = {
  async create(data: CreateMessageRequest): Promise<Message> {
    return request<Message>('/messages', {
      method: 'POST',
      data
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

  async getRoomMessages(roomId: number, page: number = 1, pageSize: number = 50): Promise<PageResponse<Message>> {
    return request<PageResponse<Message>>(`/chat-rooms/${roomId}/messages`, {
      params: { page, page_size: pageSize }
    })
  },

  async listBySender(senderId: number, page: number = 1, pageSize: number = 50): Promise<PageResponse<Message>> {
    return request<PageResponse<Message>>(`/messages/sender/${senderId}`, {
      params: { page, page_size: pageSize }
    })
  }
}

export const uploadApi = {
  async uploadFile(file: File): Promise<{ url: string }> {
    const formData = new FormData()
    formData.append('file', file)
    
    const response = await axiosInstance.post<{ url: string }>('/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
    
    return response.data
  },

  async uploadRoomLogo(file: File): Promise<{ url: string }> {
    const formData = new FormData()
    formData.append('file', file)
    
    const response = await axiosInstance.post<{ url: string }>('/upload/room-logo', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
    
    return response.data
  }
}
