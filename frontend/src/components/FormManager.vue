<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../services/api'
import { auth } from '../services/auth'
import type { Form, EventReaction, FormResponse } from '../api'
import BCard from './base/BCard.vue'
import BButton from './base/BButton.vue'
import BInput from './base/BInput.vue'
import BSelect from './base/BSelect.vue'
import BBadge from './base/BBadge.vue'

const props = defineProps<{
    isAdmin?: boolean
}>()

const forms = ref<Form[]>([])
const loading = ref(false)
const selectedForm = ref<Form | null>(null)
const answers = ref<Record<string, any>>({})
const submitted = ref(false)

// Form Editor state
const showCreator = ref(false)
const editingFormId = ref<string | null>(null)
const newFormTitle = ref('')
const newFormOrg = ref('')
const newFormFields = ref<{ name: string, label: string, type: 'text' | 'textarea' | 'select', options: string }[]>([])
const organizations = ref<{ id: string, name: string }[]>([])
const dependentReactions = ref<EventReaction[]>([])

// Response Viewer state
const viewingResponsesFor = ref<Form | null>(null)
const responses = ref<FormResponse[]>([])

const triggerExport = (format: 'csv' | 'xlsx') => {
    if (!viewingResponsesFor.value) return
    api.exportData('form-responses', { 
        form_id: viewingResponsesFor.value.id!, 
        format 
    }, `${viewingResponsesFor.value.title}.${format}`)
}

const addField = () => {
    newFormFields.value.push({ name: '', label: '', type: 'text', options: '' })
}

const removeField = (index: number) => {
    newFormFields.value.splice(index, 1)
}

const saveForm = async () => {
    if (!newFormTitle.value || !newFormOrg.value) return
    
    const schema = {
        fields: newFormFields.value.map(f => ({
            ...f,
            options: typeof f.options === 'string' ? f.options.split(',').map(s => s.trim()) : f.options
        }))
    }

    try {
        if (editingFormId.value) {
            await api.updateForm({
                id: editingFormId.value,
                title: newFormTitle.value,
                schema
            })
        } else {
            await api.createForm({
                title: newFormTitle.value,
                org_id: newFormOrg.value,
                schema
            })
        }
        showCreator.value = false
        editingFormId.value = null
        newFormTitle.value = ''
        newFormOrg.value = ''
        newFormFields.value = []
        dependentReactions.value = []
        fetchForms()
    } catch (e: any) {
        alert(e.message)
    }
}

const deleteForm = async () => {
    if (!editingFormId.value) return
    if (!confirm('Er du sikker på at du vil slette dette skjemaet? Alle innsendte svar vil også bli fjernet permanent.')) return
    
    try {
        await api.deleteForm(editingFormId.value)
        showCreator.value = false
        editingFormId.value = null
        fetchForms()
    } catch (e: any) {
        alert(e.message)
    }
}

const startEdit = async (form: Form) => {
    editingFormId.value = form.id!
    newFormTitle.value = form.title!
    newFormOrg.value = form.org_id!
    newFormFields.value = (form.schema as any).fields.map((f: any) => ({
        ...f,
        options: Array.isArray(f.options) ? f.options.join(', ') : ''
    }))
    showCreator.value = true

    // Sjekk om det finnes avhengige pipelines
    try {
        const reactions = await api.getReactions(form.org_id!, form.id!)
        dependentReactions.value = Array.isArray(reactions) ? reactions.filter(r => r.trigger_event === 'FormResponseSubmitted') : []
    } catch (e) {
        console.error("Failed to check dependencies", e)
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
        organizations.value = Array.isArray(data) ? (data as any).map((d: any) => {
            if (typeof d === 'string') return { id: d, name: d }
            return d
        }) : []
    } catch (e) {
        console.error(e)
    }
}

const fetchResponses = async (form: Form) => {
    loading.value = true
    try {
        const data = await api.getFormResponses(form.id!)
        responses.value = Array.isArray(data) ? data : []
        viewingResponsesFor.value = form
    } catch (e) {
        console.error(e)
    } finally {
        loading.value = false
    }
}

const startForm = (form: Form) => {
    selectedForm.value = form
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
            <BButton v-if="props.isAdmin && !showCreator && !viewingResponsesFor" @click="showCreator = true" variant="primary" class="text-xs py-2">
                + NYTT SKJEMA
            </BButton>
        </div>

        <!-- Form Creator / Editor -->
        <div v-if="showCreator">
            <BCard class="max-w-4xl mx-auto !p-0 overflow-hidden">
                <div class="bg-black text-white p-4">
                    <h3 class="text-xl font-black uppercase italic">{{ editingFormId ? 'Rediger' : 'Konfigurer nytt' }} skjema</h3>
                </div>
                
                <div v-if="dependentReactions.length > 0" class="bg-red-500 text-white p-6 border-b-4 border-black">
                    <div class="flex items-center gap-4">
                        <span class="text-4xl">⚠️</span>
                        <div>
                            <h4 class="font-black uppercase tracking-tight">ADVARSEL: AKTIVE PIPELINES</h4>
                            <p class="text-sm font-bold opacity-90">Dette skjemaet er koblet til {{ dependentReactions.length }} aktive pipelines. Endringer i felt-IDer kan ødelegge automatikken!</p>
                        </div>
                    </div>
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
                            :disabled="!!editingFormId"
                        >
                            <option value="">Velg organisasjon...</option>
                            <option v-for="org in organizations" :key="org.id" :value="org.id">{{ org.name }}</option>
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
                    <BButton @click="saveForm" variant="success" class="flex-1 text-xl py-4" :disabled="!newFormTitle || !newFormOrg">
                        {{ editingFormId ? 'OPPDATER' : 'LAGRE OG PUBLISER' }}
                    </BButton>
                    <BButton v-if="editingFormId" @click="deleteForm" variant="danger" class="px-10 italic">
                        SLETT SKJEMA
                    </BButton>
                    <BButton @click="showCreator = false; editingFormId = null" variant="secondary" class="px-10">
                        AVBRYT
                    </BButton>
                </div>
            </BCard>
        </div>

        <!-- Forms List -->
        <div v-if="!selectedForm && !showCreator && !viewingResponsesFor" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            <BCard v-for="f in forms" :key="f.id" class="flex flex-col h-full hover:bg-yellow-50 group">
                <div class="flex-1">
                    <div class="flex justify-between items-start mb-4">
                        <span class="bg-black text-white text-[10px] font-black px-2 py-1 uppercase">{{ f.org_id ? f.org_id.split('-')[0] : '' }}</span>
                        <div class="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                            <BButton v-if="props.isAdmin" @click="fetchResponses(f)" variant="ghost" class="text-[8px] py-1 border-2 font-black italic">SVAR</BButton>
                            <BButton v-if="props.isAdmin" @click="startEdit(f)" variant="ghost" class="text-[8px] py-1 border-2 font-black italic">REDIGER</BButton>
                        </div>
                    </div>
                    <h3 class="font-black text-2xl mb-6 uppercase tracking-tighter leading-none">{{ f.title }}</h3>
                </div>
                <BButton @click="startForm(f)" variant="primary" class="w-full text-sm italic font-black">
                    SVAR PÅ SKJEMA
                </BButton>
            </BCard>

            <div v-if="forms.length === 0 && !loading" class="col-span-full py-32 text-center border-8 border-black border-dashed bg-white">
                <p class="text-gray-300 text-6xl font-black uppercase mb-4 italic opacity-20 underline">Tomt ark</p>
                <p class="text-gray-400 font-bold uppercase tracking-widest">Ingen aktive undersøkelser i databasen.</p>
            </div>
        </div>

        <!-- Form Responses View -->
        <div v-if="viewingResponsesFor" class="space-y-8 animate-in fade-in">
            <div class="flex flex-col md:flex-row justify-between items-start md:items-center border-b-4 border-black pb-4 gap-4">
                <div>
                    <h3 class="text-3xl font-black uppercase italic tracking-tighter leading-none">{{ viewingResponsesFor.title }}</h3>
                    <p class="text-xs font-bold uppercase text-gray-400 mt-2">Innsendte svar fra medlemmer</p>
                </div>
                <div class="flex flex-wrap gap-2">
                    <BButton @click="triggerExport('csv')" variant="secondary" class="text-[10px] py-1">EKSPORTER CSV</BButton>
                    <BButton @click="triggerExport('xlsx')" variant="primary" class="text-[10px] py-1 shadow-[4px_4px_0px_0px_white]">EKSPORTER XLSX</BButton>
                    <BButton @click="viewingResponsesFor = null" variant="secondary" class="text-[10px] py-1 ml-4 border-2 border-black">← TILBAKE</BButton>
                </div>
            </div>

            <div class="overflow-x-auto border-4 border-black shadow-[12px_12px_0px_0px_rgba(0,0,0,1)]">
                <table class="brutalist-table bg-white">
                    <thead>
                        <tr>
                            <th class="brutalist-th">Medlem</th>
                            <th v-for="field in (viewingResponsesFor.schema as any).fields" :key="field.name" class="brutalist-th">
                                {{ field.label }}
                            </th>
                            <th class="brutalist-th">Innsendt</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="resp in responses" :key="resp.id" class="hover:bg-yellow-50">
                            <td class="brutalist-td">
                                <div class="font-black uppercase text-xs">{{ resp.user_name }}</div>
                                <div class="text-[8px] font-bold text-gray-400 font-mono">{{ resp.user_email }}</div>
                            </td>
                            <td v-for="field in (viewingResponsesFor.schema as any).fields" :key="field.name" class="brutalist-td italic font-bold text-xs">
                                {{ resp.answers?.[field.name] || '-' }}
                            </td>
                            <td class="brutalist-td text-[10px] font-mono whitespace-nowrap">
                                {{ new Date(resp.created_at || '').toLocaleDateString() }}
                            </td>
                        </tr>
                        <tr v-if="responses.length === 0">
                            <td :colspan="(viewingResponsesFor.schema as any).fields.length + 2" class="brutalist-td text-center py-20 font-black uppercase text-gray-300 italic text-2xl opacity-20">
                                Ingen svar mottatt ennå
                            </td>
                        </tr>
                    </tbody>
                </table>
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
                        <BButton @click="submit" variant="primary" class="flex-1 text-2xl py-6 italic font-black">
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
