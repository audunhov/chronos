<script setup lang="ts">
import { ref } from 'vue'
import { api } from '../services/api'
import RegisterForm from './RegisterForm.vue'
import BCard from './base/BCard.vue'
import BButton from './base/BButton.vue'
import BInput from './base/BInput.vue'

const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')
const showRegister = ref(false)

const handleSubmit = async () => {
  loading.value = true
  error.value = ''
  try {
    await api.login({ email: email.value, password: password.value })
  } catch (e: any) {
    console.error('Login error:', e)
    error.value = e.message || e.payload || (e.status === 'unknown' ? e.payload : null) || 'Invalid credentials'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="max-w-md w-full">
    <div v-if="!showRegister">
        <BCard class="!p-0 overflow-hidden">
            <div class="bg-yellow-400 p-8 border-b-4 border-black">
                <h2 class="text-6xl font-black uppercase tracking-tighter italic leading-none">Login</h2>
            </div>
            
            <form @submit.prevent="handleSubmit" class="p-8 space-y-6">
                <BInput 
                    v-model="email" 
                    type="email" 
                    label="E-post" 
                    placeholder="DIN@EPOST.NO" 
                    required 
                />
                
                <BInput 
                    v-model="password" 
                    type="password" 
                    label="Passord" 
                    placeholder="********" 
                    required 
                />

                <div v-if="error" class="bg-red-500 text-white p-4 border-4 border-black font-black uppercase text-xs italic">
                    ⚠️ {{ error }}
                </div>

                <BButton 
                    type="submit" 
                    :loading="loading" 
                    variant="primary" 
                    class="w-full text-2xl py-4 italic"
                >
                    LOGG INN
                </BButton>
                
                <button 
                    type="button" 
                    @click="showRegister = true" 
                    class="w-full text-center text-[10px] font-black uppercase mt-4 hover:underline tracking-widest"
                >
                    INGEN KONTO? REGISTRER DEG HER
                </button>

                <p class="text-[8px] text-gray-400 text-center uppercase font-black mt-8 tracking-widest">
                    Secure encrypted transmission active
                </p>
            </form>
        </BCard>
    </div>

    <div v-else>
        <BCard>
            <RegisterForm @registered="showRegister = false" @cancel="showRegister = false" />
        </BCard>
    </div>
  </div>
</template>
