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

const { nodes, edges, onConnect, addEdges, addNodes, toObject } = useVueFlow()

const loading = ref(false)

onConnect((params) => {
    addEdges([params])
})

const ACTION_DEFS: Record<string, { inputs: string[], outputs: string[] }> = {
    FindOrg: { inputs: ['start_org_id', 'relation'], outputs: ['org_id', 'name'] },
    FindRole: { inputs: ['target_id', 'role_type'], outputs: ['user_id', 'email', 'name'] },
    SendEmail: { inputs: ['to_email', 'subject', 'body'], outputs: ['success'] },
    RegisterMember: { inputs: ['org_id', 'name', 'email'], outputs: ['membership_id'] },
}

const fetchPipeline = async () => {
    loading.value = true
    try {
        const orgId = route.query.org_id as string
        const data = await api.getReactions(orgId)
        const r = Array.isArray(data) ? data.find(item => item.id === pipelineId) : null
        
        if (r && r.config) {
            if (r.config.nodes) nodes.value = r.config.nodes
            if (r.config.edges) edges.value = r.config.edges
            
            // Dynamic Form Schema logic
            if (r.trigger_event === 'FormResponseSubmitted' && r.trigger_aggregate_id) {
                const formsData = await api.getForms()
                const form = Array.isArray(formsData) ? formsData.find(f => f.id === r.trigger_aggregate_id) : null
                
                if (form && form.schema && (form.schema as any).fields) {
                    const dynamicOutputs = [
                        'form_id', 'user_id', 'timestamp',
                        ...(form.schema as any).fields.map((f: any) => `ans_${f.name}`)
                    ]
                    const triggerNode = nodes.value.find(n => n.type === 'trigger')
                    if (triggerNode) triggerNode.data.outputs = dynamicOutputs
                }
            }
        }
    } catch (e) {
        console.error(e)
    } finally {
        loading.value = false
    }
}

const addActionNode = (type: string) => {
    const def = ACTION_DEFS[type]
    const newNode: Node = {
        id: `node_${nodes.value.length + 1}`,
        type: 'action',
        position: { x: Math.random() * 400, y: Math.random() * 400 },
        data: {
            label: type,
            inputs: def?.inputs || [],
            outputs: def?.outputs || []
        }
    }
    addNodes([newNode])
}

const addLogicNode = (type: 'if' | 'filter') => {
    const newNode: Node = {
        id: `node_${nodes.value.length + 1}`,
        type: 'logic',
        position: { x: Math.random() * 400, y: Math.random() * 400 },
        data: { type, inputs: ['v1', 'v2', 'operator'] }
    }
    addNodes([newNode])
}

const save = async () => {
    const flow = toObject()
    try {
        await api.updateReaction(pipelineId, flow)
        alert('Pipeline lagret og publisert!')
    } catch (e: any) {
        alert('Feil ved lagring: ' + e.message)
    }
}

onMounted(async () => {
    await fetchPipeline()
    if (nodes.value.length === 0) {
        addNodes([{ 
            id: 'trigger', 
            type: 'trigger', 
            position: { x: 50, y: 50 },
            data: { event: 'New Event', outputs: ['id', 'timestamp'] }
        }])
    }
})
</script>

<template>
  <div class="h-screen w-screen bg-orange-50 flex flex-col overflow-hidden font-mono text-black">
    <!-- Navbar -->
    <header class="bg-black text-white p-4 flex justify-between items-center border-b-8 border-black z-50">
        <div class="flex items-center gap-6">
            <BButton @click="router.back()" variant="ghost" class="text-white border-white text-xs py-1 hover:bg-white hover:text-black transition-colors italic">← TILBAKE</BButton>
            <div class="h-8 w-2 bg-yellow-400"></div>
            <h1 class="text-3xl font-black uppercase italic tracking-tighter leading-none">Visual Pipeline Builder</h1>
        </div>
        <div class="flex gap-4">
            <BButton @click="save" variant="primary" class="text-xs py-2 px-10 shadow-[4px_4px_0px_0px_white]">LAGRE ENDRINGER</BButton>
        </div>
    </header>

    <div class="flex-1 flex overflow-hidden">
        <!-- Sidebar -->
        <aside class="w-80 bg-white border-r-8 border-black p-6 space-y-10 overflow-y-auto z-40 shadow-[8px_0px_0px_0px_rgba(0,0,0,0.1)]">
            <section class="space-y-4">
                <h4 class="font-black uppercase text-xs border-b-4 border-black pb-2">Logikk</h4>
                <div class="grid grid-cols-2 gap-2">
                    <button @click="addLogicNode('if')" class="brutalist-btn bg-yellow-100 text-[10px] p-2 hover:bg-yellow-200">IF / THEN</button>
                    <button @click="addLogicNode('filter')" class="brutalist-btn bg-orange-100 text-[10px] p-2 hover:bg-orange-200">FILTER</button>
                </div>
            </section>

            <section class="space-y-4">
                <h4 class="font-black uppercase text-xs border-b-4 border-black pb-2">Handlinger</h4>
                <div class="space-y-2">
                    <button v-for="type in Object.keys(ACTION_DEFS)" :key="type" 
                        @click="addActionNode(type)"
                        class="brutalist-btn w-full bg-white text-left text-[10px] p-2 hover:bg-blue-50"
                    >
                        + {{ type }}
                    </button>
                </div>
            </section>

            <div class="brutalist-card bg-black text-white p-4 !shadow-none italic text-[8px] leading-relaxed">
                STATUS: Koble utganger (høyre) til innganger (venstre). Endringer i form-skjema oppdaterer Trigger-noden automatisk.
            </div>
        </aside>

        <!-- Canvas -->
        <main class="flex-1 relative bg-[radial-gradient(#000_1px,transparent_1px)] [background-size:20px_20px]">
            <VueFlow 
                :nodes="nodes" 
                :edges="edges" 
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
