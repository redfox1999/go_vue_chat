<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { userApi } from '@/sdk'
import type { User } from '@/sdk/types'
import { useToast } from '@/composables/useToast'

const router = useRouter()
const username = ref('')
const nickname = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)

const { error: showError, success: showSuccess } = useToast()

const handleRegister = async () => {
  if (!username.value || !email.value || !password.value) {
    showError('请填写必填项')
    return
  }

  if (password.value.length < 6) {
    showError('密码至少6位')
    return
  }

  if (password.value !== confirmPassword.value) {
    showError('两次密码不一致')
    return
  }

  loading.value = true

  try {
    await userApi.register({
      username: username.value,
      nickname: nickname.value || username.value,
      email: email.value,
      password: password.value
    })
    
    showSuccess(`注册成功！请登录`)
    setTimeout(() => {
      router.push('/login')
    }, 1500)
  } catch (e) {
    showError((e as Error).message || '注册失败，请重试')
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
              <line x1="19" y1="8" x2="19" y2="14"></line>
              <line x1="22" y1="11" x2="16" y2="11"></line>
            </svg>
          </div>
        </div>
        <CardTitle class="text-xl font-bold text-white">创建账户</CardTitle>
        <p class="text-gray-300 mt-1 text-sm">填写信息完成注册</p>
      </CardHeader>
      
      <CardContent class="space-y-3">
        <div>
          <label class="block text-sm font-medium text-gray-200 mb-1.5">用户名 <span class="text-red-400">*</span></label>
          <Input 
            v-model="username"
            placeholder="用户名"
            class="bg-white/10 border-white/20 text-white placeholder-gray-400 focus:border-indigo-400 focus:ring-indigo-400 h-10"
          />
        </div>
        
        <div>
          <label class="block text-sm font-medium text-gray-200 mb-1.5">昵称</label>
          <Input 
            v-model="nickname"
            placeholder="昵称（可选）"
            class="bg-white/10 border-white/20 text-white placeholder-gray-400 focus:border-indigo-400 focus:ring-indigo-400 h-10"
          />
        </div>
        
        <div>
          <label class="block text-sm font-medium text-gray-200 mb-1.5">邮箱 <span class="text-red-400">*</span></label>
          <Input 
            v-model="email"
            type="email"
            placeholder="邮箱地址"
            class="bg-white/10 border-white/20 text-white placeholder-gray-400 focus:border-indigo-400 focus:ring-indigo-400 h-10"
          />
        </div>
        
        <div>
          <label class="block text-sm font-medium text-gray-200 mb-1.5">密码 <span class="text-red-400">*</span></label>
          <Input 
            v-model="password"
            type="password"
            placeholder="密码（至少6位）"
            class="bg-white/10 border-white/20 text-white placeholder-gray-400 focus:border-indigo-400 focus:ring-indigo-400 h-10"
          />
        </div>
        
        <div>
          <label class="block text-sm font-medium text-gray-200 mb-1.5">确认密码 <span class="text-red-400">*</span></label>
          <Input 
            v-model="confirmPassword"
            type="password"
            placeholder="再次输入密码"
            class="bg-white/10 border-white/20 text-white placeholder-gray-400 focus:border-indigo-400 focus:ring-indigo-400 h-10"
          />
        </div>
        
        <Button 
          class="w-full h-10 bg-gradient-to-r from-indigo-500 to-purple-600 hover:from-indigo-400 hover:to-purple-500 text-white font-medium transition-all shadow-md hover:shadow-indigo-500/25"
          :disabled="loading"
          @click="handleRegister"
        >
          <span v-if="loading" class="flex items-center gap-2">
            <svg class="animate-spin h-4 w-4" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            注册中
          </span>
          <span v-else>注 册</span>
        </Button>
        
        <div class="pt-3 border-t border-white/10 text-center">
          <p class="text-gray-400 text-xs">
            已有账户？
            <router-link to="/login" class="text-indigo-400 hover:text-indigo-300 transition-colors">立即登录</router-link>
          </p>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
