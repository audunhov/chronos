<script setup lang="ts">
import { ref } from 'vue'
import { api } from '../services/api'
import BButton from './base/BButton.vue'
import BInput from './base/BInput.vue'
import BCard from './base/BCard.vue'

const props = defineProps<{
    organizations?: { id: string, name: string }[]
}>()

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
    console.error('Registration error:', e)
    const errorMsg = e.message || e.payload || 'Failed to register'
    error.value = errorMsg
    if (errorMsg.includes('already a member') || e.status === 409) {
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
  <div class="space-y-8">
    <h2 class="text-6xl font-black uppercase tracking-tighter italic border-b-8 border-blue-400 pb-2 inline-block leading-none">REGISTRER</h2>
    
    <div v-if="magicLinkSent" class="bg-green-100 border-4 border-black p-8 shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] text-center">
        <h4 class="text-3xl font-black uppercase mb-4 italic">SJEKK EPOST!</h4>
        <p class="text-sm font-bold uppercase tracking-widest">Vi har sendt en magisk lenke til <br><span class="text-blue-600">{{ form.email }}</span></p>
        <BButton @click="emit('cancel')" variant="primary" class="mt-8 w-full italic">FORSTÅTT</BButton>
    </div>

    <form v-else @submit.prevent="handleSubmit" class="space-y-6">
      <BInput v-model="form.name" label="Fullt navn" placeholder="NAVNESEN" required />
      
      <BInput v-model="form.email" type="email" label="E-postadresse" placeholder="DIN@EPOST.NO" required />

      <div class="grid grid-cols-2 gap-6">
        <BInput v-model="form.birth_year" type="number" label="Fødselsår" required />
        <BSelect v-model="form.org_id" label="Organisasjon">
            <option value="">Ingen (Global)</option>
            <option v-for="org in props.organizations" :key="org.id" :value="org.id">
                {{ org.name }}
            </option>
        </BSelect>
      </div>
      
      <div v-if="error" class="bg-red-500 text-white border-4 border-black p-4 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
        <p class="font-black italic uppercase">⚠️ {{ error }}</p>
        <div v-if="conflict" class="mt-4 pt-4 border-t-2 border-black space-y-4">
            <p class="text-xs font-black mb-4 uppercase tracking-widest">Du har allerede en profil hos oss!</p>
            <BButton 
                type="button" 
                @click="requestMagicLink"
                variant="secondary"
                class="w-full text-xs"
            >
                SEND MEG EN MAGIC LINK
            </BButton>
        </div>
      </div>

      <div class="mt-12 flex gap-4">
        <BButton type="submit" :loading="loading" variant="primary" class="flex-1 py-6 text-2xl italic">
          FULLFØR
        </BButton>
        <BButton type="button" @click="emit('cancel')" variant="secondary" class="px-8">
          NEI
        </BButton>
      </div>
    </form>
  </div>
</template>
