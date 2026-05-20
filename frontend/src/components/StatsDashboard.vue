<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../services/api'
import { Bar } from 'vue-chartjs'
import { Chart as ChartJS, Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale } from 'chart.js'

ChartJS.register(Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale)

const stats = ref<any>(null)
const loading = ref(false)

const chartData = ref({
    labels: [] as string[],
    datasets: [
        {
            label: 'Nye Medlemmer',
            backgroundColor: '#fbbf24', // Yellow 400
            borderColor: '#000',
            borderWidth: 4,
            data: [] as number[]
        }
    ]
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
            labels: { font: { weight: 'black', family: 'monospace' } }
        }
    }
}

const fetchStats = async () => {
    loading.value = true
    try {
        const data = await api.getStats()
        if (Array.isArray(data)) {
            chartData.value = {
                labels: data.map(d => d.label),
                datasets: [{
                    ...chartData.value.datasets[0],
                    data: data.map(d => d.value)
                }]
            }
            stats.value = data
        }
    } catch (e) {
        console.error(e)
    } finally {
        loading.value = false
    }
}

onMounted(fetchStats)
</script>

<template>
    <div class="space-y-8">
        <h2 class="text-2xl font-black uppercase">Statistikk & Vekst</h2>
        
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div class="brutalist-card bg-blue-400">
                <p class="text-xs font-black uppercase mb-1">Totale Hendelser</p>
                <p class="text-4xl font-black italic">PRO</p>
            </div>
            <div class="brutalist-card bg-green-400">
                <p class="text-xs font-black uppercase mb-1">Vekst denne måned</p>
                <p class="text-4xl font-black italic">+{{ stats?.[stats.length-1]?.value || 0 }}</p>
            </div>
            <div class="brutalist-card bg-yellow-400">
                <p class="text-xs font-black uppercase mb-1">Systemstatus</p>
                <p class="text-4xl font-black italic">NOMINAL</p>
            </div>
        </div>

        <div class="brutalist-card bg-white h-[400px]">
            <Bar v-if="!loading" :data="chartData" :options="chartOptions as any" />
            <div v-else class="h-full flex items-center justify-center animate-pulse font-black uppercase">
                Aggregerer data...
            </div>
        </div>
    </div>
</template>
