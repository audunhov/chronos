<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../services/api'
import type { TreasuryItem } from '../api'
import BCard from '../components/base/BCard.vue'
import BButton from '../components/base/BButton.vue'

const treasuryReport = ref<TreasuryItem[]>([])
const loading = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)
const reconciling = ref(false)
const reconcileResults = ref<any>(null)

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

const exportCSV = () => {
    if (treasuryReport.value.length === 0) return
    const headers = ['ID', 'Navn', 'Path', 'Lokal Saldo (kr)', 'Total Branch Saldo (kr)']
    const rows = treasuryReport.value.map(t => [
        t.id, 
        t.name, 
        t.path, 
        (t.local_balance / 100).toFixed(2), 
        (t.total_branch_balance / 100).toFixed(2)
    ])
    
    const csvContent = [
        headers.join(','),
        ...rows.map(e => e.join(','))
    ].join('\n')

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.setAttribute('href', url)
    link.setAttribute('download', `chronos_treasury_${new Date().toISOString().slice(0,10)}.csv`)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
}

const triggerFileUpload = () => {
    fileInput.value?.click()
}

const handleFileUpload = async (event: Event) => {
    const target = event.target as HTMLInputElement
    const file = target.files?.[0]
    if (!file) return

    reconciling.value = true
    reconcileResults.value = null
    try {
        const res = await api.reconcileTreasury(file)
        reconcileResults.value = res
        fetchTreasuryReport() // Refresh balances
    } catch (e: any) {
        alert('Avstemming feilet: ' + e.message)
    } finally {
        reconciling.value = false
        if (fileInput.value) fileInput.value.value = ''
    }
}

onMounted(fetchTreasuryReport)
</script>

<template>
    <div class="space-y-8 animate-in fade-in">
        <header class="flex flex-col md:flex-row justify-between items-end border-b-8 border-black pb-4 gap-4">
            <div>
                <h2 class="text-4xl font-black uppercase tracking-tighter italic">Treasury</h2>
                <p class="text-xs font-bold uppercase text-gray-500">Økonomisk oversikt og avstemming</p>
            </div>
            <div class="flex gap-4">
                <input type="file" ref="fileInput" @change="handleFileUpload" accept=".csv" class="hidden" />
                <BButton @click="triggerFileUpload" variant="secondary" class="text-xs py-2" :disabled="reconciling">
                    {{ reconciling ? 'BEHANDLER...' : 'LAST OPP BANKFIL (CSV)' }}
                </BButton>
                <BButton @click="exportCSV" variant="primary" class="text-xs py-2 shadow-[4px_4px_0px_0px_white]">EKSPORTER CSV</BButton>
            </div>
        </header>

        <div v-if="reconcileResults" class="bg-green-100 border-4 border-black p-6 flex justify-between items-center animate-in slide-in-from-top-4">
            <div>
                <h3 class="text-xl font-black uppercase text-green-800 italic">Avstemming Fullført</h3>
                <p class="text-xs font-bold text-green-900 mt-1">
                    Fant {{ reconcileResults.matched }} treff i bankfilen. Automatiske betalings-events er trigget.
                </p>
            </div>
            <div class="text-right">
                <p class="text-[10px] font-black uppercase text-green-700 mb-1">Totalt Reconciliert</p>
                <p class="text-3xl font-black italic">{{ (reconcileResults.amount / 100).toFixed(2) }} kr</p>
            </div>
        </div>

        <div class="grid grid-cols-1 gap-8 md:grid-cols-2 lg:grid-cols-3">
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
    </div>
</template>
