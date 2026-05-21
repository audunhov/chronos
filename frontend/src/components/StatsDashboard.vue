<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { api } from '../services/api'
import { Bar, Line } from 'vue-chartjs'
import { 
    Chart as ChartJS, 
    Title, Tooltip, Legend, 
    BarElement, CategoryScale, LinearScale,
    LineElement, PointElement
} from 'chart.js'
import BButton from './base/BButton.vue'

ChartJS.register(Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale, LineElement, PointElement)

const rawData = ref<any>(null)
const loading = ref(false)
const currentMetric = ref<'new' | 'churn' | 'total' | 'payments'>('total')

const chartData = computed(() => {
    if (!rawData.value) return { labels: [], datasets: [] }

    let dataset: any = {
        borderColor: '#000',
        borderWidth: 4,
        data: []
    }

    let labels: string[] = []
    
    switch (currentMetric.value) {
        case 'total':
            labels = rawData.value.total_active.map((d: any) => d.label)
            dataset.label = 'Totale Medlemmer (Akkumulert)'
            dataset.backgroundColor = '#60a5fa' // Blue 400
            dataset.data = rawData.value.total_active.map((d: any) => d.value)
            dataset.fill = true
            break
        case 'new':
            labels = rawData.value.new_members.map((d: any) => d.label)
            dataset.label = 'Nye Medlemmer'
            dataset.backgroundColor = '#fbbf24' // Yellow 400
            dataset.data = rawData.value.new_members.map((d: any) => d.value)
            break
        case 'churn':
            labels = rawData.value.churn.map((d: any) => d.label)
            dataset.label = 'Utmeldte (Shredded)'
            dataset.backgroundColor = '#f87171' // Red 400
            dataset.data = rawData.value.churn.map((d: any) => d.value)
            break
        case 'payments':
            labels = rawData.value.payments.map((d: any) => d.label)
            dataset.label = 'Mottatte Betalinger'
            dataset.backgroundColor = '#4ade80' // Green 400
            dataset.data = rawData.value.payments.map((d: any) => d.value)
            break
    }

    return {
        labels,
        datasets: [dataset]
    }
})

const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    scales: {
        y: {
            beginAtZero: true,
            ticks: { font: { weight: 'bold', family: 'monospace' } },
            border: { width: 4, color: 'black' }
        },
        x: {
            ticks: { font: { weight: 'bold', family: 'monospace' } },
            border: { width: 4, color: 'black' }
        }
    },
    plugins: {
        legend: {
            display: true,
            position: 'top' as const,
            labels: { font: { weight: 'black', family: 'monospace' } }
        }
    }
}

const fetchStats = async () => {
    loading.value = true
    try {
        const data = await api.getStats()
        rawData.value = data
    } catch (e) {
        console.error(e)
    } finally {
        loading.value = false
    }
}

onMounted(fetchStats)
</script>

<template>
    <div class="space-y-12">
        <header class="border-b-8 border-black pb-4 flex justify-between items-end">
            <div>
                <h2 class="text-4xl font-black uppercase italic tracking-tighter">System Analytics</h2>
                <p class="text-xs font-bold uppercase text-gray-500">Historisk data fra hendelsesstrømmen</p>
            </div>
            <div class="flex gap-2">
                <BButton 
                    @click="currentMetric = 'total'" 
                    :variant="currentMetric === 'total' ? 'primary' : 'ghost'"
                    class="text-[10px] py-1 px-3"
                >TOTAL</BButton>
                <BButton 
                    @click="currentMetric = 'new'" 
                    :variant="currentMetric === 'new' ? 'primary' : 'ghost'"
                    class="text-[10px] py-1 px-3"
                >VEKST</BButton>
                <BButton 
                    @click="currentMetric = 'churn'" 
                    :variant="currentMetric === 'churn' ? 'primary' : 'ghost'"
                    class="text-[10px] py-1 px-3"
                >AVSKALLING</BButton>
                <BButton 
                    @click="currentMetric = 'payments'" 
                    :variant="currentMetric === 'payments' ? 'primary' : 'ghost'"
                    class="text-[10px] py-1 px-3"
                >AKTIVITET</BButton>
            </div>
        </header>
        
        <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
            <div class="brutalist-card bg-blue-400">
                <p class="text-[10px] font-black uppercase mb-1">Total Population</p>
                <p class="text-4xl font-black italic">{{ rawData?.total_active?.[rawData.total_active.length-1]?.value || 0 }}</p>
            </div>
            <div class="brutalist-card bg-green-400">
                <p class="text-[10px] font-black uppercase mb-1">New This Month</p>
                <p class="text-4xl font-black italic">+{{ rawData?.new_members?.[rawData.new_members.length-1]?.value || 0 }}</p>
            </div>
            <div class="brutalist-card bg-red-400">
                <p class="text-[10px] font-black uppercase mb-1">Churn Rate</p>
                <p class="text-4xl font-black italic">{{ rawData?.churn?.[rawData.churn.length-1]?.value || 0 }}</p>
            </div>
            <div class="brutalist-card bg-yellow-400">
                <p class="text-[10px] font-black uppercase mb-1">Payments Processed</p>
                <p class="text-4xl font-black italic">{{ rawData?.payments?.[rawData.payments.length-1]?.value || 0 }}</p>
            </div>
        </div>

        <div class="brutalist-card bg-white h-[500px]">
            <Line v-if="!loading && currentMetric === 'total'" :data="chartData" :options="chartOptions as any" />
            <Bar v-else-if="!loading" :data="chartData" :options="chartOptions as any" />
            <div v-else class="h-full flex items-center justify-center animate-pulse font-black uppercase text-2xl italic">
                Scanning Event Stream...
            </div>
        </div>
    </div>
</template>
