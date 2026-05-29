<script setup lang="ts">
import { ref, onMounted, markRaw, computed, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { VueFlow, useVueFlow, useNodesInitialized, type Node, type Edge, type Connection, type NodeChange, type EdgeChange, applyNodeChanges, applyEdgeChanges } from '@vue-flow/core'
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

import { PIPELINE_PRESETS } from './presets'

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
    HTTPRequest: { inputs: ['url', 'method', 'payload', 'secret_id'], outputs: ['status', 'response'] },
}

const route = useRoute()
const router = useRouter()
const pipelineId = route.params.id as string
const isHistoryMode = route.name === 'pipeline-history'
const executionId = route.params.id as string

// --- VUE FLOW STATE ---
const { 
    nodes, edges, onConnect, addEdges, addNodes, removeEdges, removeNodes,
    updateNodeData, toObject, project, screenToFlowCoordinate,
    fitView, findNode, setCenter, onPaneReady
} = useVueFlow('main')

const containerRef = ref<HTMLElement | null>(null)
const dataLoaded = ref(false)
const loading = ref(false)
const nodesInitialized = useNodesInitialized()
const reaction = ref<any>(null)
const execution = ref<any>(null)
const forms = ref<any[]>([])
const referenceData = ref({
    organizations: [] as any[],
    organs: [] as any[],
    secrets: [] as any[],
    roles: ['Leader', 'Deputy', 'Secretary', 'Treasurer', 'Member']
})

// --- HELPERS ---
const safeFitView = () => {
    if (!containerRef.value || containerRef.value.offsetWidth === 0) return
    try {
        fitView({ padding: 0.2, duration: 800 })
    } catch (e) {
        // Silently fail, it's just a view adjustment
    }
}

const flowKey = ref(0)
watch(nodes, () => {
    if (nodes.value.length > 0 && dataLoaded.value) {
        flowKey.value++
        nextTick(() => setTimeout(safeFitView, 300))
    }
}, { once: true })

const getNodeColor = (node: any) => {
    if (node.type === 'trigger') return '#a855f7'
    if (node.type === 'logic') return '#eab308'
    if (node.type === 'action') return '#3b82f6'
    if (node.type === 'code') return '#22c55e'
    return '#000'
}

const focusNode = (nodeId: string) => {
    const node = findNode(nodeId)
    if (node) {
        setCenter(node.position.x + (node.dimensions?.width || 200) / 2, node.position.y + (node.dimensions?.height || 150) / 2, { zoom: 1.2, duration: 800 })
    }
}

// --- API ACTIONS ---
const fetchPipeline = async () => {
    loading.value = true
    try {
        const orgId = route.query.org_id as string
        const targetPipelineId = isHistoryMode ? (route.query.pipeline_id as string) : pipelineId

        const [data, formsData, orgsData, organsData, execsData, secretsData] = await Promise.all([
            api.getReactions(orgId),
            api.getForms().catch(() => []),
            api.getOrganizationHierarchy().catch(() => []),
            api.getOrgans(orgId).catch(() => []),
            isHistoryMode ? api.getPipelineExecutions(orgId) : Promise.resolve([]),
            api.getSecrets(orgId).catch(() => [])
        ])
        
        forms.value = Array.isArray(formsData) ? formsData : []
        referenceData.value.organizations = Array.isArray(orgsData) ? orgsData : []
        referenceData.value.organs = Array.isArray(organsData) ? organsData : []
        referenceData.value.secrets = Array.isArray(secretsData) ? secretsData : []

        if (isHistoryMode) {
            execution.value = Array.isArray(execsData) ? execsData.find(e => e.id === executionId) : null
        }

        const r = Array.isArray(data) ? data.find(item => item.id === targetPipelineId) : null
        if (r && r.config) {
            reaction.value = r
            
            // 1. Load Nodes First
            nodes.value = (r.config.nodes || []).map((n: any) => ({
                ...n,
                position: n.position || { x: n.pos_x || 0, y: n.pos_y || 0 },
                dimensions: n.dimensions || { width: 200, height: 150 },
                data: {
                    ...n.data,
                    referenceData: referenceData.value,
                    onUpdateData: isHistoryMode ? undefined : (k: string, v: any) => updateNodeData(n.id, { [k]: v })
                }
            }))

            // Wait for nodes to be rendered and measured
            await nextTick()

            // 2. Load Edges Second
            edges.value = (r.config.edges || []).map((e: any) => ({
                ...e,
                type: 'smoothstep',
                animated: true,
                markerEnd: { type: 'arrowclosed', color: '#000' },
                style: { strokeWidth: 4, stroke: '#000' }
            }))
            
            if (r.trigger_event) {
                await updateTriggerOutputs(r.trigger_event, r.trigger_aggregate_id || undefined)
            }
            
            nextTick(() => {
                updateAllConnectedInputs()
                setTimeout(safeFitView, 500)
            })
        }
        dataLoaded.value = true
    } catch (e) {
        console.error('Fetch failed:', e)
    } finally {
        loading.value = false
    }
}

const save = async () => {
    const flow: any = toObject()
    const triggerNode = nodes.value.find((n: any) => n.type === 'trigger')
    if (triggerNode) {
        flow.trigger_event = triggerNode.data.event
        flow.trigger_aggregate_id = triggerNode.data.aggregateId || null
    }

    try {
        await api.updateReaction(pipelineId, flow)
        alert('Pipeline lagret!')
    } catch (e: any) {
        console.error('Save error:', e)
        alert('Feil ved lagring: ' + (e.response?.data?.message || e.message))
    }
}

const deletePipeline = async () => {
    if (!confirm('Slette pipelinen?')) return
    try {
        await api.deleteReaction(pipelineId)
        router.back()
    } catch (e: any) { alert(e.message) }
}

// --- CORE LOGIC ---
const updateTriggerOutputs = async (event: string, aggregateId?: string) => {
    const triggerNode = nodes.value.find((n: Node) => n.type === 'trigger')
    if (!triggerNode) return

    let outputs = TRIGGER_VAR_MAP[event] || ['timestamp']
    if (event === 'FormResponseSubmitted' && aggregateId) {
        const form = forms.value.find((f: any) => f.id === aggregateId)
        if (form?.schema?.fields) {
            outputs = [...outputs, ...form.schema.fields.map((f: any) => `ans_${f.name}`)]
        }
    }
    
    updateNodeData(triggerNode.id, { 
        event, 
        aggregateId, 
        outputs,
        availableEvents: Object.keys(TRIGGER_VAR_MAP),
        forms: forms.value,
        onUpdateEvent: (e: string) => updateTriggerOutputs(e, aggregateId),
        onUpdateAggregate: (a: string) => updateTriggerOutputs(event, a)
    })
}

const updateAllConnectedInputs = () => {
    nodes.value.forEach(node => {
        const connectedInputs = new Set(edges.value.filter(e => e.target === node.id).map(e => e.targetHandle))
        const current = Array.from(node.data.connectedInputs || []).sort().join(',')
        const next = Array.from(connectedInputs).sort().join(',')
        if (current !== next) updateNodeData(node.id, { connectedInputs })
    })
}

onConnect((params: any) => {
    addEdges([{ ...params, type: 'smoothstep', animated: true, markerEnd: { type: 'arrowclosed', color: '#000' }, style: { strokeWidth: 4, stroke: '#000' } }])
    nextTick(updateAllConnectedInputs)
})

const checkValidConnection = (connection: Connection) => {
    if (connection.source === connection.target) return false;
    const targetNode = findNode(connection.target);
    return targetNode?.type !== 'trigger';
}

const nodeErrorsMap = computed(() => {
    const errorMap: Record<string, string[]> = {}
    const trigger = nodes.value.find(n => n.type === 'trigger')
    nodes.value.forEach(node => {
        const errors: string[] = []
        if (trigger) {
            const visited = new Set<string>()
            const queue = [trigger.id]
            while (queue.length > 0) {
                const curr = queue.shift()!
                visited.add(curr)
                edges.value.filter(e => e.source === curr).forEach(e => { if (!visited.has(e.target)) queue.push(e.target) })
            }
            if (!visited.has(node.id)) errors.push('Ikke koblet til flyt')
        }
        if (node.type === 'action' && node.data.inputs) {
            node.data.inputs.forEach((inp: string) => {
                if (!edges.value.some(e => e.target === node.id && e.targetHandle === inp) && !node.data[inp]) {
                    errors.push(`Mangler ${inp}`)
                }
            })
        }
        if (errors.length > 0) errorMap[node.id] = errors
    })
    return errorMap
})

const validationErrors = computed(() => {
    const errors: string[] = []
    if (!nodes.value.some(n => n.type === 'trigger')) errors.push('Mangler Start-node')
    Object.values(nodeErrorsMap.value).forEach(ne => errors.push(...ne))
    return Array.from(new Set(errors))
})

watch(nodeErrorsMap, (newMap) => {
    nodes.value.forEach(node => {
        const hasError = !!newMap[node.id]
        if (node.data.hasError !== hasError) updateNodeData(node.id, { hasError, errorMessages: newMap[node.id] || [] })
    })
}, { deep: true })

// --- INTERACTION ---
const onDrop = (event: DragEvent) => {
    const dataStr = event.dataTransfer?.getData('application/vueflow')
    if (!dataStr) return
    const { type, nodeClass } = JSON.parse(dataStr)
    const position = screenToFlowCoordinate({ x: event.clientX, y: event.clientY })
    const id = `node_${Date.now()}`

    const newNode: Node = {
        id, position, type: nodeClass, dimensions: { width: 200, height: 150 },
        data: { 
            label: type, 
            inputs: ACTION_DEFS[type]?.inputs || (nodeClass === 'logic' ? ['v1', 'v2'] : []), 
            outputs: ACTION_DEFS[type]?.outputs || (nodeClass === 'logic' ? ['true', 'false'] : ['result']),
            referenceData: referenceData.value,
            onUpdateData: (k: string, v: any) => updateNodeData(id, { [k]: v })
        }
    }
    addNodes([newNode])
}

const usePreset = async (preset: any) => {
    if (nodes.value.length > 1 && !confirm('Erstatt nåværende design?')) return
    
    // 1. Load Nodes
    nodes.value = preset.nodes.map((n: any) => ({ 
        ...n, 
        data: { 
            ...n.data, 
            referenceData: referenceData.value, 
            onUpdateData: (k: string, v: any) => updateNodeData(n.id, { [k]: v }) 
        } 
    }))
    
    await nextTick()

    // 2. Load Edges
    edges.value = preset.edges.map((e: any) => ({ 
        ...e, 
        type: 'smoothstep', 
        animated: true, 
        markerEnd: { type: 'arrowclosed', color: '#000' }, 
        style: { strokeWidth: 4, stroke: '#000' } 
    }))
    
    nextTick(() => setTimeout(safeFitView, 500))
}

const deleteSelected = () => {
    const sn = nodes.value.filter((n: any) => n.selected && n.type !== 'trigger')
    const se = edges.value.filter((e: any) => e.selected)
    if (sn.length) removeNodes(sn); if (se.length) removeEdges(se)
    nextTick(updateAllConnectedInputs)
}

onMounted(async () => {
    await fetchPipeline()
    if (nodes.value.length === 0) {
        addNodes([{
            id: 'trigger', type: 'trigger', position: { x: 50, y: 50 }, dimensions: { width: 220, height: 150 },
            data: { event: 'MembershipCreated', outputs: TRIGGER_VAR_MAP['MembershipCreated'], availableEvents: Object.keys(TRIGGER_VAR_MAP), forms: forms.value, 
            onUpdateEvent: (e: string) => updateTriggerOutputs(e), onUpdateAggregate: (a: string) => updateTriggerOutputs(findNode('trigger')?.data.event, a) }
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
            <BButton v-if="!isHistoryMode" @click="deletePipeline" variant="danger" class="text-xs py-2 px-6">SLETT PIPELINE</BButton>
            <BButton v-if="!isHistoryMode" @click="save" variant="primary" class="text-xs py-2 px-10 shadow-[4px_4px_0px_0px_white]">PUBLISER ENDRINGER</BButton>
        </div>
    </header>

    <div class="flex-1 flex overflow-hidden min-h-0 relative" ref="containerRef">
        <div v-if="loading" class="absolute inset-0 z-[60] bg-orange-50/80 backdrop-blur-sm flex items-center justify-center">
            <div class="flex flex-col items-center gap-4">
                <div class="w-16 h-16 border-8 border-black border-t-yellow-400 animate-spin"></div>
                <p class="font-black uppercase italic tracking-tighter">Henter Pipeline...</p>
            </div>
        </div>
        
        <aside v-if="!isHistoryMode" class="w-80 bg-white border-r-8 border-black p-6 space-y-10 overflow-y-auto z-40 shadow-[8px_0px_0px_0px_rgba(0,0,0,0.1)] shrink-0">
            <section class="space-y-4">
                <h4 class="font-black uppercase text-xs border-b-4 border-black pb-2">Maler (Presets)</h4>
                <div class="space-y-2">
                    <button v-for="p in PIPELINE_PRESETS" :key="p.name" 
                        @click="usePreset(p)"
                        class="w-full text-left p-3 border-4 border-black hover:bg-yellow-400 transition-colors group"
                    >
                        <div class="font-black uppercase text-[10px] italic">{{ p.name }}</div>
                        <div class="text-[8px] font-bold text-gray-500 mt-1 group-hover:text-black">{{ p.description }}</div>
                    </button>
                </div>
            </section>

            <section class="space-y-4">
                <div class="flex justify-between items-center border-b-4 border-black pb-2">
                    <h4 class="font-black uppercase text-xs">Navigator</h4>
                    <BButton @click="safeFitView" variant="ghost" class="text-[8px] py-0 px-2 border-black">SENTRÉR</BButton>
                </div>
                <div class="max-h-64 overflow-y-auto space-y-1 border-2 border-black p-2 bg-gray-50">
                    <button 
                        v-for="node in nodes" :key="node.id"
                        @click="focusNode(node.id)"
                        class="w-full text-left px-2 py-1 text-[10px] font-black uppercase italic hover:bg-yellow-400 border border-transparent hover:border-black transition-all flex justify-between items-center group"
                    >
                        <span class="truncate">
                            <span v-if="node.type === 'trigger'" class="text-purple-600">⚡</span>
                            <span v-else-if="node.type === 'logic'" class="text-orange-600">?</span>
                            <span v-else-if="node.type === 'action'" class="text-blue-600">+</span>
                            <span v-else-if="node.type === 'code'" class="text-green-600">JS</span>
                            {{ node.data?.label || node.data?.type || (node.type === 'trigger' ? 'START' : node.id) }}
                        </span>
                        <span class="text-[8px] opacity-0 group-hover:opacity-100 font-mono">GO →</span>
                    </button>
                </div>
            </section>

            <section class="space-y-4">
                <h4 class="font-black uppercase text-xs border-b-4 border-black pb-2">Logikk</h4>
                <div class="grid grid-cols-2 gap-2">
                    <button draggable="true" @click="onDrop({ clientX: 400, clientY: 200, dataTransfer: { getData: () => JSON.stringify({ type: 'if', nodeClass: 'logic' }) } } as any)" class="brutalist-btn bg-yellow-100 text-[10px] p-2 hover:bg-yellow-200 font-black italic cursor-grab">IF / THEN</button>
                    <button draggable="true" @click="onDrop({ clientX: 400, clientY: 200, dataTransfer: { getData: () => JSON.stringify({ type: 'list_filter', nodeClass: 'logic' }) } } as any)" class="brutalist-btn bg-blue-100 text-[10px] p-2 hover:bg-blue-200 font-black italic cursor-grab">FILTER (LIST)</button>
                </div>
            </section>
            <section class="space-y-4">
                <h4 class="font-black uppercase text-xs border-b-4 border-black pb-2">Handlinger</h4>
                <div class="space-y-2">
                    <button v-for="type in Object.keys(ACTION_DEFS)" :key="type" 
                        draggable="true"
                        @click="onDrop({ clientX: 400, clientY: 200, dataTransfer: { getData: () => JSON.stringify({ type, nodeClass: 'action' }) } } as any)"
                        class="brutalist-btn w-full bg-white text-left text-[10px] p-2 hover:bg-blue-50 font-black italic cursor-grab"
                    >
                        + {{ type }}
                    </button>
                </div>
            </section>
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
                v-if="dataLoaded"
                :key="flowKey"
                id="main"
                :nodes="nodes"
                :edges="edges"
                :node-types="nodeTypes"
                :is-valid-connection="checkValidConnection"
                class="brutalist-flow"
            >
                <Background pattern-color="#000" :gap="20" />
                <Controls position="bottom-right" />
                <MiniMap 
                    pannable 
                    zoomable 
                    :node-color="getNodeColor"
                    :node-stroke-color="'#000'"
                    class="!border-4 !border-black !rounded-none !shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] !bg-white" 
                />
            </VueFlow>
        </main>
    </div>
  </div>
</template>

<style>
.brutalist-flow { width: 100%; height: 100%; min-height: 500px; }
.brutalist-flow .vue-flow__node { padding: 0; border: none; background: transparent; }
.brutalist-flow .vue-flow__edge-path { stroke: black !important; stroke-width: 6 !important; }
.brutalist-flow .vue-flow__edge.selected .vue-flow__edge-path { stroke: #f87171 !important; }
.brutalist-flow .vue-flow__handle { width: 14px; height: 14px; border: 3px solid black; background: white; }
.brutalist-flow .vue-flow__controls { border: 4px solid black; box-shadow: 8px 8px 0px 0px rgba(0,0,0,1); }
</style>
