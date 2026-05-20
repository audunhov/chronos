<script setup lang="ts">
import { ref } from 'vue'
import { api } from '../services/api'

const emit = defineEmits(['registered', 'cancel'])

const form = ref({
  name: '',
  email: '',
  org_id: '',
  birth_year: 2000,
  metadata: {}
})

const loading = ref(false)
const error = ref('')
const conflict = ref(false)
const magicLinkSent = ref(false)

const handleSubmit = async () => {
  loading.value = true
  error.value = ''
  conflict.value = false
  try {
    const payload = {
      ...form.value,
      metadata: { 
        ...form.value.metadata, 
        birth_year: form.value.birth_year 
      }
    }
    await api.registerMember(payload)
    emit('registered')
  } catch (e: any) {
    error.value = e.message
    if (e.message.includes('already a member') || e.status === 409) {
        conflict.value = true
    }
  } finally {
    loading.value = false
  }
}

const requestMagicLink = async () => {
    loading.value = true
    try {
        await api.requestMagicLink(form.value.email)
        magicLinkSent.value = true
    } catch (e: any) {
        error.value = e.message
    } finally {
        loading.value = false
    }
}
</script>

<template>
  <div>
    <h2 class="text-4xl font-black uppercase tracking-tighter mb-8 italic border-b-8 border-blue-400 pb-2 inline-block">REGISTRER</h2>
    
    <div v-if="magicLinkSent" class="bg-green-100 border-4 border-black p-6 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
        <h4 class="text-xl font-black uppercase mb-2">SJEKK EPOST!</h4>
        <p class="text-sm font-bold">Vi har sendt en magisk lenke til {{ form.email }}.</p>
        <button @click="emit('cancel')" class="mt-6 brutalist-btn-primary w-full">FORSTÅTT</button>
    </div>

    <form v-else @submit.prevent="handleSubmit" class="space-y-6">
      <div class="flex flex-col">
        <label for="name" class="brutalist-label">Fullt navn</label>
        <input type="text" id="name" v-model="form.name" required class="brutalist-input" placeholder="NAVNESEN" />
      </div>
      <div class="flex flex-col">
        <label for="email" class="brutalist-label">E-postadresse</label>
        <input type="email" id="email" v-model="form.email" required class="brutalist-input" placeholder="DIN@EPOST.NO" />
      </div>
      <div class="grid grid-cols-2 gap-4">
          <div class="flex flex-col">
            <label for="birth_year" class="brutalist-label">Fødselsår</label>
            <input type="number" id="birth_year" v-model="form.birth_year" required class="brutalist-input" />
          </div>
          <div class="flex flex-col">
            <label for="org_id" class="brutalist-label">Org-ID</label>
            <input type="text" id="org_id" v-model="form.org_id" placeholder="ID" class="brutalist-input" />
          </div>
      </div>
      
      <div v-if="error" class="bg-red-500 text-white border-4 border-black p-4 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
        <p class="font-black">⚠️ {{ error }}</p>
        <div v-if="conflict" class="mt-4 pt-4 border-t-2 border-black">
            <p class="text-xs font-bold mb-4 uppercase">Du har allerede en profil hos oss!</p>
            <button 
                type="button" 
                @click="requestMagicLink"
                class="brutalist-btn bg-white text-black text-xs w-full py-2 hover:bg-yellow-400"
            >
                SEND MEG EN MAGIC LINK
            </button>
        </div>
      </div>

      <div class="mt-8 flex gap-4">
        <button type="submit" :disabled="loading" class="brutalist-btn-primary flex-1 py-4 text-xl">
          {{ loading ? 'LAGRER...' : 'FULLFØR' }}
        </button>
        <button type="button" @click="emit('cancel')" class="brutalist-btn-secondary px-6">
          NEI
        </button>
      </div>
    </form>
  </div>
</template>
