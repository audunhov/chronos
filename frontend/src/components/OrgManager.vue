<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../services/api'
import type { OrgNode } from '../api'

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

const deleteOrg = async (id: string) => {
    if (!confirm('Slett denne organisasjonen?')) return
    try {
        await api.deleteOrganization(id)
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
        <div class="flex justify-between items-center">
            <h2 class="text-2xl font-black uppercase">Organisasjonskart</h2>
            <button @click="showCreate = true" class="brutalist-btn-primary text-xs">+ NY ENHET</button>
        </div>

        <div v-if="showCreate" class="brutalist-card bg-yellow-50 mb-8">
            <h3 class="font-black uppercase mb-4 underline">Opprett ny enhet</h3>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div class="flex flex-col">
                    <label class="brutalist-label">Navn</label>
                    <input v-model="newOrg.name" type="text" class="brutalist-input" placeholder="Navn..." />
                </div>
                <div class="flex flex-col">
                    <label class="brutalist-label">Overordnet Enhet</label>
                    <select v-model="newOrg.parent_id" class="brutalist-input">
                        <option value="">Ingen (Rot)</option>
                        <option v-for="node in hierarchy" :key="node.id" :value="node.id">
                            {{ node.path }} ({{ node.name }})
                        </option>
                    </select>
                </div>
                <div class="flex items-center gap-3">
                    <input type="checkbox" v-model="newOrg.allow_multiple" id="allow_mult" class="w-6 h-6 border-4 border-black" />
                    <label for="allow_mult" class="font-bold uppercase text-xs">Tillat flere medlemskap (Umbrella)</label>
                </div>
            </div>
            <div class="mt-8 flex gap-4">
                <button @click="createOrg" class="brutalist-btn bg-black text-white px-8">LAGRE</button>
                <button @click="showCreate = false" class="brutalist-btn bg-white">AVBRYT</button>
            </div>
        </div>

        <div class="overflow-x-auto shadow-[12px_12px_0px_0px_rgba(0,0,0,1)]">
            <table class="brutalist-table bg-white">
                <thead>
                    <tr>
                        <th class="brutalist-th">Hierarki & Navn</th>
                        <th class="brutalist-th text-center">Policy</th>
                        <th class="brutalist-th text-right">Handlinger</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="node in hierarchy" :key="node.id" class="hover:bg-gray-50 transition-colors">
                        <td class="brutalist-td">
                            <div :style="{ marginLeft: (getDepth(node.path!) * 2) + 'rem' }" class="flex items-center gap-2">
                                <span v-if="getDepth(node.path!) > 0" class="text-gray-400 font-black">↳</span>
                                <span class="font-black text-lg uppercase">{{ node.name }}</span>
                                <span class="text-[10px] bg-gray-100 px-1 border border-black font-mono">{{ node.path }}</span>
                            </div>
                        </td>
                        <td class="brutalist-td text-center">
                            <span :class="['brutalist-badge', node.policy?.allow_multiple ? 'bg-blue-400' : 'bg-orange-400']">
                                {{ node.policy?.allow_multiple ? 'ADDITIV' : 'EKSKLUSIV' }}
                            </span>
                        </td>
                        <td class="brutalist-td text-right">
                            <button @click="deleteOrg(node.id!)" class="brutalist-btn bg-white hover:bg-red-500 hover:text-white px-3 py-1 text-[10px]">SLETT</button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>
