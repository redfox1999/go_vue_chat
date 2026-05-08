export interface User {
  id: number
  username: string
  nickname: string
  email: string
  birthday?: string
  sign: string
  status: number
  create_at: string
  update_at: string
}

export interface CreateUserRequest {
  username: string
  nickname?: string
  email: string
  password: string
  birthday?: string
  sign?: string
  status?: number
}

export interface UpdateUserRequest {
  username?: string
  nickname?: string
  email?: string
  password?: string
  birthday?: string
  sign?: string
  status?: number
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  user: User
  token: string
}

export interface ChatRoom {
  id: number
  name: string
  logo?: string
  desc: string
  owner_id: number
  group: string
  status: number
  create_at: string
  update_at: string
  client_num?: number
}

export interface CreateChatRoomRequest {
  name: string
  logo?: string
  desc?: string
  owner_id: number
  group?: string
}

export interface UpdateChatRoomRequest {
  name?: string
  logo?: string
  desc?: string
  owner_id?: number
  group?: string
  status?: number
}

export interface Message {
  id: number
  room_id: number
  sender: number
  notify?: string
  message: string
  send_time: string
}

export interface CreateMessageRequest {
  room_id: number
  sender: number
  notify?: string
  message: string
}

export interface PageResponse<T> {
  data: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface Response<T> {
  success: boolean
  message: string
  data?: T
  error?: string
}
