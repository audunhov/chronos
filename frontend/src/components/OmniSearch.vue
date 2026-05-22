<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../services/api'
import BCard from './base/BCard.vue'
import BInput from './base/BInput.vue'

const isOpen = ref(false)
const query = ref('')
const results = ref<any[]>([])
const loading = ref(false)
const activeIndex = ref(0)
const router = useRouter()

const handleKeydown = (e: KeyboardEvent) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        isOpen.value = true
    }
    if (e.key === 'Escape' && isOpen.value) {
        close()
    }
}

const onInputKeydown = (e: KeyboardEvent) => {
    if (e.key === 'ArrowDown') {
        e.preventDefault()
        activeIndex.value = Math.min(activeIndex.value + 1, results.value.length - 1)
    } else if (e.key === 'ArrowUp') {
        e.preventDefault()
        activeIndex.value = Math.max(activeIndex.value - 1, 0)
    } else if (e.key === 'Enter' && results.value.length > 0) {
        e.preventDefault()
        selectResult(results.value[activeIndex.value])
    }
}

const fetchResults = async () => {
    if (query.value.length < 2) {
        results.value = []
        return
    }
    loading.value = true
    try {
        const data = await api.globalSearch(query.value)
        results.value = Array.isArray(data) ? data : []
        activeIndex.value = 0
    } catch (e) {
        console.error(e)
        results.value = []
    } finally {
        loading.value = false
    }
}

let timeout: any
watch(query, () => {
    clearTimeout(timeout)
    timeout = setTimeout(fetchResults, 300)
})

const selectResult = (result: any) => {
    if (result.type === 'user') {
        router.push(`/admin/users/${result.id}`)
    } else if (result.type === 'pipeline') {
        router.push(`/admin/pipelines/edit/${result.id}`)
    } else if (result.type === 'org') {
        // Just navigate to members list
        router.push(`/admin/members`)
    }
    close()
}

const close = () => {
    isOpen.value = false
    query.value = ''
    results.value = []
}

onMounted(() => {
    window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
    window.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-[200] flex items-start justify-center pt-[15vh] p-4 bg-black/80 backdrop-blur-sm" @click.self="close">
        <div class="w-full max-w-2xl bg-white border-8 border-black shadow-[16px_16px_0px_0px_rgba(0,0,0,1)] flex flex-col overflow-hidden animate-in slide-in-from-top-10 duration-200">
            <div class="p-4 border-b-8 border-black flex items-center gap-4 bg-yellow-400">
                <span class="text-3xl font-black italic">SEARCH</span>
                <input 
                    v-model="query" 
                    @keydown="onInputKeydown"
                    ref="searchInput"
                    v-focus
                    placeholder="Søk etter medlemmer, organs, pipelines... (Cmd+K)" 
                    class="flex-1 bg-transparent text-black placeholder-black/50 font-mono text-xl focus:outline-none"
                />
                <button @click="close" class="bg-black text-white w-8 h-8 font-black hover:bg-white hover:text-black border-4 border-black transition-colors">X</button>
            </div>

            <div v-if="loading" class="p-8 text-center">
                <div class="inline-block animate-spin rounded-full h-8 w-8 border-4 border-black border-t-yellow-400"></div>
            </div>

            <div v-else-if="results.length > 0" class="max-h-[60vh] overflow-y-auto">
                <div 
                    v-for="(res, i) in results" :key="res.id" 
                    @click="selectResult(res)"
                    @mousemove="activeIndex = i"
                    :class="['p-4 border-b-2 border-dashed border-gray-300 cursor-pointer flex justify-between items-center transition-colors', activeIndex === i ? 'bg-black text-white' : 'hover:bg-gray-100']"
                >
                    <div>
                        <div :class="['font-black uppercase italic tracking-tighter text-xl', activeIndex === i ? 'text-yellow-400' : '']">{{ res.title }}</div>
                        <div :class="['text-xs font-mono', activeIndex === i ? 'text-gray-400' : 'text-gray-500']">{{ res.subtitle }}</div>
                    </div>
                    <span :class="['text-[10px] font-black uppercase px-2 py-1', activeIndex === i ? 'bg-white text-black' : 'bg-black text-white']">{{ res.type }}</span>
                </div>
            </div>

            <div v-else-if="query.length > 1" class="p-16 text-center text-gray-400">
                <p class="font-black uppercase tracking-widest text-xl italic mb-2">Ingen resultater</p>
                <p class="text-xs font-mono">Prøv et annet søkeord</p>
            </div>
            <div v-else class="p-16 text-center text-gray-400">
                <p class="font-black uppercase tracking-widest text-xl italic mb-2">Hva leter du etter?</p>
                <p class="text-xs font-mono">Søkefeltet finner medlemmer, hierarki og automasjoner.</p>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
// Custom directive to auto-focus
export default {
    directives: {
        focus: {
            mounted(el: HTMLElement) {
                el.focus()
            }
        }
    }
}
</script>
