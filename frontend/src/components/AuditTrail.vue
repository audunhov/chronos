<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api } from '../services/api'
import type { AuditLogEntry } from '../api'
import BSelect from './base/BSelect.vue'
import BCard from './base/BCard.vue'
import BBadge from './base/BBadge.vue'

const props = defineProps<{
    orgs: {id: string, name: string}[]
}>()

const logs = ref<AuditLogEntry[]>([])
const loading = ref(false)
const selectedOrg = ref('')

const fetchLogs = async () => {
    loading.value = true
    try {
        const data = await api.getAuditLogs(selectedOrg.value)
        logs.value = Array.isArray(data) ? data : []
    } catch (e) {
        console.error(e)
    } finally {
        loading.value = false
    }
}

const formatDate = (date: string) => {
    return new Date(date).toLocaleString('no-NO')
}

watch(selectedOrg, fetchLogs)
onMounted(fetchLogs)
</script>

<template>
    <div class="space-y-8">
        <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
            <h2 class="text-4xl font-black uppercase italic tracking-tighter">System Audit Trail</h2>
            <BSelect v-model="selectedOrg" label="Filter på organisasjon" class="min-w-[300px]">
                <option value="">ALLE (GLOBAL)</option>
                <option v-for="org in orgs" :key="org.id" :value="org.id">{{ org.name }}</option>
            </BSelect>
        </div>

        <div class="border-4 border-black shadow-[16px_16px_0px_0px_rgba(0,0,0,1)] overflow-hidden">
            <table class="w-full border-collapse">
                <thead class="bg-black text-white uppercase text-xs">
                    <tr>
                        <th class="p-4 text-left font-black italic">Tidspunkt</th>
                        <th class="p-4 text-left font-black italic">Handling</th>
                        <th class="p-4 text-left font-black italic">Aktør</th>
                        <th class="p-4 text-left font-black italic">Organisasjon</th>
                        <th class="p-4 text-left font-black italic">Detaljer</th>
                    </tr>
                </thead>
                <tbody class="bg-white divide-y-2 divide-black">
                    <tr v-if="loading" v-for="i in 5" :key="i">
                        <td colspan="5" class="p-4 animate-pulse bg-gray-50 h-16"></td>
                    </tr>
                    <tr v-else v-for="entry in logs" :key="entry.id" class="hover:bg-yellow-50 transition-colors">
                        <td class="p-4 text-[10px] font-bold font-mono">
                            {{ formatDate(entry.created_at || '') }}
                        </td>
                        <td class="p-4">
                            <BBadge class="bg-indigo-600 text-white border-none">{{ entry.action }}</BBadge>
                            <div class="text-[8px] mt-1 font-mono text-gray-400">CID: {{ (entry.correlation_id || '').split('-')[0] }}...</div>
                        </td>
                        <td class="p-4 font-black text-sm italic">
                            {{ entry.actor_email || 'ANONYM' }}
                        </td>
                        <td class="p-4 font-bold text-xs uppercase">
                            {{ entry.org_name || 'SYSTEM' }}
                        </td>
                        <td class="p-4">
                            <pre class="text-[10px] bg-gray-100 p-2 border-2 border-black font-mono overflow-auto max-w-xs max-h-24">{{ JSON.stringify(entry.detail, null, 2) }}</pre>
                        </td>
                    </tr>
                    <tr v-if="logs.length === 0 && !loading">
                        <td colspan="5" class="p-20 text-center font-black uppercase text-gray-400 italic text-2xl opacity-20">
                            Ingen hendelser funnet i loggen
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>
