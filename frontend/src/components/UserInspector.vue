<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api } from '../services/api'
import type { User, MyMembership, AuditLogEntry } from '../api'
import BCard from './base/BCard.vue'
import BBadge from './base/BBadge.vue'
import BButton from './base/BButton.vue'

const props = defineProps<{
    userId: string
}>()

const profile = ref<User | null>(null)
const memberships = ref<MyMembership[]>([])
const auditLogs = ref<AuditLogEntry[]>([])
const loading = ref(false)

const fetchData = async () => {
    if (!props.userId) return
    loading.value = true
    try {
        const [user, ms, logs] = await Promise.all([
            api.getUserProfile(props.userId),
            api.getUserMemberships(props.userId),
            api.getAuditLogs('', props.userId, props.userId)
        ])
        profile.value = user
        memberships.value = Array.isArray(ms) ? ms : []
        auditLogs.value = Array.isArray(logs) ? logs : []
    } catch (e) {
        console.error("Failed to inspect user:", e)
    } finally {
        loading.value = false
    }
}

const formatDate = (date: string) => {
    return new Date(date).toLocaleString('no-NO')
}

watch(() => props.userId, fetchData)
onMounted(fetchData)
</script>

<template>
    <div class="space-y-8 animate-in fade-in duration-500">
        <div v-if="loading" class="py-20 text-center">
            <div class="inline-block animate-spin rounded-full h-8 w-8 border-4 border-black border-t-yellow-400"></div>
            <p class="mt-4 font-black uppercase italic tracking-widest">Inspeksjon pågår...</p>
        </div>

        <div v-else-if="profile" class="space-y-12">
            <!-- Header Profile Info -->
            <div class="flex flex-col md:flex-row justify-between items-start gap-6 border-b-8 border-black pb-8">
                <div>
                    <h2 class="text-6xl font-black uppercase italic tracking-tighter">{{ profile.name }}</h2>
                    <p class="text-xl font-bold bg-yellow-400 px-4 py-1 border-4 border-black inline-block mt-4">{{ profile.email }}</p>
                    <div class="mt-4 font-mono text-xs text-gray-500 uppercase">UUID: {{ profile.id }}</div>
                </div>
                <div class="text-right">
                    <p class="text-[10px] font-black uppercase text-gray-500">System-opprettet:</p>
                    <p class="font-black italic text-lg">{{ formatDate(profile.created_at!) }}</p>
                </div>
            </div>

            <div class="grid grid-cols-1 xl:grid-cols-2 gap-12">
                <!-- Memberships & Finances -->
                <div class="space-y-6">
                    <h3 class="text-3xl font-black uppercase italic border-b-4 border-black pb-2">Medlemskap & Saldo</h3>
                    <div v-if="memberships.length > 0" class="space-y-4">
                        <div v-for="ms in memberships" :key="ms.id" class="brutalist-card bg-white group hover:bg-yellow-50 transition-colors">
                            <div class="flex justify-between items-start mb-4">
                                <div>
                                    <h4 class="font-black text-xl uppercase">{{ ms.org_name }}</h4>
                                    <BBadge class="bg-indigo-100 text-indigo-900 border-none">{{ ms.role }}</BBadge>
                                </div>
                                <span :class="['brutalist-badge', ms.status === 'ACTIVE' ? 'bg-green-400' : 'bg-red-400']">{{ ms.status }}</span>
                            </div>
                            <div class="mt-6 pt-4 border-t-2 border-black border-dashed flex justify-between items-end">
                                <div>
                                    <p class="text-[10px] font-black uppercase text-gray-400 mb-1">Saldo på dette medlemskapet:</p>
                                    <p :class="['text-2xl font-black italic', (ms.balance || 0) < 0 ? 'text-red-600' : 'text-black']">
                                        {{ ((ms.balance || 0) / 100).toFixed(2) }} <small class="text-sm not-italic">NOK</small>
                                    </p>
                                </div>
                                <div class="text-right text-[10px] text-gray-400 font-bold uppercase">
                                    Sist oppdatert:<br>{{ formatDate(ms.updated_at!) }}
                                </div>
                            </div>
                        </div>
                    </div>
                    <div v-else class="py-10 text-center border-4 border-black border-dashed bg-gray-50">
                        <p class="font-black uppercase text-gray-400">Ingen registrerte medlemskap</p>
                    </div>
                </div>

                <!-- Activity Log -->
                <div class="space-y-6">
                    <h3 class="text-3xl font-black uppercase italic border-b-4 border-black pb-2">Siste Aktivitet</h3>
                    <div class="border-4 border-black shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] overflow-hidden">
                        <table class="w-full border-collapse">
                            <thead class="bg-black text-white uppercase text-[8px]">
                                <tr>
                                    <th class="p-2 text-left font-black italic">Tid</th>
                                    <th class="p-2 text-left font-black italic">Handling</th>
                                    <th class="p-2 text-left font-black italic">Org</th>
                                </tr>
                            </thead>
                            <tbody class="bg-white divide-y-2 divide-black">
                                <tr v-for="log in auditLogs.slice(0, 15)" :key="log.id" class="hover:bg-gray-50 text-[10px] font-bold">
                                    <td class="p-2 font-mono whitespace-nowrap">{{ new Date(log.created_at!).toLocaleDateString() }}</td>
                                    <td class="p-2">
                                        <div class="font-black uppercase italic">{{ log.action }}</div>
                                        <div v-if="log.target_id === profile.id" class="text-[8px] text-blue-600 uppercase underline">MÅL-OBJEKT</div>
                                        <div v-else-if="log.actor_email === profile.email" class="text-[8px] text-indigo-600 uppercase underline">UTFØRT AV BRUKER</div>
                                    </td>
                                    <td class="p-2 uppercase truncate max-w-[100px]">{{ log.org_name || 'SYSTEM' }}</td>
                                </tr>
                                <tr v-if="auditLogs.length === 0">
                                    <td colspan="3" class="p-10 text-center font-black uppercase text-gray-300 italic">Ingen historikk</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
