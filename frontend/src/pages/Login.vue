<script setup lang="ts">
import { ref } from 'vue'
import { userApi } from '@/sdk'
import type { LoginResponse } from '@/sdk/types'
import { useToast } from '@/composables/useToast'
import { useRouter } from 'vue-router'

const username = ref('')
const password = ref('')
const loading = ref(false)

const { error: showError, success: showSuccess } = useToast()
const router = useRouter()

const handleLogin = async () => {
  if (!username.value || !password.value) {
    showError('请输入用户名和密码')
    return
  }

  loading.value = true

  try {
    const result: LoginResponse = await userApi.login({
      username: username.value,
      password: password.value
    })
    
    if (result.token) {
      showSuccess(`登录成功！欢迎, ${result.user.nickname}`)
      router.push('/chat')
    }
  } catch (e) {
    showError((e as Error).message || '登录失败，请重试')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-900 via-purple-900 to-pink-800 p-4">
    <Card class="w-full max-w-sm shadow-xl border-none bg-white/10 backdrop-blur-lg">
      <CardHeader class="text-center pb-4">
        <div class="mb-3 flex justify-center">
          <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"></path>
              <circle cx="9" cy="7" r="4"></circle>
              <path d="M22 21v-2a4 4 0 0 0-3-3.87"></path>
              <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
            </svg>
          </div>
        </div>
        <CardTitle class="text-xl font-bold text-white">欢迎回来</CardTitle>
        <p class="text-gray-300 mt-1 text-sm">请登录您的账户</p>
      </CardHeader>
      
      <CardContent class="space-y-4">
        <div class="space-y-3">
          <div>
            <label class="block text-sm font-medium text-gray-200 mb-1.5">用户名</label>
            <Input 
              v-model="username"
              placeholder="用户名"
              class="bg-white/10 border-white/20 text-white placeholder-gray-400 focus:border-indigo-400 focus:ring-indigo-400 h-10"
            />
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-200 mb-1.5">密码</label>
            <Input 
              v-model="password"
              type="password"
              placeholder="密码"
              class="bg-white/10 border-white/20 text-white placeholder-gray-400 focus:border-indigo-400 focus:ring-indigo-400 h-10"
            />
          </div>
        </div>
        
        <Button 
          class="w-full h-10 bg-gradient-to-r from-indigo-500 to-purple-600 hover:from-indigo-400 hover:to-purple-500 text-white font-medium transition-all shadow-md hover:shadow-indigo-500/25"
          :disabled="loading"
          @click="handleLogin"
        >
          <span v-if="loading" class="flex items-center gap-2">
            <svg class="animate-spin h-4 w-4" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            登录中
          </span>
          <span v-else>登 录</span>
        </Button>
        
        <div class="flex items-center justify-between text-xs text-gray-400">
          <label class="flex items-center gap-1.5 cursor-pointer">
            <input type="checkbox" class="rounded border-gray-500 bg-white/10 text-indigo-500 focus:ring-indigo-400 w-3.5 h-3.5" />
            <span>记住我</span>
          </label>
          <a href="#" class="hover:text-white transition-colors">忘记密码？</a>
        </div>
        
        <div class="pt-3 border-t border-white/10 text-center">
          <p class="text-gray-400 text-xs">
            还没有账户？
            <a href="/register" @click.prevent="router.push('/register')" class="text-indigo-400 hover:text-indigo-300 transition-colors">立即注册</a>
          </p>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
