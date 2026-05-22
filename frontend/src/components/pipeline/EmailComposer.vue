<script setup lang="ts">
import { ref } from 'vue'
import { api } from '../../services/api'
import BCard from '../base/BCard.vue'
import BButton from '../base/BButton.vue'
import BInput from '../base/BInput.vue'

const props = defineProps<{
    recipientIds: string[];
    recipientNames: string; // e.g. "Audun, Bente, ..." or "All Board Members"
    onClose: () => void;
}>()

const subject = ref('')
const body = ref('')
const loading = ref(false)
const success = ref(false)

const send = async () => {
    if (!subject.value || !body.value) return
    loading.value = true
    try {
        await api.sendEmail(props.recipientIds, subject.value, body.value)
        success.value = true
        setTimeout(() => {
            props.onClose()
        }, 2000)
    } catch (e: any) {
        alert('Feil ved sending: ' + e.message)
    } finally {
        loading.value = false
    }
}
</script>

<template>
    <div class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm">
        <BCard class="w-full max-w-2xl !p-0 overflow-hidden bg-white border-8 border-black shadow-[16px_16px_0px_0px_rgba(0,0,0,1)]">
            <div class="bg-blue-600 text-white p-6 border-b-8 border-black flex justify-between items-center">
                <div>
                    <h3 class="text-2xl font-black uppercase italic tracking-tighter">Send E-post</h3>
                    <p class="text-[10px] font-bold opacity-80 mt-1">Mottakere: {{ recipientNames }} ({{ recipientIds.length }})</p>
                </div>
                <button @click="onClose" class="bg-black text-white w-10 h-10 border-4 border-white font-black hover:bg-white hover:text-black transition-colors">X</button>
            </div>

            <div v-if="success" class="p-20 text-center space-y-6">
                <h4 class="text-5xl font-black text-green-600 uppercase italic">SENDT!</h4>
                <p class="font-bold uppercase tracking-widest">E-postene er lagt i utboksen og vil bli sendt fortløpende.</p>
            </div>

            <div v-else class="p-8 space-y-6">
                <BInput 
                    v-model="subject" 
                    label="Emnefelt" 
                    placeholder="Viktig melding fra admin..." 
                    class="w-full"
                />

                <div class="flex flex-col">
                    <label class="text-xs font-black uppercase mb-1">Melding (HTML støttes)</label>
                    <textarea 
                        v-model="body" 
                        class="w-full h-64 bg-orange-50 border-4 border-black p-4 font-mono text-xs focus:outline-none focus:bg-yellow-50 resize-none"
                        placeholder="Skriv din melding her..."
                    ></textarea>
                </div>

                <div class="pt-4 flex gap-4">
                    <BButton 
                        @click="send" 
                        :disabled="loading || !subject || !body" 
                        variant="primary" 
                        class="flex-1 text-xl py-4 font-black italic"
                    >
                        {{ loading ? 'SENDER...' : 'SEND MELDING' }}
                    </BButton>
                    <BButton @click="onClose" variant="secondary" class="px-10">AVBRYT</BButton>
                </div>
            </div>
        </BCard>
    </div>
</template>
