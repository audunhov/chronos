<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../services/api'
import type { Form } from '../api'
import BCard from './base/BCard.vue'
import BButton from './base/BButton.vue'
import BInput from './base/BInput.vue'
import BSelect from './base/BSelect.vue'

const props = defineProps<{
    isAdmin?: boolean
}>()

const forms = ref<Form[]>([])
const loading = ref(false)
const selectedForm = ref<Form | null>(null)
const answers = ref<Record<string, any>>({})
const submitted = ref(false)

// Form Creation state
const showCreator = ref(false)
const newFormTitle = ref('')
const newFormOrg = ref('')
const newFormFields = ref<{ name: string, label: string, type: 'text' | 'textarea' | 'select', options?: string }[]>([])
const organizations = ref<string[]>([])

const addField = () => {
    newFormFields.value.push({ name: '', label: '', type: 'text' })
}

const removeField = (index: number) => {
    newFormFields.value.splice(index, 1)
}

const createForm = async () => {
    if (!newFormTitle.value || !newFormOrg.value) return
    
    const schema = {
        fields: newFormFields.value.map(f => ({
            ...f,
            options: f.options ? f.options.split(',').map(s => s.trim()) : undefined
        }))
    }

    try {
        await api.createForm({
            title: newFormTitle.value,
            org_id: newFormOrg.value,
            schema
        })
        showCreator.value = false
        newFormTitle.value = ''
        newFormOrg.value = ''
        newFormFields.value = []
        fetchForms()
    } catch (e: any) {
        alert(e.message)
    }
}

const fetchForms = async () => {
    loading.value = true
    try {
        const data = await api.getForms()
        forms.value = Array.isArray(data) ? data : []
    } catch (e) {
        console.error(e)
    } finally {
        loading.value = false
    }
}

const fetchOrgs = async () => {
    try {
        const data = await api.getOrganizations()
        organizations.value = Array.isArray(data) ? data : []
    } catch (e) {
        console.error(e)
    }
}

const startForm = (form: Form) => {
    selectedForm.value = form
    // Initialize answers for all fields in the schema
    const initialAnswers: Record<string, any> = {}
    if (form.schema && (form.schema as any).fields) {
        (form.schema as any).fields.forEach((f: any) => {
            initialAnswers[f.name] = ''
        })
    }
    answers.value = initialAnswers
    submitted.value = false
}

const submit = async () => {
    if (!selectedForm.value) return
    try {
        await api.submitForm(selectedForm.value.id!, answers.value)
        submitted.value = true
        setTimeout(() => {
            selectedForm.value = null
        }, 2000)
    } catch (e: any) {
        alert(e.message)
    }
}

onMounted(() => {
    fetchForms()
    if (props.isAdmin) fetchOrgs()
})
</script>

<template>
    <div class="space-y-12">
        <div class="flex justify-between items-end border-b-8 border-black pb-4">
            <div>
                <h2 class="text-4xl font-black uppercase tracking-tighter italic">Undersøkelser</h2>
                <p class="text-xs font-bold uppercase text-gray-500">Datainnsamling og tilbakemeldinger</p>
            </div>
            <BButton v-if="props.isAdmin && !showCreator" @click="showCreator = true" variant="primary" class="text-xs py-2">
                + NYTT SKJEMA
            </BButton>
        </div>

        <!-- Form Creator -->
        <div v-if="showCreator">
            <BCard class="max-w-4xl mx-auto !p-0 overflow-hidden">
                <div class="bg-black text-white p-4">
                    <h3 class="text-xl font-black uppercase italic">Konfigurer nytt skjema</h3>
                </div>
                
                <div class="p-8 space-y-8">
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <BInput 
                            v-model="newFormTitle" 
                            label="Tittel" 
                            placeholder="F.eks. Medlemsundersøkelse 2026"
                        />
                        <BSelect 
                            v-model="newFormOrg" 
                            label="Tilhører Organisasjon"
                        >
                            <option value="">Velg organisasjon...</option>
                            <option v-for="org in organizations" :key="org" :value="org">{{ org }}</option>
                        </BSelect>
                    </div>

                    <div class="space-y-4">
                        <div class="flex justify-between items-center border-b-4 border-black pb-2">
                            <h4 class="font-black uppercase text-lg italic">Felter & Spørsmål</h4>
                            <BButton @click="addField" variant="secondary" class="text-[10px] py-1 px-3">+ LEGG TIL</BButton>
                        </div>
                        
                        <div class="space-y-4">
                            <div v-for="(field, index) in newFormFields" :key="index" class="p-6 border-4 border-black bg-orange-50 relative group">
                                <button 
                                    @click="removeField(index)" 
                                    class="absolute -top-4 -right-4 bg-red-500 text-white w-8 h-8 border-4 border-black font-black flex items-center justify-center hover:bg-red-400 transition-colors shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]"
                                >
                                    X
                                </button>
                                
                                <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                                    <BInput v-model="field.name" label="ID (Teknisk navn)" placeholder="f_navn" class="!p-2 text-xs" />
                                    <BInput v-model="field.label" label="Vist Navn (Spørsmål)" placeholder="Hva heter du?" class="!p-2 text-xs" />
                                    <BSelect v-model="field.type" label="Input Type" class="!p-2 text-xs">
                                        <option value="text">Kort tekst</option>
                                        <option value="textarea">Lang tekst</option>
                                        <option value="select">Nedtrekk / Valg</option>
                                    </BSelect>
                                </div>
                                
                                <div v-if="field.type === 'select'" class="mt-4 pt-4 border-t-2 border-black border-dashed">
                                    <BInput 
                                        v-model="field.options" 
                                        label="Valgmuligheter (kommaseparert)" 
                                        placeholder="Ja, Nei, Kanskje" 
                                        class="!p-2 text-xs" 
                                    />
                                </div>
                            </div>

                            <div v-if="newFormFields.length === 0" class="py-12 text-center border-4 border-black border-dashed bg-gray-50">
                                <p class="font-black uppercase text-gray-400 italic">Ingen felter lagt til ennå</p>
                                <BButton @click="addField" variant="ghost" class="mt-4 text-xs">Klikk her for å starte</BButton>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="bg-gray-100 p-6 border-t-4 border-black flex gap-4">
                    <BButton @click="createForm" variant="success" class="flex-1 text-xl py-4" :disabled="!newFormTitle || !newFormOrg">
                        LAGRE OG PUBLISER
                    </BButton>
                    <BButton @click="showCreator = false" variant="secondary" class="px-10">
                        AVBRYT
                    </BButton>
                </div>
            </BCard>
        </div>

        <!-- Forms List -->
        <div v-if="!selectedForm && !showCreator" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            <BCard v-for="f in forms" :key="f.id" class="flex flex-col h-full hover:bg-yellow-50">
                <div class="flex-1">
                    <div class="flex justify-between items-start mb-4">
                        <span class="bg-black text-white text-[10px] font-black px-2 py-1 uppercase">{{ f.org_id.split('-')[0] }}</span>
                        <span class="text-[10px] font-bold text-gray-400 uppercase">v1.0</span>
                    </div>
                    <h3 class="font-black text-2xl mb-6 uppercase tracking-tighter leading-none">{{ f.title }}</h3>
                </div>
                <BButton @click="startForm(f)" variant="primary" class="w-full text-sm">
                    SVAR PÅ SKJEMA
                </BButton>
            </BCard>

            <div v-if="forms.length === 0 && !loading" class="col-span-full py-32 text-center border-8 border-black border-dashed bg-white">
                <p class="text-gray-300 text-6xl font-black uppercase mb-4 italic opacity-20 underline">Tomt ark</p>
                <p class="text-gray-400 font-bold uppercase tracking-widest">Ingen aktive undersøkelser i databasen.</p>
            </div>
        </div>

        <!-- Form Respondent View -->
        <div v-else-if="selectedForm" class="max-w-3xl mx-auto">
            <BCard class="!p-0 overflow-hidden">
                <div v-if="submitted" class="py-20 text-center space-y-6">
                    <h3 class="text-6xl font-black text-green-600 italic uppercase tracking-tighter">SUKSESS!</h3>
                    <div class="w-24 h-4 bg-green-600 mx-auto"></div>
                    <p class="font-black text-xl uppercase italic">Ditt svar er trygt lagret i Chronos-arkivet.</p>
                </div>
                
                <div v-else>
                    <div class="bg-yellow-400 p-8 border-b-4 border-black">
                        <h3 class="text-4xl font-black uppercase italic tracking-tighter leading-none">{{ selectedForm.title }}</h3>
                        <p class="mt-2 text-xs font-bold uppercase opacity-70">Vennligst fyll ut alle feltene under</p>
                    </div>

                    <div class="p-8 space-y-8">
                        <div v-for="field in (selectedForm.schema as any).fields" :key="field.name">
                            <BInput 
                                v-if="field.type === 'text'"
                                v-model="answers[field.name]"
                                :label="field.label"
                                class="w-full"
                            />
                            <div v-else-if="field.type === 'textarea'" class="flex flex-col">
                                <label class="block text-sm font-black uppercase tracking-tighter mb-1">{{ field.label }}</label>
                                <textarea 
                                    v-model="answers[field.name]"
                                    class="border-4 border-black p-3 font-mono h-40 focus:outline-none focus:bg-blue-50 focus:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] transition-all bg-white"
                                ></textarea>
                            </div>
                            <BSelect 
                                v-else-if="field.type === 'select'"
                                v-model="answers[field.name]"
                                :label="field.label"
                                class="w-full"
                            >
                                <option value="">Velg...</option>
                                <option v-for="opt in field.options" :key="opt" :value="opt">{{ opt }}</option>
                            </BSelect>
                        </div>
                    </div>

                    <div class="p-8 bg-gray-50 border-t-4 border-black flex gap-4">
                        <BButton @click="submit" variant="primary" class="flex-1 text-2xl py-6 italic">
                            SEND INN SVAR
                        </BButton>
                        <BButton @click="selectedForm = null" variant="secondary" class="px-8">
                            AVBRYT
                        </BButton>
                    </div>
                </div>
            </BCard>
        </div>
    </div>
</template>
