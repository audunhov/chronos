<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { api } from '../services/api'
import BButton from './base/BButton.vue'
import BCard from './base/BCard.vue'
import BBadge from './base/BBadge.vue'
import BInput from './base/BInput.vue'
import BSelect from './base/BSelect.vue'

const props = defineProps<{
    orgs: { id: string, name: string }[]
}>()

const selectedOrg = ref('')
const reactions = ref<any[]>([])
const loading = ref(false)
const showCreate = ref(false)

// --- DAG PIPELINE BUILDER STATE ---
const triggerEvent = ref('MembershipCreated')

type NodeType = 'FindOrg' | 'FindOrgan' | 'FindRole' | 'Template' | 'SendEmail'

type PipelineNode = {
    id: string;
    type: NodeType;
    inputs: any;
}

const pipelineNodes = ref<PipelineNode[]>([])

const NODE_DEFS: Record<NodeType, { name: string, inputs: Record<string, string>, outputs: Record<string, string> }> = {
    FindOrg: {
        name: 'Finn Organisasjon',
        inputs: { start_org_id: 'Start Org-ID', relation: 'Relasjon (parent/child)' },
        outputs: { org_id: 'Funnet Org-ID', name: 'Org Navn' }
    },
    FindOrgan: {
        name: 'Finn Organ (Styre/Utvalg)',
        inputs: { org_id: 'Org-ID', organ_name: 'Navn på organ' },
        outputs: { organ_id: 'Funnet Organ-ID' }
    },
    FindRole: {
        name: 'Finn Person via Rolle',
        inputs: { target_id: 'Org/Organ-ID', role_type: 'Rollenavn (f.eks. leader)' },
        outputs: { user_id: 'Bruker-ID', email: 'E-post', name: 'Navn' }
    },
    Template: {
        name: 'Generer Tekst (Template)',
        inputs: { template_name: 'Mal-navn', user_id: 'Mål Bruker-ID', org_id: 'Mål Org-ID' },
        outputs: { subject: 'Emnefelt', body: 'Tekstkropp' }
    },
    SendEmail: {
        name: 'Send E-post',
        inputs: { to_email: 'Mottaker E-post', subject: 'Emne', body: 'Melding' },
        outputs: {}
    }
}

const TRIGGER_OUTPUTS: Record<string, string> = {
    user_id: 'Hendelse Bruker-ID',
    org_id: 'Hendelse Org-ID',
    timestamp: 'Tidspunkt'
}

const addNode = (type: NodeType) => {
    const id = `node_${pipelineNodes.value.length + 1}`
    const inputs: any = {}
    for (const key of Object.keys(NODE_DEFS[type].inputs)) {
        inputs[key] = { mode: 'static', value: '' }
    }
    pipelineNodes.value.push({ id, type, inputs })
}

const removeNode = (index: number) => {
    pipelineNodes.value.splice(index, 1)
}

// Compute available variables for a node based on its position
const getAvailableVariables = (nodeIndex: number) => {
    const vars: { label: string, value: string }[] = []
    
    // Add trigger outputs
    for (const [key, desc] of Object.entries(TRIGGER_OUTPUTS)) {
        vars.push({ label: `Trigger: ${desc}`, value: `trigger.${key}` })
    }

    // Add previous nodes outputs
    for (let i = 0; i < nodeIndex; i++) {
        const prevNode = pipelineNodes.value[i]
        if (!prevNode) continue
        const def = NODE_DEFS[prevNode.type]
        for (const [key, desc] of Object.entries(def.outputs)) {
            vars.push({ label: `[${prevNode.id}] ${def.name}: ${desc}`, value: `${prevNode.id}.${key}` })
        }
    }

    return vars
}
// ---------------------------------

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

const createReaction = async () => {
    try {
        await api.createReaction({
            org_id: selectedOrg.value,
            trigger_event: triggerEvent.value,
            action_type: 'PIPELINE_DAG',
            config: {
                nodes: pipelineNodes.value
            }
        })
        showCreate.value = false
        pipelineNodes.value = []
        fetchReactions()
    } catch (e: any) {
        alert(e.message)
    }
}

watch(selectedOrg, fetchReactions)

onMounted(() => {
    if (props.orgs && props.orgs.length > 0) {
        selectedOrg.value = props.orgs[0]!.id
    }
})
</script>

<template>
    <div class="space-y-8">
        <header class="border-b-8 border-black pb-4">
            <h2 class="text-4xl font-black uppercase tracking-tighter italic">Event Pipelines</h2>
            <p class="text-xs font-bold uppercase text-gray-500">Konfigurer reaksjoner via datakjeder (DAG)</p>
        </header>

        <div class="flex flex-wrap gap-6 items-end bg-black text-white p-6 shadow-[8px_8px_0px_0px_rgba(0,0,0,0.3)]">
            <BSelect v-model="selectedOrg" label="Velg Organisasjon" class="bg-black text-white border-white min-w-[300px]">
                <option v-for="org in orgs" :key="org.id" :value="org.id">{{ org.name }}</option>
            </BSelect>
            <BButton @click="showCreate = true" variant="primary" class="italic text-xs py-2">+ NY PIPELINE</BButton>
        </div>

        <div v-if="showCreate" class="space-y-6">
            <BCard class="bg-yellow-50 max-w-4xl mx-auto !p-0">
                <div class="bg-black text-white p-6 flex justify-between items-center">
                    <h3 class="text-2xl font-black uppercase italic tracking-tighter">Pipeline Builder</h3>
                    <BButton @click="showCreate = false" variant="danger" class="text-[10px] py-1">LUKK</BButton>
                </div>
                
                <div class="p-8 space-y-8">
                    <!-- Trigger Selection -->
                    <div class="border-4 border-black p-6 bg-white relative">
                        <BBadge class="absolute -top-3 left-4 bg-purple-400">1. START (TRIGGER)</BBadge>
                        <BSelect v-model="triggerEvent" label="Når dette skjer i systemet:">
                            <option value="MembershipCreated">Nytt Medlemskap Opprettet</option>
                            <option value="FeeGenerated">Faktura Generert</option>
                            <option value="RoleAssigned">Rolle Tildelt</option>
                        </BSelect>
                        <div class="mt-4 text-[10px] text-gray-500 font-mono">
                            Eksponerer: trigger.user_id, trigger.org_id, trigger.timestamp
                        </div>
                    </div>

                    <!-- Nodes Chain -->
                    <div class="space-y-6 relative border-l-8 border-black ml-8 pl-8 py-4">
                        <div v-for="(node, index) in pipelineNodes" :key="node.id" class="border-4 border-black p-6 bg-white relative group">
                            <!-- Connection Line Visual -->
                            <div class="absolute w-8 h-2 bg-black top-1/2 -left-9"></div>

                            <button @click="removeNode(index)" class="absolute -top-4 -right-4 bg-red-500 text-white w-8 h-8 border-4 border-black font-black hover:bg-red-400 shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">X</button>
                            
                            <BBadge class="absolute -top-3 left-4 bg-blue-400">Trinn {{ index + 2 }}: {{ node.id }}</BBadge>
                            <h4 class="text-xl font-black uppercase italic tracking-tighter mt-2 mb-6">{{ NODE_DEFS[node.type].name }}</h4>
                            
                            <div class="space-y-4">
                                <div v-for="(desc, inputKey) in NODE_DEFS[node.type].inputs" :key="inputKey" class="grid grid-cols-1 md:grid-cols-3 gap-4 items-end border-b-2 border-dashed border-gray-200 pb-4">
                                    <div class="col-span-1">
                                        <label class="block text-[10px] font-black uppercase">{{ desc }}</label>
                                        <code class="text-[8px] text-gray-400">{{ inputKey }}</code>
                                    </div>
                                    <div class="col-span-2 flex gap-2">
                                        <select v-model="node.inputs[inputKey].mode" class="border-4 border-black p-2 font-mono text-xs focus:outline-none focus:bg-blue-50">
                                            <option value="static">Statisk verdi</option>
                                            <option value="ref">Referer variabel</option>
                                        </select>
                                        
                                        <input 
                                            v-if="node.inputs[inputKey].mode === 'static'" 
                                            v-model="node.inputs[inputKey].value" 
                                            class="border-4 border-black p-2 font-mono text-xs w-full focus:outline-none focus:bg-blue-50"
                                            placeholder="..."
                                        />
                                        
                                        <select 
                                            v-else 
                                            v-model="node.inputs[inputKey].value" 
                                            class="border-4 border-black p-2 font-mono text-xs w-full bg-yellow-100 focus:outline-none"
                                        >
                                            <option value="">Velg variabel...</option>
                                            <option v-for="v in getAvailableVariables(index)" :key="v.value" :value="v.value">
                                                {{ v.label }} ({{ v.value }})
                                            </option>
                                        </select>
                                    </div>
                                </div>
                            </div>
                            
                            <div v-if="Object.keys(NODE_DEFS[node.type].outputs).length > 0" class="mt-6 pt-4 border-t-4 border-black">
                                <span class="text-[10px] font-black uppercase text-green-600 block mb-1">Eksponerer nye variabler:</span>
                                <div class="flex gap-2 flex-wrap">
                                    <span v-for="(outDesc, outKey) in NODE_DEFS[node.type].outputs" :key="outKey" class="bg-gray-100 border-2 border-black px-2 py-1 text-[8px] font-mono">
                                        {{ node.id }}.{{ outKey }}
                                    </span>
                                </div>
                            </div>
                        </div>

                        <!-- Add Node Actions -->
                        <div class="relative mt-8">
                            <div class="absolute w-8 h-2 bg-black top-1/2 -left-9"></div>
                            <div class="bg-white border-4 border-black p-6 border-dashed flex flex-col items-center gap-4">
                                <h4 class="font-black uppercase text-sm italic">Legg til neste steg</h4>
                                <div class="flex flex-wrap gap-2 justify-center">
                                    <BButton v-for="(def, type) in NODE_DEFS" :key="type" @click="addNode(type as NodeType)" variant="ghost" class="text-[10px] py-1 border-dashed">+ {{ def.name }}</BButton>
                                </div>
                            </div>
                        </div>
                    </div>

                </div>

                <div class="bg-gray-100 p-6 border-t-8 border-black flex gap-4">
                    <BButton @click="createReaction" variant="success" class="flex-1 text-2xl py-6 italic" :disabled="pipelineNodes.length === 0">
                        LAGRE PIPELINE
                    </BButton>
                </div>
            </BCard>
        </div>

        <!-- LIST EXISTING REACTIONS -->
        <div class="grid grid-cols-1 gap-6">
            <BCard v-for="r in reactions" :key="r.id" :class="r.is_inherited ? 'bg-gray-50 opacity-80 border-dashed' : 'bg-white border-solid'">
                <div class="flex justify-between items-start mb-6 border-b-4 border-black pb-4">
                    <div class="flex items-center gap-4">
                        <BBadge class="bg-purple-400 text-lg">TRIGGER: {{ r.trigger_event }}</BBadge>
                        <span v-if="r.is_inherited" class="text-xs font-black uppercase text-blue-600 italic">↑ Arvet nedover</span>
                    </div>
                    <BBadge class="bg-black text-white">{{ r.action_type }}</BBadge>
                </div>
                
                <div v-if="r.action_type === 'PIPELINE_DAG'" class="space-y-2 ml-4 border-l-4 border-black pl-4">
                    <div v-for="(node, idx) in r.config.nodes" :key="idx" class="flex items-center gap-3">
                        <span class="font-black text-xl leading-none">{{ Number(idx) + 1 }}.</span>
                        <div class="bg-gray-100 border-2 border-black px-3 py-1 flex-1 flex justify-between items-center">
                            <span class="font-bold uppercase text-xs">{{ NODE_DEFS[node.type as NodeType]?.name || node.type }}</span>
                            <span class="font-mono text-[8px] text-gray-500">{{ node.id }}</span>
                        </div>
                    </div>
                </div>
                <div v-else class="bg-gray-100 p-4 font-mono text-xs border-2 border-black">
                    LEGACY CONFIG: {{ r.config }}
                </div>
            </BCard>

            <div v-if="reactions.length === 0 && !loading" class="py-32 text-center border-8 border-black border-dashed bg-white">
                <h3 class="text-6xl font-black uppercase italic opacity-10 mb-4">Ingen dataflyt</h3>
                <p class="font-bold uppercase text-gray-400 tracking-widest">Ingen pipelines aktive for denne organisasjonen.</p>
            </div>
        </div>
    </div>
</template>
