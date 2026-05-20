<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../services/api'
import type { OrgNode } from '../api'
import BButton from './base/BButton.vue'
import BCard from './base/BCard.vue'
import BBadge from './base/BBadge.vue'
import BInput from './base/BInput.vue'
import BSelect from './base/BSelect.vue'

const hierarchy = ref<OrgNode[]>([])
const loading = ref(false)
const showCreate = ref(false)

const newOrg = ref({
    name: '',
    parent_id: '',
    allow_multiple: true
})

const fetchHierarchy = async () => {
    loading.value = true
    try {
        const data = await api.getOrganizationHierarchy()
        hierarchy.value = Array.isArray(data) ? data : []
    } catch (e) {
        console.error(e)
    } finally {
        loading.value = false
    }
}

const createOrg = async () => {
    try {
        await api.createOrganization({
            name: newOrg.value.name,
            parent_id: newOrg.value.parent_id,
            policy: { allow_multiple: newOrg.value.allow_multiple }
        })
        newOrg.value = { name: '', parent_id: '', allow_multiple: true }
        showCreate.value = false
        fetchHierarchy()
    } catch (e: any) {
        alert(e.message)
    }
}

const deleteOrg = async (node: OrgNode) => {
    const msg = `ADVARSEL: Er du helt sikker på at du vil slette "${node.name}"?\n\n` +
                `Dette vil permanent slette:\n` +
                `- Alle under-organisasjoner\n` +
                `- Alle skjemaer og svar tilknyttet denne grenen\n` +
                `- Alle rolletildelinger og policyer\n\n` +
                `HANDLINGEN KAN IKKE ANGRES.`
    
    if (!confirm(msg)) return
    
    try {
        await api.deleteOrganization(node.id!)
        fetchHierarchy()
    } catch (e: any) {
        alert(e.message)
    }
}

const getDepth = (path: string) => {
    return (path.split('.').length - 1)
}

onMounted(fetchHierarchy)
</script>

<template>
    <div class="space-y-8">
        <div class="flex justify-between items-end border-b-8 border-black pb-4">
            <div>
                <h2 class="text-4xl font-black uppercase tracking-tighter italic">Organisasjonskart</h2>
                <p class="text-xs font-bold uppercase text-gray-500">Struktur og tilgangsstyring</p>
            </div>
            <BButton @click="showCreate = true" variant="primary" class="text-xs py-2">+ NY ENHET</BButton>
        </div>

        <div v-if="showCreate">
            <BCard class="bg-yellow-50 max-w-2xl mx-auto">
                <h3 class="text-xl font-black uppercase mb-6 italic">Opprett ny enhet</h3>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <BInput v-model="newOrg.name" label="Navn" placeholder="F.eks. Oslo Brettspillklubb" required />
                    
                    <BSelect v-model="newOrg.parent_id" label="Overordnet Enhet">
                        <option value="">Ingen (Rot)</option>
                        <option v-for="node in hierarchy" :key="node.id" :value="node.id">
                            {{ node.path }} ({{ node.name }})
                        </option>
                    </BSelect>
                    
                    <div class="flex items-center gap-4 md:col-span-2 p-4 border-2 border-black bg-white">
                        <input type="checkbox" v-model="newOrg.allow_multiple" id="allow_mult" class="w-8 h-8 border-4 border-black" />
                        <label for="allow_mult" class="font-black uppercase text-xs cursor-pointer">
                            Tillat flere medlemskap (ADDITIV / UMBRELLA)
                        </label>
                    </div>
                </div>
                <div class="mt-8 flex gap-4">
                    <BButton @click="createOrg" variant="primary" class="flex-1 italic">LAGRE ENHET</BButton>
                    <BButton @click="showCreate = false" variant="secondary">AVBRYT</BButton>
                </div>
            </BCard>
        </div>

        <BCard class="!p-0 overflow-hidden shadow-[16px_16px_0px_0px_rgba(0,0,0,1)]">
            <table class="brutalist-table">
                <thead>
                    <tr>
                        <th class="brutalist-th">Hierarki & Navn</th>
                        <th class="brutalist-th text-center w-32">Policy</th>
                        <th class="brutalist-th text-right w-40">Handlinger</th>
                    </tr>
                </thead>
                <tbody class="bg-white">
                    <tr v-if="loading" v-for="i in 5">
                        <td colspan="3" class="brutalist-td animate-pulse bg-gray-50 h-12"></td>
                    </tr>
                    <tr v-else v-for="node in hierarchy" :key="node.id" class="hover:bg-blue-50 transition-colors">
                        <td class="brutalist-td">
                            <div :style="{ marginLeft: (getDepth(node.path!) * 2) + 'rem' }" class="flex items-center gap-3">
                                <span v-if="getDepth(node.path!) > 0" class="text-gray-400 font-black text-2xl">↳</span>
                                <div>
                                    <span class="font-black text-xl uppercase tracking-tighter">{{ node.name }}</span>
                                    <div class="text-[8px] font-black text-gray-400 tracking-widest">{{ node.path }}</div>
                                </div>
                            </div>
                        </td>
                        <td class="brutalist-td text-center">
                            <BBadge :class="node.policy?.allow_multiple ? 'bg-blue-400' : 'bg-orange-400'">
                                {{ node.policy?.allow_multiple ? 'ADDITIV' : 'EKSKLUSIV' }}
                            </BBadge>
                        </td>
                        <td class="brutalist-td text-right">
                            <BButton @click="deleteOrg(node)" variant="danger" class="text-[10px] py-1 px-4 italic uppercase">
                                SLETT
                            </BButton>
                        </td>
                    </tr>
                </tbody>
            </table>
            
            <div v-if="hierarchy.length === 0 && !loading" class="py-20 text-center">
                <p class="font-black uppercase text-gray-400 italic">Ingen organisasjoner definert i systemet.</p>
            </div>
        </BCard>
    </div>
</template>
