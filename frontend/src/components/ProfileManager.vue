<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../services/api'
import { auth } from '../services/auth'
import BButton from './base/BButton.vue'
import BCard from './base/BCard.vue'
import BInput from './base/BInput.vue'

const profile = ref({
    name: '',
    email: ''
})

const loading = ref(false)
const message = ref('')
const error = ref('')

const fetchProfile = async () => {
    loading.value = true
    try {
        const data = await api.getMyProfile()
        profile.value = {
            name: data.name || '',
            email: data.email || ''
        }
    } catch (e: any) {
        error.value = e.message
    } finally {
        loading.value = false
    }
}

const updateProfile = async () => {
    loading.value = true
    message.value = ''
    error.value = ''
    try {
        await api.updateMyProfile(profile.value)
        message.value = 'Profil oppdatert!'
        // Oppdater auth state
        if (auth.user) {
            (auth.user as any).name = profile.value.name
            auth.user.email = profile.value.email
        }
    } catch (e: any) {
        error.value = e.message
    } finally {
        loading.value = false
    }
}

onMounted(fetchProfile)
</script>

<template>
    <div class="max-w-2xl mx-auto space-y-8">
        <header class="border-b-8 border-black pb-4">
            <h2 class="text-4xl font-black uppercase tracking-tighter italic">Personalia</h2>
            <p class="text-xs font-bold uppercase text-gray-500">Administrer din identitet i Chronos</p>
        </header>

        <BCard class="p-8 space-y-6">
            <BInput 
                v-model="profile.name" 
                label="Fullt Navn" 
                placeholder="NAVNESEN" 
                required 
            />
            
            <BInput 
                v-model="profile.email" 
                type="email" 
                label="E-postadresse" 
                placeholder="DIN@EPOST.NO" 
                required 
            />

            <div v-if="message" class="bg-green-400 border-4 border-black p-4 font-black uppercase text-xs italic">
                ✓ {{ message }}
            </div>
            
            <div v-if="error" class="bg-red-500 text-white p-4 border-4 border-black font-black uppercase text-xs italic">
                ⚠️ {{ error }}
            </div>

            <BButton 
                @click="updateProfile" 
                :loading="loading" 
                variant="primary" 
                class="w-full text-2xl py-4 italic"
            >
                LAGRE ENDRINGER
            </BButton>
        </BCard>

        <BCard class="bg-orange-50 border-dashed">
            <h4 class="font-black uppercase text-sm mb-2 italic underline">Sikkerhet & Data</h4>
            <p class="text-xs font-bold leading-relaxed">
                Chronos benytter kryptering for alle sensitive data. Dine opplysninger er kun tilgjengelige for administratorer i de organisasjonene du er medlem av.
            </p>
        </BCard>
    </div>
</template>
