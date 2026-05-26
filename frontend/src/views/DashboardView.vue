<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../services/api'
import { auth } from '../services/auth'
import type { MyMembership } from '../api'
import BCard from '../components/base/BCard.vue'
import BBadge from '../components/base/BBadge.vue'
import BButton from '../components/base/BButton.vue'

const myMemberships = ref<MyMembership[]>([])
const loading = ref(false)
const showModal = ref(false)

const fetchMyMemberships = async () => {
  if (!auth.user) return
  loading.value = true
  try {
    const data = await api.getMyMemberships()
    myMemberships.value = Array.isArray(data) ? data : []
  } catch (e) {
    console.error('Failed to fetch my memberships:', e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchMyMemberships)
</script>

<template>
  <div class="space-y-8">
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <h2 class="text-3xl font-black uppercase tracking-tighter underline decoration-yellow-400 decoration-8">Velkommen tilbake!</h2>
        <BButton
            @click="$emit('open-register')"
            variant="primary"
        >
            + BLI MEDLEM
        </BButton>
    </div>

    <div v-if="loading" class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
        <div v-for="i in 3" :key="i" class="brutalist-card animate-pulse h-48 bg-gray-100"></div>
    </div>
    <div v-else-if="myMemberships.length > 0" class="grid grid-cols-1 gap-8 sm:grid-cols-2 lg:grid-cols-3">
        <div v-for="ms in myMemberships" :key="ms.id" class="brutalist-card bg-white group">
            <div class="flex justify-between items-start mb-6">
                <div>
                    <h3 class="font-black text-2xl uppercase leading-none mb-1 group-hover:text-blue-600 transition-colors">{{ ms.org_name }}</h3>
                    <BBadge class="bg-gray-200">{{ ms.role }}</BBadge>
                </div>
                <BBadge :variant="ms.status === 'ACTIVE' ? 'success' : 'danger'">
                    {{ ms.status }}
                </BBadge>
            </div>
            <div class="mt-8 pt-6 border-t-4 border-black flex justify-between items-end">
                <div v-if="(ms.balance || 0) < 0">
                    <p class="text-[10px] font-black uppercase text-gray-500 mb-1">Ditt utestående:</p>
                    <p class="text-3xl font-black italic text-red-600">
                        {{ (Math.abs(ms.balance || 0) / 100).toFixed(2) }} <small class="text-sm not-italic">NOK</small>
                    </p>
                </div>
                <div v-else>
                    <p class="text-[10px] font-black uppercase text-green-600 mb-1">Økonomi:</p>
                    <p class="text-3xl font-black italic text-black uppercase tracking-tighter">BETALT</p>
                </div>
                <BButton v-if="(ms.balance || 0) < 0" variant="primary" class="text-[10px] px-3 py-1">BETAL</BButton>
            </div>
        </div>
    </div>
    <div v-else class="brutalist-card bg-white border-dashed border-gray-400 py-20 text-center">
        <p class="text-xl font-bold uppercase text-gray-400">Ingen aktive medlemskap funnet.</p>
    </div>
  </div>
</template>
