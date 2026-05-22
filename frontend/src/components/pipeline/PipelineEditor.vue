<script setup lang="ts">
import { ref, onMounted, markRaw, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { VueFlow, useVueFlow, useNodesInitialized, type Node, type Edge, type Connection } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { MiniMap } from '@vue-flow/minimap'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import '@vue-flow/minimap/dist/style.css'

import { api } from '../../services/api'
import BButton from '../base/BButton.vue'

// Custom Nodes
import TriggerNode from './TriggerNode.vue'
import ActionNode from './ActionNode.vue'
import LogicNode from './LogicNode.vue'
import CodeNode from './CodeNode.vue'

const nodeTypes = {
  trigger: markRaw(TriggerNode),
  action: markRaw(ActionNode),
  logic: markRaw(LogicNode),
  code: markRaw(CodeNode),
}

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
    FindOrgan: { inputs: ['org_id', 'organ_name'], outputs: ['organ_id', 'name'] },
    ListOrgans: { inputs: ['org_id'], outputs: ['organs'] },
    FindMemberByRole: { inputs: ['org_id', 'role_name'], outputs: ['email', 'name', 'user_id'] },
    SendEmail: { inputs: ['to_email', 'subject', 'body'], outputs: ['success'] },
    ForEach: { inputs: ['list'], outputs: ['item'] },
    Collect: { inputs: ['value'], outputs: ['collected_item'] },
    FormatText: { inputs: ['template', 'user_name', 'org_name'], outputs: ['result'] },
}

const route = useRoute()
const router = useRouter()
const pipelineId = route.params.id as string

const { nodes, edges, onConnect, addEdges, addNodes, removeEdges, updateNodeData, toObject, project, screenToFlowCoordinate } = useVueFlow()
const nodesInitialized = useNodesInitialized()

const loading = ref(false)
const reaction = ref<any>(null)
const forms = ref<any[]>([])
const referenceData = ref({
    organizations: [] as any[],
    organs: [] as any[],
    roles: ['Leader', 'Deputy', 'Secretary', 'Treasurer', 'Member']
})

const fetchPipeline = async () => {
    loading.value = true
    try {
        const orgId = route.query.org_id as string
        const [data, formsData, orgsData, organsData] = await Promise.all([
            api.getReactions(orgId),
            api.getForms().catch(() => []),
            api.getOrganizationHierarchy().catch(() => []),
            api.getOrgans(orgId).catch(() => [])
        ])
        forms.value = Array.isArray(formsData) ? formsData : []
        referenceData.value.organizations = Array.isArray(orgsData) ? orgsData : []
        referenceData.value.organs = Array.isArray(organsData) ? organsData : []

        const r = Array.isArray(data) ? data.find(item => item.id === pipelineId) : null
        if (r && r.config) {
            reaction.value = r
            const mappedNodes = (r.config.nodes || []).map((n: any) => {
                const node = {
                    ...n,
                    position: n.position || { x: n.pos_x || 0, y: n.pos_y || 0 }
                }
                
                // Inject callbacks and data during load
                if (node.type === 'trigger') {
                    node.data = {
                        ...node.data,
                        aggregateId: r.trigger_aggregate_id || node.data?.aggregateId,
                        availableEvents: Object.keys(TRIGGER_VAR_MAP),
                        forms: forms.value,
                        onUpdateEvent: handleTriggerEventChange,
                        onUpdateAggregate: handleTriggerAggregateChange
                    }
                } else if (node.type === 'action' || node.type === 'logic' || node.type === 'code') {
                    node.data = {
                        ...node.data,
                        referenceData: referenceData.value,
                        onUpdateData: handleActionDataChange
                    }
                }
                return node
            })
            nodes.value = mappedNodes
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
    const triggerNode = nodes.value.find((n: Node) => n.type === 'trigger')
    if (!triggerNode) return

    let outputs = TRIGGER_VAR_MAP[event] || ['timestamp']
    
    if (event === 'FormResponseSubmitted' && aggregateId) {
        const form = forms.value.find((f: any) => f.id === aggregateId)
        if (form && form.schema && (form.schema as any).fields) {
            outputs = [
                ...outputs,
                ...(form.schema as any).fields.map((f: any) => `ans_${f.name}`)
            ]
        }
    }
    
    updateNodeData(triggerNode.id, {
        event,
        aggregateId,
        outputs,
    })
}

const handleTriggerEventChange = (newEvent: string) => {
    const triggerNode = nodes.value.find((n: Node) => n.type === 'trigger')
    if (triggerNode) {
        updateTriggerOutputs(newEvent, triggerNode.data.aggregateId)
    }
}

const handleTriggerAggregateChange = (newAgg: string) => {
    const triggerNode = nodes.value.find((n: Node) => n.type === 'trigger')
    if (triggerNode) {
        updateTriggerOutputs(triggerNode.data.event, newAgg)
    }
}

const handleActionDataChange = (nodeId: string, key: string, val: any) => {
    updateNodeData(nodeId, { [key]: val })
}

// Watch for edge changes to update node connection status
watch(edges, () => {
    nodes.value.forEach(node => {
        const edgesToNode = edges.value.filter(e => e.target === node.id)
        const connectedInputs = new Set(edgesToNode.map(e => e.targetHandle))
        if (JSON.stringify(Array.from(node.data.connectedInputs || [])) !== JSON.stringify(Array.from(connectedInputs))) {
            updateNodeData(node.id, { connectedInputs })
        }
    })
}, { deep: true })

onConnect((params: any) => {
    addEdges([params])
})

const checkValidConnection = (connection: Connection) => {
    if (connection.source === connection.target) return false;
    const targetNode = nodes.value.find(n => n.id === connection.target);
    if (targetNode?.type === 'trigger') return false;
    return true;
}

const validationErrors = computed(() => {
    const errors: string[] = []
    const trigger = nodes.value.find(n => n.type === 'trigger')
    if (!trigger) {
        errors.push('Mangler Start-node (Trigger)')
        return errors
    }
    const visited = new Set<string>()
    const queue = [trigger.id]
    while (queue.length > 0) {
        const curr = queue.shift()!
        visited.add(curr)
        const outgoingEdges = edges.value.filter(e => e.source === curr)
        for (const edge of outgoingEdges) {
            if (!visited.has(edge.target)) {
                queue.push(edge.target)
            }
        }
    }
    nodes.value.forEach(node => {
        if (!visited.has(node.id)) {
            errors.push(`Node "${node.data.label || node.data.type || node.id}" er ikke koblet til flyten.`)
        }
        if (node.type === 'action' && node.data.inputs) {
            node.data.inputs.forEach((inp: string) => {
                const hasConnection = edges.value.some(e => e.target === node.id && e.targetHandle === inp)
                if (!hasConnection && !node.data[inp]) {
                    errors.push(`Node "${node.data.label}" mangler inndata for "${inp}".`)
                }
            })
        }
    })
    return errors
})

const addActionNode = (type: string) => {
    const def = ACTION_DEFS[type]
    const id = `node_${Date.now()}`
    const newNode: Node = {
        id,
        type: 'action',
        position: { x: 400, y: 100 },
        data: {
            label: type,
            inputs: def?.inputs || [],
            outputs: def?.outputs || [],
            referenceData: referenceData.value,
            onUpdateData: (key: string, val: any) => handleActionDataChange(id, key, val)
        }
    }
    addNodes([newNode])
}

const addLogicNode = (type: 'if' | 'filter' | 'list_filter') => {
    const id = `node_${Date.now()}`
    const newNode: Node = {
        id,
        type: 'logic',
        position: { x: 400, y: 100 },
        data: { 
            type, 
            operator: '==', 
            value1: '', 
            value2: '',
            inputs: type === 'list_filter' ? ['list', 'operator', 'value'] : ['v1', 'v2', 'operator'],
            onUpdateData: (key: string, val: any) => handleActionDataChange(id, key, val)
        }
    }
    addNodes([newNode])
}

const onDragStart = (event: DragEvent, type: string, nodeClass: 'action' | 'logic' | 'code') => {
    if (event.dataTransfer) {
        event.dataTransfer.setData('application/vueflow', JSON.stringify({ type, nodeClass }))
        event.dataTransfer.effectAllowed = 'move'
    }
}

const onDrop = (event: DragEvent) => {
    const dataStr = event.dataTransfer?.getData('application/vueflow')
    if (!dataStr) return
    const { type, nodeClass } = JSON.parse(dataStr)
    
    const position = screenToFlowCoordinate({
        x: event.clientX,
        y: event.clientY,
    })

    if (nodeClass === 'action') addActionNode(type)
    else if (nodeClass === 'logic') addLogicNode(type as any)
    else if (nodeClass === 'code') addNodes([{ 
        id: `node_${Date.now()}`, 
        type: 'code', 
        position, 
        data: { 
            code: '', 
            inputs: [], 
            referenceData: referenceData.value,
            onUpdateData: handleActionDataChange
        } 
    }])
}

const save = async () => {
    const flow: any = toObject()
    const triggerNode = nodes.value.find((n: any) => n.type === 'trigger')
    if (triggerNode) {
        flow.trigger_event = triggerNode.data.event
        flow.trigger_aggregate_id = triggerNode.data.aggregateId
    }

    try {
        await api.updateReaction(pipelineId, flow)
        alert('Pipeline lagret og publisert!')
    } catch (e: any) {
        alert('Feil ved lagring: ' + e.message)
    }
}

const deletePipeline = async () => {
    if (!confirm('Er du sikker på at du vil slette denne pipelinen?')) return
    try {
        await api.deleteReaction(pipelineId)
        router.back()
    } catch (e: any) {
        alert('Feil ved sletting: ' + e.message)
    }
}

const showTestModal = ref(false)
const testDataInput = ref('{\n  "user_id": "123",\n  "email": "test@test.no",\n  "list": [1,2,3]\n}')
const testLogs = ref<string[]>([])
const isTesting = ref(false)
const openTestModal = () => {
    showTestModal.value = true
    testLogs.value = []
}
const runTest = async () => {
    try {
        const triggerData = JSON.parse(testDataInput.value)
        isTesting.value = true
        testLogs.value = ['Starter test...']
        
        nodes.value.forEach(node => {
            if (node.type === 'code') updateNodeData(node.id, { logs: [] })
        })

        const flow: any = toObject()
        const triggerNode = nodes.value.find((n: any) => n.type === 'trigger')
        if (triggerNode) {
            flow.trigger_event = triggerNode.data.event
            flow.trigger_aggregate_id = triggerNode.data.aggregateId
        }

        const res = await api.testPipeline(pipelineId, triggerData, flow)
        if (res.success) {
            testLogs.value.push('✅ Test fullført uten feil.')
        } else {
            testLogs.value.push('❌ Test feilet: ' + res.error)
        }
        
        if (res.logs) {
            testLogs.value.push(...res.logs)
            res.logs.forEach((log: string) => {
                nodes.value.forEach(node => {
                    if (node.type === 'code' && (log.includes(node.id) || log.includes('JS Error'))) {
                        const currentLogs = node.data.logs || []
                        updateNodeData(node.id, { logs: [...currentLogs, log] })
                    }
                })
            })
        }
    } catch (e: any) {
        testLogs.value = ['❌ Ugyldig JSON eller systemfeil: ' + e.message]
    } finally {
        isTesting.value = false
    }
}

const deleteSelected = () => {
    const selectedEdges = edges.value.filter((e: any) => e.selected)
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
                availableEvents: Object.keys(TRIGGER_VAR_MAP),
                forms: forms.value,
                onUpdateEvent: handleTriggerEventChange,
                onUpdateAggregate: handleTriggerAggregateChange
            }
        }])
    }
})
</script>

<template>
  <div class="fixed inset-0 bg-orange-50 flex flex-col overflow-hidden font-mono text-black" @keydown.backspace="deleteSelected" @keydown.delete="deleteSelected" tabindex="0">
    <header class="bg-black text-white p-4 flex justify-between items-center border-b-8 border-black z-50 shrink-0">
        <div class="flex items-center gap-6">
            <BButton @click="router.back()" variant="ghost" class="text-white border-white text-xs py-1 hover:bg-white hover:text-black transition-colors italic">← TILBAKE</BButton>
            <div class="h-8 w-2 bg-yellow-400"></div>
            <h1 class="text-3xl font-black uppercase italic tracking-tighter leading-none">Visual Pipeline Builder</h1>
        </div>
        <div class="flex gap-4">
            <BButton @click="deletePipeline" variant="danger" class="text-xs py-2 px-6">SLETT PIPELINE</BButton>
            <BButton @click="openTestModal" variant="secondary" class="text-xs py-2 px-6">TEST KJØRING</BButton>
            <BButton @click="save" variant="primary" class="text-xs py-2 px-10 shadow-[4px_4px_0px_0px_white]">PUBLISER ENDRINGER</BButton>
        </div>
    </header>

    <div class="flex-1 flex overflow-hidden min-h-0 relative">
        <div v-if="loading" class="absolute inset-0 z-[60] bg-orange-50/80 backdrop-blur-sm flex items-center justify-center">
            <div class="flex flex-col items-center gap-4">
                <div class="w-16 h-16 border-8 border-black border-t-yellow-400 animate-spin"></div>
                <p class="font-black uppercase italic tracking-tighter">Henter Pipeline...</p>
            </div>
        </div>
        
        <aside class="w-80 bg-white border-r-8 border-black p-6 space-y-10 overflow-y-auto z-40 shadow-[8px_0px_0px_0px_rgba(0,0,0,0.1)] shrink-0">
            <section class="space-y-4">
                <h4 class="font-black uppercase text-xs border-b-4 border-black pb-2">Logikk</h4>
                <div class="grid grid-cols-2 gap-2">
                    <button draggable="true" @dragstart="onDragStart($event, 'if', 'logic')" @click="addLogicNode('if')" class="brutalist-btn bg-yellow-100 text-[10px] p-2 hover:bg-yellow-200 font-black italic cursor-grab">IF / THEN</button>
                    <button draggable="true" @dragstart="onDragStart($event, 'list_filter', 'logic')" @click="addLogicNode('list_filter')" class="brutalist-btn bg-blue-100 text-[10px] p-2 hover:bg-blue-200 font-black italic cursor-grab">FILTER (LIST)</button>
                    <button draggable="true" @dragstart="onDragStart($event, 'javascript', 'code')" @click="addNodes([{ id: `node_${Date.now()}`, type: 'code', position: { x: 400, y: 100 }, data: { code: '', inputs: [] } }])" class="brutalist-btn bg-gray-900 text-white text-[10px] p-2 hover:bg-black font-black italic cursor-grab col-span-2">CUSTOM JS CODE</button>
                </div>
            </section>
            <section class="space-y-4">
                <h4 class="font-black uppercase text-xs border-b-4 border-black pb-2">Handlinger</h4>
                <div class="space-y-2">
                    <button v-for="type in Object.keys(ACTION_DEFS)" :key="type" 
                        draggable="true"
                        @dragstart="onDragStart($event, type, 'action')"
                        @click="addActionNode(type)"
                        class="brutalist-btn w-full bg-white text-left text-[10px] p-2 hover:bg-blue-50 font-black italic cursor-grab"
                    >
                        + {{ type }}
                    </button>
                </div>
            </section>
            <div class="brutalist-card bg-black text-white p-4 !shadow-none italic text-[8px] leading-relaxed">
                STATUS: Markér en kobling og trykk [Backspace] for å slette. <br><br>
                TIPS: Dra og slipp (Drag & Drop) noder fra menyen inn på lerretet, eller klikk på dem for å legge til.<br><br>
                NYTT: TEST KJØRING lar deg se resultatet uten å endre data permanent.
            </div>
            <section v-if="validationErrors.length > 0" class="space-y-2 bg-red-50 border-4 border-red-500 p-2">
                <h4 class="font-black uppercase text-[10px] text-red-600 border-b-2 border-red-500 pb-1">Advarsler ({{ validationErrors.length }})</h4>
                <ul class="text-[9px] space-y-1 font-mono text-red-800">
                    <li v-for="(err, i) in validationErrors" :key="i" class="flex gap-1">
                        <span class="font-black">!</span> <span>{{ err }}</span>
                    </li>
                </ul>
            </section>
        </aside>

        <main class="flex-1 relative overflow-hidden min-w-0" @dragover.prevent @drop="onDrop">
            <VueFlow 
                v-model:nodes="nodes"
                v-model:edges="edges"
                :node-types="nodeTypes"
                :is-valid-connection="checkValidConnection"
                fit-view-on-init
                class="brutalist-flow"
            >
                <Background pattern-color="#000" :gap="20" />
                <Controls position="bottom-right" />
                <MiniMap pannable zoomable class="!border-4 !border-black !rounded-none !shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] !bg-white" />
            </VueFlow>
        </main>
    </div>

    <div v-if="showTestModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm">
        <div class="bg-white border-8 border-black p-6 w-full max-w-2xl shadow-[16px_16px_0px_0px_rgba(0,0,0,1)] flex flex-col max-h-[90vh]">
            <div class="flex justify-between items-center border-b-4 border-black pb-4 mb-4">
                <h2 class="text-2xl font-black uppercase italic tracking-tighter">Test Pipeline</h2>
                <button @click="showTestModal = false" class="bg-red-500 text-white border-4 border-black w-8 h-8 font-black hover:bg-black hover:text-red-500">X</button>
            </div>
            <div class="flex-1 overflow-y-auto space-y-4 min-h-0 flex flex-col">
                <div class="flex-1 flex flex-col min-h-[200px]">
                    <label class="text-xs font-black uppercase mb-2">Simulert Trigger-data (JSON)</label>
                    <textarea v-model="testDataInput" class="flex-1 w-full bg-orange-50 border-4 border-black p-4 font-mono text-xs focus:outline-none focus:bg-yellow-50 resize-none"></textarea>
                </div>
                <div class="h-64 flex flex-col">
                    <label class="text-xs font-black uppercase mb-2">Utførelseslogg</label>
                    <div class="flex-1 bg-black text-green-400 p-4 font-mono text-[10px] overflow-y-auto border-4 border-black">
                        <div v-for="(log, i) in testLogs" :key="i" class="mb-1">{{ log }}</div>
                        <div v-if="testLogs.length === 0" class="text-gray-500 italic">Kjør test for å se output...</div>
                    </div>
                </div>
            </div>
            <div class="mt-4 pt-4 border-t-4 border-black text-right shrink-0">
                <BButton @click="runTest" :disabled="isTesting" variant="primary" class="w-full text-sm py-3">
                    {{ isTesting ? 'KJØRER...' : 'KJØR TEST' }}
                </BButton>
            </div>
        </div>
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
    stroke: #f87171 !important;
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
