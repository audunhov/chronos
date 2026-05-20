<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api } from '../services/api'
import BButton from './base/BButton.vue'
import BCard from './base/BCard.vue'
import BInput from './base/BInput.vue'
import BSelect from './base/BSelect.vue'

const props = defineProps<{
    orgs: { id: string, name: string }[]
}>()

const selectedOrg = ref('')
const organs = ref<any[]>([])
const loading = ref(false)
const showCreate = ref(false)

const newOrgan = ref({
    name: '',
    parent_organ_id: null as string | null
})

const fetchOrgans = async () => {
    if (!selectedOrg.value) return
    loading.value = true
    try {
        const data = await api.getOrgans(selectedOrg.value)
        organs.value = Array.isArray(data) ? data : []
    } catch (e) {
        console.error(e)
        organs.value = []
    } finally {
        loading.value = false
    }
}

const createOrgan = async () => {
    try {
        await api.createOrgan({
            org_id: selectedOrg.value,
            ...newOrgan.value
        })
        showCreate.value = false
        newOrgan.value = { name: '', parent_organ_id: null }
        fetchOrgans()
    } catch (e: any) {
        alert(e.message)
    }
}

watch(selectedOrg, fetchOrgans)

onMounted(() => {
    if (props.orgs.length > 0) {
        selectedOrg.value = props.orgs[0].id
    }
})
</script>

<template>
    <div class="space-y-8">
        <header class="border-b-8 border-black pb-4">
            <h2 class="text-4xl font-black uppercase tracking-tighter italic">Organer</h2>
            <p class="text-xs font-bold uppercase text-gray-500">Administrer styrer, komiteer og andre underledd</p>
        </header>

        <div class="flex flex-wrap gap-6 items-end bg-black text-white p-6 shadow-[8px_8px_0px_0px_rgba(0,0,0,0.3)]">
            <BSelect v-model="selectedOrg" label="Velg Organisasjon" class="bg-black text-white border-white min-w-[300px]">
                <option v-for="org in orgs" :key="org.id" :value="org.id">{{ org.name }}</option>
            </BSelect>
            <BButton @click="showCreate = true" variant="primary" class="italic text-xs py-2">+ NYTT ORGAN</BButton>
        </div>

        <div v-if="showCreate">
            <BCard class="bg-yellow-50 max-w-2xl mx-auto">
                <h3 class="text-xl font-black uppercase mb-6 italic">Opprett nytt organ</h3>
                <div class="space-y-6">
                    <BInput v-model="newOrgan.name" label="Navn på Organ" placeholder="F.eks. Lokallagsstyre" required />
                    
                    <BSelect v-model="newOrgan.parent_organ_id" label="Overordnet Organ (Valgfritt)">
                        <option :value="null">Ingen (Toppnivå organ)</option>
                        <option v-for="o in organs" :key="o.id" :value="o.id">{{ o.name }}</option>
                    </BSelect>
                </div>
                <div class="mt-8 flex gap-4">
                    <BButton @click="createOrgan" variant="primary" class="flex-1 italic">LAGRE ORGAN</BButton>
                    <BButton @click="showCreate = false" variant="secondary">AVBRYT</BButton>
                </div>
            </BCard>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-8">
            <BCard v-for="o in organs" :key="o.id" class="hover:bg-blue-50 transition-colors flex flex-col justify-between">
                <div>
                    <span class="text-[8px] font-black uppercase text-gray-400 tracking-widest mb-1 block">ID: {{ o.id.split('-')[0] }}</span>
                    <h3 class="text-2xl font-black uppercase tracking-tighter leading-none mb-4">{{ o.name }}</h3>
                </div>
                <div class="pt-4 border-t-2 border-black border-dashed flex justify-between items-center">
                    <span class="text-[10px] font-bold uppercase">Medlemmer: 0</span>
                    <BButton variant="ghost" class="text-[10px] py-1 px-2">ADMINISTRER</BButton>
                </div>
            </BCard>

            <div v-if="organs.length === 0 && !loading" class="col-span-full py-20 text-center border-8 border-black border-dashed">
                <p class="font-black uppercase text-gray-400 italic">Ingen organer registrert for denne enheten</p>
            </div>
        </div>
    </div>
</template>
