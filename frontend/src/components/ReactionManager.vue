<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../services/api'
import BButton from './base/BButton.vue'
import BCard from './base/BCard.vue'
import BBadge from './base/BBadge.vue'
import BInput from './base/BInput.vue'
import BSelect from './base/BSelect.vue'

const props = defineProps<{
    orgs: { id: string, name: string }[]
}>()

const router = useRouter()
const selectedOrg = ref('')
const triggerAggregateID = ref('')
const forms = ref<any[]>([])
const reactions = ref<any[]>([])
const executions = ref<any[]>([])
const loading = ref(false)
const showCreate = ref(false)
const activeTab = ref<'pipelines' | 'history'>('pipelines')

// --- DAG PIPELINE BUILDER STATE (Legacy fallback if needed, but mostly Visual Editor is used now) ---
const triggerEvent = ref('MembershipCreated')
const triggerInterval = ref('daily')

const fetchReactions = async () => {
    if (!selectedOrg.value) return
    loading.value = true
    try {
        const data = await api.getReactions(selectedOrg.value)
        reactions.value = Array.isArray(data) ? data : []
    } catch (e) {
        console.error(e)
        reactions.value = []
    } finally {
        loading.value = false
    }
}

const fetchExecutions = async () => {
    if (!selectedOrg.value) return
    try {
        const data = await api.getPipelineExecutions(selectedOrg.value)
        executions.value = Array.isArray(data) ? data : []
    } catch (e) {
        console.error(e)
    }
}

const fetchForms = async () => {
    if (!selectedOrg.value) return
    try {
        const data = await api.getForms(selectedOrg.value)
        forms.value = Array.isArray(data) ? data : []
    } catch (e) {
        console.error(e)
    }
}

const createVisualPipeline = async () => {
    try {
        const res = await api.createReaction({
            org_id: selectedOrg.value,
            trigger_event: 'MembershipCreated',
            action_type: 'PIPELINE_DAG',
            config: { nodes: [], edges: [] }
        })
        // @ts-ignore
        if (res && res.id) {
            router.push(`/admin/pipelines/edit/${res.id}?org_id=${selectedOrg.value}`)
        }
    } catch (e: any) {
        alert(e.message)
    }
}

watch(selectedOrg, () => {
    fetchReactions()
    fetchForms()
    fetchExecutions()
})

onMounted(() => {
    if (props.orgs && props.orgs.length > 0) {
        selectedOrg.value = props.orgs[0]!.id
    }
})
</script>

<template>
    <div class="space-y-8">
        <header class="border-b-8 border-black pb-4 flex flex-col md:flex-row justify-between items-end gap-4">
            <div>
                <h2 class="text-4xl font-black uppercase tracking-tighter italic">Event Pipelines</h2>
                <p class="text-xs font-bold uppercase text-gray-500">Konfigurer reaksjoner via datakjeder (DAG)</p>
            </div>
            <div class="flex gap-4">
                <button 
                    @click="activeTab = 'pipelines'" 
                    :class="['text-xl font-black uppercase italic px-4 py-2 border-b-8 transition-colors', activeTab === 'pipelines' ? 'border-yellow-400 text-black' : 'border-transparent text-gray-400 hover:text-black']"
                >
                    Konfigurasjon
                </button>
                <button 
                    @click="activeTab = 'history'" 
                    :class="['text-xl font-black uppercase italic px-4 py-2 border-b-8 transition-colors', activeTab === 'history' ? 'border-yellow-400 text-black' : 'border-transparent text-gray-400 hover:text-black']"
                >
                    Historikk
                </button>
            </div>
        </header>

        <div v-if="activeTab === 'pipelines'" class="space-y-8 animate-in fade-in duration-300">
            <div class="flex flex-wrap gap-6 items-end bg-black text-white p-6 shadow-[8px_8px_0px_0px_rgba(0,0,0,0.3)]">
                <BSelect v-model="selectedOrg" label="Velg Organisasjon" class="min-w-[300px]">
                    <option v-for="org in orgs" :key="org.id" :value="org.id">{{ org.name }}</option>
                </BSelect>
                <div class="flex gap-2">
                    <BButton @click="createVisualPipeline" variant="success" class="italic text-xs py-2 bg-green-500 shadow-[4px_4px_0px_0px_white]">+ DESIGN NY PIPELINE</BButton>
                </div>
            </div>

            <!-- LIST EXISTING REACTIONS -->
            <div class="grid grid-cols-1 gap-6">
                <BCard v-for="r in reactions" :key="r.id" :class="r.is_inherited ? 'bg-gray-50 opacity-80 border-dashed' : 'bg-white border-solid'">
                    <div class="flex flex-col md:flex-row justify-between items-start md:items-center mb-6 border-b-4 border-black pb-4 gap-4">
                        <div class="flex flex-col gap-2">
                            <div class="flex items-center gap-4">
                                <BBadge class="bg-purple-400 text-lg italic">TRIGGER: {{ r.trigger_event }}</BBadge>
                                <span v-if="r.is_inherited" class="text-xs font-black uppercase text-blue-600 italic">↑ Arvet nedover</span>
                            </div>
                            <div v-if="r.trigger_aggregate_id" class="text-[10px] font-mono text-gray-500">
                                ID-FILTER: {{ r.trigger_aggregate_id }}
                            </div>
                            <div v-if="r.trigger_event === 'TimedSchedule'" class="text-[10px] font-black uppercase text-orange-600 bg-orange-100 px-2 border-2 border-orange-600 inline-block">
                                INTERVALL: {{ r.config?.interval || 'daily' }}
                            </div>
                        </div>
                        <div class="flex gap-4 items-center">
                            <router-link :to="'/admin/pipelines/edit/' + r.id + '?org_id=' + selectedOrg">
                                <BButton variant="primary" class="text-[10px] py-1 px-4 italic font-black underline">ÅPNE VISUELL EDITOR</BButton>
                            </router-link>
                            <BBadge class="bg-black text-white px-4">{{ r.action_type }}</BBadge>
                        </div>
                    </div>
                    
                    <div v-if="r.action_type === 'PIPELINE_DAG'" class="space-y-2 ml-4 border-l-4 border-black pl-4">
                        <div v-for="(node, idx) in r.config.nodes" :key="idx" class="flex items-center gap-3">
                            <span class="font-black text-xl leading-none italic">{{ Number(idx) + 1 }}.</span>
                            <div class="bg-gray-100 border-2 border-black px-3 py-1 flex-1 flex justify-between items-center">
                                <span class="font-black uppercase text-xs italic tracking-tight">{{ node.data?.label || node.type }}</span>
                                <span class="font-mono text-[8px] text-gray-500">{{ node.id }}</span>
                            </div>
                        </div>
                    </div>
                </BCard>

                <div v-if="reactions.length === 0 && !loading" class="py-32 text-center border-8 border-black border-dashed bg-white">
                    <h3 class="text-6xl font-black uppercase italic opacity-10 mb-4 tracking-tighter">Ingen dataflyt</h3>
                    <p class="font-bold uppercase text-gray-400 tracking-widest">Ingen pipelines aktive for denne organisasjonen.</p>
                </div>
            </div>
        </div>

        <div v-else-if="activeTab === 'history'" class="space-y-8 animate-in fade-in duration-300">
            <div class="flex flex-wrap gap-6 items-end bg-black text-white p-6 shadow-[8px_8px_0px_0px_rgba(0,0,0,0.3)]">
                <BSelect v-model="selectedOrg" label="Velg Organisasjon" class="min-w-[300px]">
                    <option v-for="org in orgs" :key="org.id" :value="org.id">{{ org.name }}</option>
                </BSelect>
                <BButton @click="fetchExecutions" variant="secondary" class="italic text-xs py-2 shadow-[4px_4px_0px_0px_white]">OPPDATER HISTORIKK</BButton>
            </div>

            <BCard class="!p-0 overflow-hidden shadow-[16px_16px_0px_0px_rgba(0,0,0,1)]">
                <table class="brutalist-table bg-white">
                    <thead>
                        <tr>
                            <th class="brutalist-th">Dato</th>
                            <th class="brutalist-th">Trigger Event</th>
                            <th class="brutalist-th">Status</th>
                            <th class="brutalist-th text-right">Detaljer</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y-2 divide-black">
                        <tr v-for="exec in executions" :key="exec.id" class="hover:bg-blue-50 transition-colors">
                            <td class="brutalist-td text-[10px] font-mono whitespace-nowrap">{{ new Date(exec.executed_at).toLocaleString('no-NO') }}</td>
                            <td class="brutalist-td font-black uppercase text-xs">
                                <BBadge class="bg-purple-100 text-purple-900 border-none">{{ exec.trigger_event }}</BBadge>
                                <div class="font-mono text-[8px] text-gray-500 mt-1">ID: {{ exec.pipeline_id }}</div>
                            </td>
                            <td class="brutalist-td">
                                <BBadge :variant="exec.status === 'SUCCESS' ? 'success' : 'danger'">{{ exec.status }}</BBadge>
                            </td>
                            <td class="brutalist-td text-right">
                                <!-- Temporary placeholder for "View Path" -> Will open a modal or dedicated route later -->
                                <button class="brutalist-btn bg-gray-100 text-[8px] py-1 px-2 border-2 italic font-bold text-gray-400 cursor-not-allowed">SE PATH (TODO)</button>
                            </td>
                        </tr>
                        <tr v-if="executions.length === 0">
                            <td colspan="4" class="p-20 text-center font-black uppercase text-gray-300 italic text-2xl">Ingen kjøringer registrert</td>
                        </tr>
                    </tbody>
                </table>
            </BCard>
        </div>
    </div>
</template>
