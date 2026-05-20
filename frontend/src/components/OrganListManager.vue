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

const selectedOrgan = ref<any | null>(null)
const organMembers = ref<any[]>([])
const membersLoading = ref(false)
const showAddMember = ref(false)

const newOrgan = ref({
    name: '',
    parent_organ_id: ''
})

const allMembers = ref<any[]>([]) // For å velge hvem som skal legges til
const newMemberID = ref('')
const newMemberRole = ref('Styremedlem')

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

const fetchOrganMembers = async (organ: any) => {
    selectedOrgan.value = organ
    membersLoading.value = true
    try {
        organMembers.value = await api.getOrganMembers(organ.id)
        // Hent alle medlemmer i org-en for å kunne legge til i organ
        const membersData = await api.getMembers(selectedOrg.value)
        allMembers.value = Array.isArray(membersData) ? membersData : []
    } catch (e) {
        console.error(e)
    } finally {
        membersLoading.value = false
    }
}

const addOrganMember = async () => {
    if (!selectedOrgan.value || !newMemberID.value) return
    try {
        await api.assignOrganMember({
            organ_id: selectedOrgan.value.id,
            user_id: newMemberID.value,
            role_type: newMemberRole.value
        })
        newMemberID.value = ''
        showAddMember.value = false
        fetchOrganMembers(selectedOrgan.value)
        fetchOrgans() // For å oppdatere teller
    } catch (e: any) {
        alert(e.message)
    }
}

const removeOrganMember = async (id: string) => {
    if (!confirm('Fjern dette medlemmet fra organet?')) return
    try {
        await api.revokeOrganMember(id)
        fetchOrganMembers(selectedOrgan.value)
        fetchOrgans() // For å oppdatere teller
    } catch (e: any) {
        alert(e.message)
    }
}

const createOrgan = async () => {
    try {
        await api.createOrgan({
            org_id: selectedOrg.value,
            ...newOrgan.value
        })
        showCreate.value = false
        newOrgan.value = { name: '', parent_organ_id: '' }
        fetchOrgans()
    } catch (e: any) {
        alert(e.message)
    }
}

watch(selectedOrg, fetchOrgans)

onMounted(() => {
    if (props.orgs && props.orgs.length > 0) {
        selectedOrg.value = props.orgs[0]!.id
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
                        <option value="">Ingen (Toppnivå organ)</option>
                        <option v-for="o in organs" :key="o.id" :value="o.id">{{ o.name }}</option>
                    </BSelect>
                </div>
                <div class="mt-8 flex gap-4">
                    <BButton @click="createOrgan" variant="primary" class="flex-1 italic">LAGRE ORGAN</BButton>
                    <BButton @click="showCreate = false" variant="secondary">AVBRYT</BButton>
                </div>
            </BCard>
        </div>

        <div v-if="selectedOrgan">
            <header class="flex justify-between items-end border-b-4 border-black pb-4 mb-8">
                <div>
                    <BButton @click="selectedOrgan = null" variant="ghost" class="text-[10px] mb-4">← TILBAKE TIL LISTE</BButton>
                    <h3 class="text-4xl font-black uppercase italic tracking-tighter leading-none">{{ selectedOrgan.name }}</h3>
                    <p class="text-xs font-bold uppercase text-gray-400 mt-1 italic">Administrasjon av medlemmer og roller</p>
                </div>
                <BButton @click="showAddMember = true" variant="primary" class="text-xs py-2">+ LEGG TIL MEDLEM</BButton>
            </header>

            <div v-if="showAddMember" class="mb-12">
                <BCard class="bg-yellow-50 max-w-2xl mx-auto">
                    <h4 class="text-xl font-black uppercase mb-6 italic">Oppnevning til {{ selectedOrgan.name }}</h4>
                    <div class="space-y-6">
                        <BSelect v-model="newMemberID" label="Velg Medlem">
                            <option value="">Søk etter medlem...</option>
                            <option v-for="m in allMembers" :key="m.user_id" :value="m.user_id">
                                {{ m.name }} ({{ m.email }})
                            </option>
                        </BSelect>
                        <BInput v-model="newMemberRole" label="Rolle/Tittel" placeholder="F.eks. Styreleder" />
                    </div>
                    <div class="mt-8 flex gap-4">
                        <BButton @click="addOrganMember" variant="primary" class="flex-1 italic" :disabled="!newMemberID">FULLFØR OPPNEVNING</BButton>
                        <BButton @click="showAddMember = false" variant="secondary">AVBRYT</BButton>
                    </div>
                </BCard>
            </div>

            <BCard class="!p-0 overflow-hidden shadow-[16px_16px_0px_0px_rgba(0,0,0,1)]">
                <table class="brutalist-table">
                    <thead>
                        <tr>
                            <th class="brutalist-th">Medlem</th>
                            <th class="brutalist-th">Rolle</th>
                            <th class="brutalist-th">Oppnevnt</th>
                            <th class="brutalist-th text-right">Handlinger</th>
                        </tr>
                    </thead>
                    <tbody class="bg-white">
                        <tr v-if="membersLoading" v-for="i in 3">
                            <td colspan="4" class="brutalist-td animate-pulse bg-gray-50 h-12"></td>
                        </tr>
                        <tr v-else v-for="m in organMembers" :key="m.id" class="hover:bg-blue-50 transition-colors">
                            <td class="brutalist-td">
                                <div class="font-black uppercase tracking-tighter">{{ m.name }}</div>
                                <div class="text-[8px] font-black text-gray-400 tracking-widest">{{ m.email }}</div>
                            </td>
                            <td class="brutalist-td font-black uppercase text-xs">
                                <BBadge class="bg-gray-100">{{ m.role_type }}</BBadge>
                            </td>
                            <td class="brutalist-td text-[10px] font-bold">{{ new Date(m.created_at).toLocaleDateString() }}</td>
                            <td class="brutalist-td text-right">
                                <BButton @click="removeOrganMember(m.id)" variant="danger" class="text-[8px] py-1 px-2 uppercase">FJERN</BButton>
                            </td>
                        </tr>
                    </tbody>
                </table>
                <div v-if="organMembers.length === 0 && !membersLoading" class="py-20 text-center">
                    <p class="font-black uppercase text-gray-400 italic">Ingen medlemmer i dette organet ennå</p>
                </div>
            </BCard>
        </div>

        <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-8">
            <BCard v-for="o in organs" :key="o.id" class="hover:bg-blue-50 transition-colors flex flex-col justify-between">
                <div>
                    <span class="text-[8px] font-black uppercase text-gray-400 tracking-widest mb-1 block">ID: {{ o.id.split('-')[0] }}</span>
                    <h3 class="text-2xl font-black uppercase tracking-tighter leading-none mb-4">{{ o.name }}</h3>
                </div>
                <div class="pt-4 border-t-2 border-black border-dashed flex justify-between items-center">
                    <span class="text-[10px] font-bold uppercase italic">Aktive medlemmer: <span class="font-black">{{ o.member_count }}</span></span>
                    <BButton @click="fetchOrganMembers(o)" variant="ghost" class="text-[10px] py-1 px-2">ADMINISTRER</BButton>
                </div>
            </BCard>

            <div v-if="organs.length === 0 && !loading" class="col-span-full py-20 text-center border-8 border-black border-dashed">
                <p class="font-black uppercase text-gray-400 italic">Ingen organer registrert for denne enheten</p>
            </div>
        </div>
    </div>
</template>
