<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../services/api'
import type { TreasuryItem } from '../api'
import BCard from '../components/base/BCard.vue'

const treasuryReport = ref<TreasuryItem[]>([])
const loading = ref(false)

const fetchTreasuryReport = async () => {
    loading.value = true
    try {
        const data = await api.getTreasuryReport()
        treasuryReport.value = Array.isArray(data) ? data : []
    } catch (e) {
        console.error('Failed to fetch treasury report:', e)
    } finally {
        loading.value = false
    }
}

onMounted(fetchTreasuryReport)
</script>

<template>
    <div class="grid grid-cols-1 gap-8 md:grid-cols-2 lg:grid-cols-3 animate-in fade-in">
        <BCard v-for="t in treasuryReport" :key="t.id" class="flex flex-col group hover:bg-green-50">
            <div class="mb-4">
                <p class="text-[8px] font-black text-gray-400 tracking-widest">{{ t.path }}</p>
                <h3 class="text-3xl font-black uppercase italic leading-none">{{ t.name }}</h3>
            </div>
            <div class="mt-auto space-y-4 pt-4 border-t-4 border-black border-dashed">
                <div class="flex justify-between items-end">
                    <span class="text-[10px] font-black uppercase text-gray-500">Lokal Saldo:</span>
                    <span class="text-xl font-bold tracking-tighter">{{ (t.local_balance / 100).toFixed(2) }} kr</span>
                </div>
                <div class="flex justify-between items-end bg-black text-white p-2">
                    <span class="text-[10px] font-black uppercase">Branch Sum:</span>
                    <span class="text-2xl font-black italic">{{ (t.total_branch_balance / 100).toFixed(2) }} kr</span>
                </div>
            </div>
        </BCard>
        <div v-if="treasuryReport.length === 0 && !loading" class="col-span-full p-20 text-center brutalist-card border-dashed">
            <p class="text-4xl font-black uppercase text-gray-300 italic opacity-20">Ingen data tilgjengelig</p>
        </div>
    </div>
</template>
