<template>
  <AppLayout>
    <div class="mx-auto flex w-full max-w-5xl flex-col gap-6 px-6 py-6">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <div class="text-sm text-gray-500 dark:text-dark-400">API Key #{{ apiKeyId }}</div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">聊天记录</h1>
          <div class="mt-2 text-sm text-gray-500 dark:text-dark-400">
            显示该 API Key 最新 {{ limit }} 条主消息
          </div>
        </div>
        <div class="flex items-center gap-3">
          <button class="btn btn-secondary" @click="goBack">返回 API Keys</button>
          <button class="btn btn-secondary" :disabled="loading" @click="loadMessages">刷新</button>
        </div>
      </div>

      <div v-if="loading" class="rounded-2xl border border-gray-200 bg-white p-6 text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-400">
        正在加载聊天内容...
      </div>

      <div
        v-for="message in messages"
        :key="message.id"
        :class="[
          'rounded-2xl border p-4 shadow-sm',
          message.direction === 'inbound'
            ? 'mr-12 border-blue-200 bg-blue-50 dark:border-blue-900/50 dark:bg-blue-950/20'
            : 'ml-12 border-emerald-200 bg-emerald-50 dark:border-emerald-900/50 dark:bg-emerald-950/20'
        ]"
      >
        <div class="mb-2 flex items-center justify-between gap-3 text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
          <span>{{ message.direction === 'inbound' ? 'User' : 'Assistant' }} / {{ message.role }}</span>
          <span>{{ formatDateTime(message.created_at) }}</span>
        </div>
        <pre class="whitespace-pre-wrap break-words font-sans text-sm text-gray-900 dark:text-white">{{ message.content_text || JSON.stringify(message.content_json || {}, null, 2) }}</pre>
      </div>

      <div v-if="!loading && messages.length === 0" class="rounded-2xl border border-gray-200 bg-white p-6 text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-400">
        当前 API Key 暂无可展示的聊天记录。
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import type { ChatMessage } from '@/types'
import { keysAPI } from '@/api'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()

const apiKeyId = computed(() => Number(route.params.id))
const loading = ref(false)
const messages = ref<ChatMessage[]>([])
const limit = 50

async function loadMessages() {
  loading.value = true
  try {
    messages.value = await keysAPI.listRecentChatMessages(apiKeyId.value, limit)
  } catch (error: any) {
    appStore.showError(error?.message || '加载聊天记录失败')
  } finally {
    loading.value = false
  }
}

function goBack() {
  router.push({ name: 'Keys' })
}

onMounted(() => {
  loadMessages()
})
</script>
