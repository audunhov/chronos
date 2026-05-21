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
const triggerAggregateID = ref('')
const forms = ref<any[]>([])
const reactions = ref<any[]>([])
const loading = ref(false)
const showCreate = ref(false)

// --- DAG PIPELINE BUILDER STATE ---
const triggerEvent = ref('MembershipCreated')

type NodeType = 'FindOrg' | 'FindOrgan' | 'FindRole' | 'Template' | 'SendEmail' | 'CreateForm' | 'RegisterMember'

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
    },
    CreateForm: {
        name: 'Opprett Skjema',
        inputs: { org_id: 'Org-ID', title: 'Skjematittel', schema: 'JSON Schema' },
        outputs: { form_id: 'Skjema-ID' }
    },
    RegisterMember: {
        name: 'Registrer Medlem',
        inputs: { org_id: 'Org-ID', name: 'Fullt Navn', email: 'E-post' },
        outputs: { membership_id: 'Medlems-ID' }
    }
}

const TRIGGER_OUTPUTS_BY_EVENT: Record<string, Record<string, string>> = {
    MembershipCreated: {
        user_id: 'Bruker-ID',
        org_id: 'Organisasjons-ID',
        role: 'Rolle',
        timestamp: 'Tidspunkt'
    },
    FeeGenerated: {
        id: 'Medlemskap-ID',
        amount: 'Beløp (øre)',
        period: 'Periode',
        timestamp: 'Tidspunkt'
    },
    RoleAssigned: {
        user_id: 'Bruker-ID',
        org_id: 'Organisasjons-ID',
        organ_id: 'Organ-ID',
        role_type: 'Rolletype',
        timestamp: 'Tidspunkt'
    },
    OrganizationCreated: {
        id: 'Org-ID',
        name: 'Navn',
        path: 'Sti',
        timestamp: 'Tidspunkt'
    },
    OrganCreated: {
        id: 'Organ-ID',
        org_id: 'Organisasjons-ID',
        name: 'Navn',
        timestamp: 'Tidspunkt'
    },
    FormCreated: {
        id: 'Skjema-ID',
        org_id: 'Organisasjons-ID',
        title: 'Tittel',
        timestamp: 'Tidspunkt'
    },
    FormResponseSubmitted: {
        form_id: 'Skjema-ID',
        user_id: 'Bruker-ID',
        answers: 'Svar (Objekt)',
        timestamp: 'Tidspunkt'
    }
}

const currentTriggerOutputs = computed(() => {
    return TRIGGER_OUTPUTS_BY_EVENT[triggerEvent.value] || {
        user_id: 'Bruker-ID',
        org_id: 'Org-ID',
        timestamp: 'Tidspunkt'
    }
})

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
    for (const [key, desc] of Object.entries(currentTriggerOutputs.value)) {
        vars.push({ label: `Trigger: ${desc}`, value: `trigger.${key}` })
        
        if (triggerEvent.value === 'FormResponseSubmitted' && key === 'answers') {
            vars.push({ label: 'Trigger: Svar (Fullt JSON)', value: 'trigger.answers' })
        }
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

const fetchForms = async () => {
    if (!selectedOrg.value) return
    try {
        const data = await api.getForms(selectedOrg.value)
        forms.value = Array.isArray(data) ? data : []
    } catch (e) {
        console.error(e)
    }
}

const createReaction = async () => {
    try {
        await api.createReaction({
            org_id: selectedOrg.value,
            trigger_event: triggerEvent.value,
            trigger_aggregate_id: triggerAggregateID.value || undefined,
            action_type: 'PIPELINE_DAG',
            config: {
                nodes: pipelineNodes.value
            }
        })
        showCreate.value = false
        pipelineNodes.value = []
        triggerAggregateID.value = ''
        fetchReactions()
    } catch (e: any) {
        alert(e.message)
    }
}

watch(selectedOrg, () => {
    fetchReactions()
    fetchForms()
})

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
            <BCard class="bg-yellow-50 max-w-4xl mx-auto !p-0 overflow-hidden">
                <div class="bg-black text-white p-6 flex justify-between items-center">
                    <h3 class="text-2xl font-black uppercase italic tracking-tighter">Pipeline Builder</h3>
                    <BButton @click="showCreate = false" variant="danger" class="text-[10px] py-1">LUKK</BButton>
                </div>
                
                <div class="p-8 space-y-8">
                    <!-- Trigger Selection -->
                    <div class="border-4 border-black p-6 bg-white relative">
                        <BBadge class="absolute -top-3 left-4 bg-purple-400">1. START (TRIGGER)</BBadge>
                        
                        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <BSelect v-model="triggerEvent" label="Når dette skjer i systemet:">
                                <option value="MembershipCreated">Nytt Medlemskap Opprettet</option>
                                <option value="FeeGenerated">Faktura Generert</option>
                                <option value="RoleAssigned">Rolle Tildelt</option>
                                <option value="OrganizationCreated">Ny Organisasjon Opprettet</option>
                                <option value="OrganCreated">Nytt Organ Opprettet</option>
                                <option value="FormCreated">Nytt Skjema Publisert</option>
                                <option value="FormResponseSubmitted">Skjema Besvart</option>
                            </BSelect>

                            <BSelect 
                                v-if="triggerEvent === 'FormResponseSubmitted'" 
                                v-model="triggerAggregateID" 
                                label="Velg spesifikt skjema (Valgfri)"
                            >
                                <option value="">Alle skjemaer</option>
                                <option v-for="f in forms" :key="f.id" :value="f.id">{{ f.title }}</option>
                            </BSelect>
                        </div>

                        <div class="mt-4 flex flex-wrap gap-2">
                            <span class="text-[10px] text-gray-500 font-black uppercase w-full mb-1">Eksponerer:</span>
                            <code v-for="(desc, key) in currentTriggerOutputs" :key="key" class="text-[8px] bg-gray-100 px-1 border border-black" :title="desc">
                                trigger.{{ key }}
                            </code>
                        </div>
                    </div>

                    <!-- Nodes Chain -->
                    <div class="space-y-6 relative border-l-8 border-black ml-8 pl-8 py-4">
                        <div v-for="(node, index) in pipelineNodes" :key="node.id" class="border-4 border-black p-6 bg-white relative group">
                            <!-- Connection Line Visual -->
                            <div class="absolute w-8 h-2 bg-black top-1/2 -left-9"></div>

                            <button @click="removeNode(index)" class="absolute -top-4 -right-4 bg-red-500 text-white w-8 h-8 border-4 border-black font-black flex items-center justify-center hover:bg-red-400 shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">X</button>
                            
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
                                    <BButton v-for="(def, type) in NODE_DEFS" :key="type" @click="addNode(type as NodeType)" variant="ghost" class="text-[10px] py-1 border-dashed italic font-bold">+ {{ def.name }}</BButton>
                                </div>
                            </div>
                        </div>
                    </div>

                </div>

                <div class="bg-gray-100 p-6 border-t-8 border-black flex gap-4">
                    <BButton @click="createReaction" variant="success" class="flex-1 text-2xl py-6 italic font-black" :disabled="pipelineNodes.length === 0">
                        LAGRE PIPELINE
                    </BButton>
                </div>
            </BCard>
        </div>

        <!-- LIST EXISTING REACTIONS -->
        <div class="grid grid-cols-1 gap-6">
            <BCard v-for="r in reactions" :key="r.id" :class="r.is_inherited ? 'bg-gray-50 opacity-80 border-dashed' : 'bg-white border-solid'">
                <div class="flex justify-between items-start mb-6 border-b-4 border-black pb-4">
                    <div class="flex flex-col gap-2">
                        <div class="flex items-center gap-4">
                            <BBadge class="bg-purple-400 text-lg italic">TRIGGER: {{ r.trigger_event }}</BBadge>
                            <span v-if="r.is_inherited" class="text-xs font-black uppercase text-blue-600 italic">↑ Arvet nedover</span>
                        </div>
                        <div v-if="r.trigger_aggregate_id" class="text-[10px] font-mono text-gray-500">
                            ID-FILTER: {{ r.trigger_aggregate_id }}
                        </div>
                    </div>
                    <div class="flex gap-4 items-center">
                        <router-link :to="'/admin/pipelines/edit/' + r.id">
                            <BButton variant="primary" class="text-[10px] py-1 px-4 italic font-black underline">ÅPNE VISUELL EDITOR</BButton>
                        </router-link>
                        <BBadge class="bg-black text-white px-4">{{ r.action_type }}</BBadge>
                    </div>
                </div>
                
                <div v-if="r.action_type === 'PIPELINE_DAG'" class="space-y-2 ml-4 border-l-4 border-black pl-4">
                    <div v-for="(node, idx) in r.config.nodes" :key="idx" class="flex items-center gap-3">
                        <span class="font-black text-xl leading-none italic">{{ Number(idx) + 1 }}.</span>
                        <div class="bg-gray-100 border-2 border-black px-3 py-1 flex-1 flex justify-between items-center">
                            <span class="font-black uppercase text-xs italic tracking-tight">{{ NODE_DEFS[node.type as NodeType]?.name || node.type }}</span>
                            <span class="font-mono text-[8px] text-gray-500">{{ node.id }}</span>
                        </div>
                    </div>
                </div>
                <div v-else class="bg-gray-100 p-4 font-mono text-xs border-2 border-black">
                    LEGACY CONFIG: {{ r.config }}
                </div>
            </BCard>

            <div v-if="reactions.length === 0 && !loading" class="py-32 text-center border-8 border-black border-dashed bg-white">
                <h3 class="text-6xl font-black uppercase italic opacity-10 mb-4 tracking-tighter">Ingen dataflyt</h3>
                <p class="font-bold uppercase text-gray-400 tracking-widest">Ingen pipelines aktive for denne organisasjonen.</p>
            </div>
        </div>
    </div>
</template>
