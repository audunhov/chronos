<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../services/api'
import type { Member } from '../api'
import BSelect from '../components/base/BSelect.vue'
import BInput from '../components/base/BInput.vue'
import BButton from '../components/base/BButton.vue'
import BBadge from '../components/base/BBadge.vue'
import EmailComposer from '../components/pipeline/EmailComposer.vue'

const router = useRouter()
const members = ref<Member[]>([])
const organizations = ref<{id: string, name: string}[]>([])
const selectedOrg = ref('')
const selectedDate = ref('')
const loading = ref(false)

// Selection & Email
const selectedMemberIds = ref<Set<string>>(new Set())
const showEmailComposer = ref(false)

const selectedUserIds = computed(() => {
    return Array.from(selectedMemberIds.value).map(id => {
        const m = members.value.find(member => member.id === id)
        return m?.user_id || ''
    }).filter(id => id !== '')
})

const selectedNamesSample = computed(() => {
    const names = Array.from(selectedMemberIds.value).slice(0, 3).map(id => {
        const m = members.value.find(member => member.id === id)
        return m?.name || 'Ukjent'
    })
    if (selectedMemberIds.value.size > 3) return names.join(', ') + '...'
    return names.join(', ')
})

const toggleAll = () => {
    if (selectedMemberIds.value.size === members.value.length) {
        selectedMemberIds.value.clear()
    } else {
        members.value.forEach(m => selectedMemberIds.value.add(m.id!))
    }
}

const toggleOne = (id: string) => {
    if (selectedMemberIds.value.has(id)) selectedMemberIds.value.delete(id)
    else selectedMemberIds.value.add(id)
}

const fetchMembers = async (silent = false) => {
  if (!silent) loading.value = true
  try {
    let data: any
    if (selectedDate.value) {
      data = await api.getMembersAsOf(selectedDate.value, selectedOrg.value)
    } else {
      data = await api.getMembers(selectedOrg.value)
    }
    members.value = Array.isArray(data) ? data : []
    selectedMemberIds.value.clear()
  } catch (e) {
    console.error(e)
  } finally {
    if (!silent) loading.value = false
  }
}

const fetchOrgs = async () => {
  try {
    const data = await api.getOrganizations()
    organizations.value = Array.isArray(data) ? (data as any).map((d: any) => {
        if (typeof d === 'string') return { id: d, name: d }
        return d
    }) : []
  } catch (e) {
    console.error(e)
  }
}

const shredMember = async (id: string) => {
  if (!confirm('Slett personopplysninger?')) return
  try {
    await api.shredMember(id)
    fetchMembers()
  } catch (e: any) {
    alert(e.message)
  }
}

const inspectUser = (userId: string) => {
    router.push(`/admin/users/${userId}`)
}

watch([selectedDate, selectedOrg], () => fetchMembers())

onMounted(() => {
    fetchMembers()
    fetchOrgs()
})
</script>

<template>
    <div class="space-y-8 animate-in fade-in">
        <div class="flex flex-wrap gap-6 items-end bg-black text-white p-6 shadow-[8px_8px_0px_0px_rgba(0,0,0,0.3)]">
            <BSelect v-model="selectedOrg" label="Org-Filter" class="bg-black text-white border-white min-w-[250px]">
                <option value="">ALLE ORGANISASJONER</option>
                <option v-for="org in organizations" :key="org.id" :value="org.id">{{ org.name }}</option>
            </BSelect>

            <BInput v-model="selectedDate" type="date" label="Historisk dato (As-Of)" class="bg-black text-white border-white" />
            
            <div class="flex gap-2">
                <BButton @click="fetchMembers()" variant="secondary" class="italic">OPPDATER</BButton>
                <BButton 
                    v-if="selectedMemberIds.size > 0"
                    @click="showEmailComposer = true" 
                    variant="primary" 
                    class="bg-blue-600 border-white hover:bg-blue-500 font-black italic"
                >
                    SEND E-POST ({{ selectedMemberIds.size }})
                </BButton>
            </div>
        </div>

        <div class="overflow-x-auto border-4 border-black shadow-[16px_16px_0px_0px_rgba(0,0,0,1)]">
            <table class="brutalist-table">
            <thead>
                <tr>
                <th class="brutalist-th w-10">
                    <input type="checkbox" @change="toggleAll" :checked="selectedMemberIds.size === members.length && members.length > 0" class="w-5 h-5 border-4 border-black accent-black" />
                </th>
                <th class="brutalist-th">Identitet</th>
                <th class="brutalist-th">Org</th>
                <th class="brutalist-th">Status</th>
                <th class="brutalist-th">Saldo</th>
                <th class="brutalist-th text-right">Handlinger</th>
                </tr>
            </thead>
            <tbody class="bg-white">
                <tr v-if="loading" v-for="i in 5" :key="i">
                    <td colspan="6" class="brutalist-td animate-pulse bg-gray-50 h-12"></td>
                </tr>
                <tr v-else v-for="m in members" :key="m.id" 
                    @click="inspectUser(m.user_id!)"
                    class="hover:bg-yellow-50 transition-colors cursor-pointer group"
                    :class="selectedMemberIds.has(m.id!) ? 'bg-yellow-100' : ''"
                >
                <td class="brutalist-td" @click.stop>
                    <input type="checkbox" :checked="selectedMemberIds.has(m.id!)" @change="toggleOne(m.id!)" class="w-5 h-5 border-4 border-black accent-black" />
                </td>
                <td class="brutalist-td">
                    <div class="font-black uppercase tracking-tighter group-hover:text-blue-600 transition-colors">{{ m.name }}</div>
                    <div class="text-[10px] font-bold text-gray-400">{{ m.email }}</div>
                </td>
                <td class="brutalist-td"><BBadge class="bg-gray-100">{{ (m.org_id || '').split('-')[0] }}</BBadge></td>
                <td class="brutalist-td">
                    <BBadge :variant="m.status === 'ACTIVE' ? 'success' : 'danger'">{{ m.status }}</BBadge>
                </td>
                <td class="brutalist-td font-black" :class="((m.balance || 0) < 0) ? 'text-red-600' : 'text-green-600'">
                    {{ ((m.balance || 0) / 100).toFixed(2) }} kr
                </td>
                <td class="brutalist-td text-right">
                    <BButton 
                        v-if="m.status !== 'SHREDDED'"
                        @click.stop="shredMember(m.id!)" 
                        variant="danger"
                        class="text-[8px] py-1 px-2 uppercase"
                    >
                        SHRED
                    </BButton>
                </td>
                </tr>
                <tr v-if="members.length === 0 && !loading">
                    <td colspan="6" class="p-20 text-center font-black uppercase text-gray-300 italic text-4xl">Tomt Ark</td>
                </tr>
            </tbody>
            </table>
        </div>

        <!-- Reusable Email Modal -->
        <EmailComposer 
            v-if="showEmailComposer"
            :recipient-ids="selectedUserIds"
            :recipient-names="selectedNamesSample"
            @close="showEmailComposer = false"
        />
    </div>
</template>
