<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api } from '../services/api'
import BCard from './base/BCard.vue'
import BInput from './base/BInput.vue'
import BButton from './base/BButton.vue'
import BSelect from './base/BSelect.vue'

const props = defineProps<{
    orgs: { id: string, name: string }[]
}>()

const selectedOrg = ref('')
const secrets = ref<any[]>([])
const loading = ref(false)
const showCreate = ref(false)
const newSecret = ref({ name: '', value: '' })

const fetchSecrets = async () => {
    if (!selectedOrg.value) return
    loading.value = true
    try {
        const data = await api.getSecrets(selectedOrg.value)
        secrets.value = Array.isArray(data) ? data : []
    } catch (e) {
        console.error(e)
    } finally {
        loading.value = false
    }
}

const createSecret = async () => {
    try {
        await api.createSecret({
            org_id: selectedOrg.value,
            ...newSecret.value
        })
        showCreate.value = false
        newSecret.value = { name: '', value: '' }
        fetchSecrets()
    } catch (e: any) {
        alert(e.message)
    }
}

watch(selectedOrg, fetchSecrets)

onMounted(() => {
    if (props.orgs && props.orgs.length > 0) {
        selectedOrg.value = props.orgs[0]!.id
    }
})
</script>

<template>
    <div class="space-y-8">
        <header class="border-b-8 border-black pb-4">
            <h2 class="text-4xl font-black uppercase tracking-tighter italic">API Nøkler (Secrets)</h2>
            <p class="text-xs font-bold uppercase text-gray-500">Kryptert lagring av nøkler for Webhooks og eksterne systemer</p>
        </header>

        <div class="flex flex-wrap gap-6 items-end bg-black text-white p-6 shadow-[8px_8px_0px_0px_rgba(0,0,0,0.3)]">
            <BSelect v-model="selectedOrg" label="Velg Organisasjon" class="min-w-[300px]">
                <option v-for="org in orgs" :key="org.id" :value="org.id">{{ org.name }}</option>
            </BSelect>
            <BButton @click="showCreate = true" variant="primary" class="italic text-xs py-2">+ LEGG TIL NØKKEL</BButton>
        </div>

        <div v-if="showCreate" class="animate-in slide-in-from-top-4">
            <BCard class="bg-yellow-50 max-w-2xl mx-auto">
                <h3 class="text-xl font-black uppercase mb-6 italic">Ny Nøkkel</h3>
                <div class="space-y-6">
                    <BInput v-model="newSecret.name" label="Navn på nøkkel" placeholder="f.eks. ZAPIER_WEBHOOK_KEY" required />
                    <BInput v-model="newSecret.value" label="Verdi (Krypteres automatisk)" type="password" required />
                </div>
                <div class="mt-8 flex gap-4">
                    <BButton @click="createSecret" variant="primary" class="flex-1 italic" :disabled="!newSecret.name || !newSecret.value">LAGRE NØKKEL</BButton>
                    <BButton @click="showCreate = false" variant="secondary">AVBRYT</BButton>
                </div>
            </BCard>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <BCard v-for="s in secrets" :key="s.id" class="bg-white group">
                <div class="flex justify-between items-start mb-4">
                    <span class="bg-black text-white text-[10px] font-black px-2 py-1 uppercase">KRYPTERT (AES-256)</span>
                </div>
                <h3 class="font-black text-xl mb-2 uppercase tracking-tighter">{{ s.name }}</h3>
                <p class="text-[10px] font-mono text-gray-500 truncate">ID: {{ s.id }}</p>
            </BCard>
        </div>
        
        <div v-if="secrets.length === 0 && !loading" class="py-20 text-center border-8 border-black border-dashed bg-white">
            <p class="font-black uppercase text-gray-400 italic">Ingen nøkler funnet</p>
        </div>
    </div>
</template>
