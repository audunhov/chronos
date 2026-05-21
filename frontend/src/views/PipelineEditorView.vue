<script setup lang="ts">
import { ref, onMounted, markRaw } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { VueFlow, useVueFlow, type Node, type Edge } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'

import { api } from '../services/api'
import BButton from '../components/base/BButton.vue'

// Custom Nodes
import TriggerNode from '../components/pipeline/TriggerNode.vue'
import ActionNode from '../components/pipeline/ActionNode.vue'
import LogicNode from '../components/pipeline/LogicNode.vue'

const nodeTypes = {
  trigger: markRaw(TriggerNode),
  action: markRaw(ActionNode),
  logic: markRaw(LogicNode),
}

const route = useRoute()
const router = useRouter()
const pipelineId = route.params.id as string

const { nodes, edges, onConnect, addEdges, addNodes, removeEdges, updateNodeData, toObject } = useVueFlow()

const loading = ref(false)
const reaction = ref<any>(null)

onConnect((params) => {
    addEdges([params])
})

const TRIGGER_VAR_MAP: Record<string, string[]> = {
    MembershipCreated: ['user_id', 'org_id', 'role', 'timestamp'],
    FeeGenerated: ['id', 'amount', 'period', 'timestamp'],
    FormResponseSubmitted: ['form_id', 'user_id', 'timestamp'],
    OrganizationCreated: ['id', 'name', 'path', 'timestamp'],
    OrganCreated: ['id', 'org_id', 'name', 'timestamp'],
    TimedSchedule: ['timestamp', 'source'],
}

const ACTION_DEFS: Record<string, { inputs: string[], outputs: string[] }> = {
    FindOrg: { inputs: ['start_org_id', 'relation'], outputs: ['org_id', 'name'] },
    FindRole: { inputs: ['target_id', 'role_type'], outputs: ['user_id', 'email', 'name'] },
    SendEmail: { inputs: ['to_email', 'subject', 'body'], outputs: ['success'] },
    ForEach: { inputs: ['list'], outputs: ['item'] },
    Collect: { inputs: ['value'], outputs: ['collected_item'] },
}

const fetchPipeline = async () => {
    loading.value = true
    try {
        const orgId = route.query.org_id as string
        const data = await api.getReactions(orgId)
        const r = Array.isArray(data) ? data.find(item => item.id === pipelineId) : null
        
        if (r && r.config) {
            reaction.value = r
            nodes.value = (r.config.nodes || []).map((n: any) => ({
                ...n,
                position: n.position || { x: n.pos_x || 0, y: n.pos_y || 0 }
            }))
            edges.value = r.config.edges || []
            
            if (r.trigger_event) {
                await updateTriggerOutputs(r.trigger_event, r.trigger_aggregate_id || undefined)
            }
        }
    } catch (e) {
        console.error(e)
    } finally {
        loading.value = false
    }
}

const updateTriggerOutputs = async (event: string, aggregateId?: string) => {
    const triggerNode = nodes.value.find(n => n.type === 'trigger')
    if (!triggerNode) return

    let outputs = TRIGGER_VAR_MAP[event] || ['timestamp']
    
    if (event === 'FormResponseSubmitted' && aggregateId) {
        try {
            const formsData = await api.getForms()
            const form = Array.isArray(formsData) ? formsData.find(f => f.id === aggregateId) : null
            if (form && form.schema && (form.schema as any).fields) {
                outputs = [
                    ...outputs,
                    ...(form.schema as any).fields.map((f: any) => `ans_${f.name}`)
                ]
            }
        } catch (e) { console.error(e) }
    }
    
    updateNodeData(triggerNode.id, {
        event,
        outputs,
        availableEvents: Object.keys(TRIGGER_VAR_MAP)
    })
}

const addActionNode = (type: string) => {
    const def = ACTION_DEFS[type]
    const newNode: Node = {
        id: `node_${Date.now()}`,
        type: 'action',
        position: { x: 400, y: 100 },
        data: {
            label: type,
            inputs: def?.inputs || [],
            outputs: def?.outputs || []
        }
    }
    addNodes([newNode])
}

const addLogicNode = (type: 'if' | 'filter' | 'list_filter') => {
    const newNode: Node = {
        id: `node_${Date.now()}`,
        type: 'logic',
        position: { x: 400, y: 100 },
        data: { 
            type, 
            operator: '==', 
            value1: '', 
            value2: '',
            inputs: type === 'list_filter' ? ['list', 'operator', 'value'] : ['v1', 'v2', 'operator']
        }
    }
    addNodes([newNode])
}

const save = async () => {
    const flow: any = toObject()
    
    const triggerNode = nodes.value.find(n => n.type === 'trigger')
    if (triggerNode) {
        flow.trigger_event = triggerNode.data.event
    }

    try {
        await api.updateReaction(pipelineId, flow)
        alert('Pipeline lagret og publisert!')
    } catch (e: any) {
        alert('Feil ved lagring: ' + e.message)
    }
}

const runTest = async () => {
    const dataStr = prompt('Skriv inn test-data (JSON):', '{"user_id": "123", "email": "test@test.no", "list": [1,2,3]}')
    if (!dataStr) return
    
    try {
        const triggerData = JSON.parse(dataStr)
        const res = await api.testPipeline(pipelineId, triggerData)
        if (res.success) {
            alert('Test fullført! Sjekk konsollen for logger.')
        } else {
            alert('Test feilet: ' + res.error)
        }
        console.log('--- PIPELINE TEST LOGS ---')
        res.logs?.forEach((l: string) => console.log(l))
    } catch (e: any) {
        alert('Ugyldig JSON eller systemfeil: ' + e.message)
    }
}

const deleteSelected = () => {
    const selectedEdges = edges.value.filter(e => e.selected)
    removeEdges(selectedEdges)
}

onMounted(async () => {
    await fetchPipeline()
    if (nodes.value.length === 0) {
        addNodes([{ 
            id: 'trigger', 
            type: 'trigger', 
            position: { x: 50, y: 50 },
            data: { 
                event: 'MembershipCreated', 
                outputs: TRIGGER_VAR_MAP['MembershipCreated'],
                availableEvents: Object.keys(TRIGGER_VAR_MAP)
            }
        }])
    }
})
</script>

<template>
  <div class="fixed inset-0 bg-orange-50 flex flex-col overflow-hidden font-mono text-black" @keydown.backspace="deleteSelected" @keydown.delete="deleteSelected" tabindex="0">
    <!-- Navbar -->
    <header class="bg-black text-white p-4 flex justify-between items-center border-b-8 border-black z-50 shrink-0">
        <div class="flex items-center gap-6">
            <BButton @click="router.back()" variant="ghost" class="text-white border-white text-xs py-1 hover:bg-white hover:text-black transition-colors italic">← TILBAKE</BButton>
            <div class="h-8 w-2 bg-yellow-400"></div>
            <h1 class="text-3xl font-black uppercase italic tracking-tighter leading-none">Visual Pipeline Builder</h1>
        </div>
        <div class="flex gap-4">
            <BButton @click="runTest" variant="secondary" class="text-xs py-2 px-6">TEST KJØRING</BButton>
            <BButton @click="save" variant="primary" class="text-xs py-2 px-10 shadow-[4px_4px_0px_0px_white]">PUBLISER ENDRINGER</BButton>
        </div>
    </header>

    <div class="flex-1 flex overflow-hidden min-h-0">
        <!-- Sidebar -->
        <aside class="w-80 bg-white border-r-8 border-black p-6 space-y-10 overflow-y-auto z-40 shadow-[8px_0px_0px_0px_rgba(0,0,0,0.1)] shrink-0">
            <section class="space-y-4">
                <h4 class="font-black uppercase text-xs border-b-4 border-black pb-2">Logikk</h4>
                <div class="grid grid-cols-2 gap-2">
                    <button @click="addLogicNode('if')" class="brutalist-btn bg-yellow-100 text-[10px] p-2 hover:bg-yellow-200 font-black italic">IF / THEN</button>
                    <button @click="addLogicNode('list_filter')" class="brutalist-btn bg-blue-100 text-[10px] p-2 hover:bg-blue-200 font-black italic">FILTER (LIST)</button>
                </div>
            </section>

            <section class="space-y-4">
                <h4 class="font-black uppercase text-xs border-b-4 border-black pb-2">Handlinger</h4>
                <div class="space-y-2">
                    <button v-for="type in Object.keys(ACTION_DEFS)" :key="type" 
                        @click="addActionNode(type)"
                        class="brutalist-btn w-full bg-white text-left text-[10px] p-2 hover:bg-blue-50 font-black italic"
                    >
                        + {{ type }}
                    </button>
                </div>
            </section>

            <div class="brutalist-card bg-black text-white p-4 !shadow-none italic text-[8px] leading-relaxed">
                STATUS: Markér en kobling og trykk [Backspace] for å slette. <br><br>
                TIPS: IF-noden kan referere variabler ved å skrive f.eks. 'trigger.user_id' i verdi-feltet. <br><br>
                NYTT: TEST KJØRING lar deg se resultatet uten å endre data permanent.
            </div>
        </aside>

        <!-- Canvas -->
        <main class="flex-1 relative overflow-hidden min-w-0">
            <VueFlow 
                v-model:nodes="nodes" 
                v-model:edges="edges" 
                :node-types="nodeTypes"
                fit-view-on-init
                class="brutalist-flow"
            >
                <Background pattern-color="#000" :gap="20" />
                <Controls position="bottom-right" />
            </VueFlow>
        </main>
    </div>
  </div>
</template>

<style>
.brutalist-flow .vue-flow__node {
    padding: 0;
    border: none;
    background: transparent;
}
.brutalist-flow .vue-flow__edge-path {
    stroke: black !important;
    stroke-width: 6 !important;
}
.brutalist-flow .vue-flow__edge.selected .vue-flow__edge-path {
    stroke: #f87171 !important; /* Red 400 */
}
.brutalist-flow .vue-flow__connection-path {
    stroke: black !important;
    stroke-width: 4 !important;
    stroke-dasharray: 8;
}
.brutalist-flow .vue-flow__handle {
    width: 14px;
    height: 14px;
    border: 3px solid black;
    background: white;
}
.brutalist-flow .vue-flow__controls {
    border: 4px solid black;
    box-shadow: 8px 8px 0px 0px rgba(0,0,0,1);
}
.brutalist-flow .vue-flow__controls-button {
    border-bottom: 4px solid black;
    background: white;
}
.brutalist-flow .vue-flow__controls-button:hover {
    background: #fbbf24;
}
</style>
