<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
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

const newReaction = ref({
    trigger_event: 'MembershipCreated',
    action_type: 'SEND_EMAIL',
    config: {
        template_name: 'Velkomst'
    }
})

const fetchReactions = async () => {
    if (!selectedOrg.value) return
    loading.value = true
    try {
        reactions.value = await api.getReactions(selectedOrg.value)
    } catch (e) {
        console.error(e)
    } finally {
        loading.value = false
    }
}

const createReaction = async () => {
    try {
        await api.createReaction({
            org_id: selectedOrg.value,
            ...newReaction.value
        })
        showCreate.value = false
        fetchReactions()
    } catch (e: any) {
        alert(e.message)
    }
}

watch(selectedOrg, fetchReactions)

onMounted(() => {
    if (props.orgs.length > 0) {
        selectedOrg.value = props.orgs[0].id
    }
})
</script>

<template>
    <div class="space-y-8">
        <header class="border-b-8 border-black pb-4">
            <h2 class="text-4xl font-black uppercase tracking-tighter italic">Event Pipelines</h2>
            <p class="text-xs font-bold uppercase text-gray-500">Konfigurer automatiske reaksjoner på systemhendelser</p>
        </header>

        <div class="flex flex-wrap gap-6 items-end bg-black text-white p-6 shadow-[8px_8px_0px_0px_rgba(0,0,0,0.3)]">
            <BSelect v-model="selectedOrg" label="Velg Organisasjon" class="bg-black text-white border-white min-w-[300px]">
                <option v-for="org in orgs" :key="org.id" :value="org.id">{{ org.name }}</option>
            </BSelect>
            <BButton @click="showCreate = true" variant="primary" class="italic text-xs py-2">+ NY PIPELINE</BButton>
        </div>

        <div v-if="showCreate">
            <BCard class="bg-yellow-50 max-w-2xl mx-auto">
                <h3 class="text-xl font-black uppercase mb-6 italic">Opprett ny reaksjon</h3>
                <div class="space-y-6">
                    <BSelect v-model="newReaction.trigger_event" label="Hendelse (Trigger)">
                        <option value="MembershipCreated">Nytt Medlemskap</option>
                        <option value="FeeGenerated">Faktura Generert</option>
                        <option value="PaymentReceived">Betaling Mottatt</option>
                    </BSelect>

                    <BSelect v-model="newReaction.action_type" label="Handling (Action)">
                        <option value="SEND_EMAIL">Send E-post</option>
                        <option value="WEBHOOK">Webhook (HTTP POST)</option>
                    </BSelect>

                    <div v-if="newReaction.action_type === 'SEND_EMAIL'" class="p-4 border-2 border-black bg-white">
                        <BInput v-model="newReaction.config.template_name" label="E-post Mal Navn" placeholder="F.eks. Velkomst" />
                    </div>

                    <div v-if="newReaction.action_type === 'WEBHOOK'" class="p-4 border-2 border-black bg-white">
                        <BInput v-model="newReaction.config.url" label="Webhook URL" placeholder="https://api.mittsystem.no/webhook" />
                    </div>
                </div>
                <div class="mt-8 flex gap-4">
                    <BButton @click="createReaction" variant="primary" class="flex-1 italic">LAGRE PIPELINE</BButton>
                    <BButton @click="showCreate = false" variant="secondary">AVBRYT</BButton>
                </div>
            </BCard>
        </div>

        <div class="grid grid-cols-1 gap-6">
            <BCard v-for="r in reactions" :key="r.id" :class="r.is_inherited ? 'bg-gray-50 opacity-80' : 'bg-white'">
                <div class="flex justify-between items-start">
                    <div class="space-y-2">
                        <div class="flex items-center gap-3">
                            <BBadge class="bg-purple-400">TRIGGER: {{ r.trigger_event }}</BBadge>
                            <span v-if="r.is_inherited" class="text-[10px] font-black uppercase text-blue-600 italic">Arvet fra overordnet enhet</span>
                        </div>
                        <h3 class="text-2xl font-black uppercase tracking-tighter">ACTION: {{ r.action_type }}</h3>
                        <div class="bg-black text-green-400 p-3 font-mono text-[10px] border-2 border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
                            CONFIG: {{ JSON.stringify(r.config) }}
                        </div>
                    </div>
                </div>
            </BCard>

            <div v-if="reactions.length === 0 && !loading" class="py-20 text-center border-8 border-black border-dashed">
                <p class="font-black uppercase text-gray-400 italic">Ingen aktive pipelines for denne organisasjonen</p>
            </div>
        </div>
    </div>
</template>
