<script setup lang="ts">
import { ref } from 'vue'
import { api } from '../services/api'

const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

const handleSubmit = async () => {
  loading.value = true
  error.value = ''
  try {
    await api.login({ email: email.value, password: password.value })
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="brutalist-card bg-white">
    <h2 class="text-4xl font-black uppercase tracking-tighter mb-8 italic border-b-8 border-yellow-400 pb-2 inline-block">Login</h2>
    <form @submit.prevent="handleSubmit" class="space-y-6">
      <div class="flex flex-col">
        <label for="email" class="brutalist-label">E-post</label>
        <input 
            type="email" 
            id="email" 
            v-model="email" 
            required 
            placeholder="DIN@EPOST.NO"
            class="brutalist-input" 
        />
      </div>
      <div class="flex flex-col">
        <label for="password" class="brutalist-label">Passord</label>
        <input 
            type="password" 
            id="password" 
            v-model="password" 
            required 
            placeholder="********"
            class="brutalist-input" 
        />
      </div>
      <div v-if="error" class="bg-red-500 text-white p-3 border-4 border-black font-bold uppercase text-xs">
        ⚠️ {{ error }}
      </div>
      <button 
        type="submit" 
        :disabled="loading" 
        class="brutalist-btn-primary w-full text-xl py-4"
      >
        {{ loading ? 'SJEKKER...' : 'LOGG INN' }}
      </button>
      
      <p class="text-[10px] text-gray-400 text-center uppercase font-black mt-8">Secure encrypted transmission active</p>
    </form>
  </div>
</template>
