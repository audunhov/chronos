<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../services/api'
import type { Form } from '../api'

const forms = ref<Form[]>([])
const loading = ref(false)
const selectedForm = ref<Form | null>(null)
const answers = ref<Record<string, any>>({})
const submitted = ref(false)

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

const startForm = (form: Form) => {
    selectedForm.value = form
    answers.value = {}
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

onMounted(fetchForms)
</script>

<template>
    <div class="space-y-8">
        <h2 class="text-2xl font-black uppercase">Undersøkelser & Skjema</h2>

        <div v-if="!selectedForm" class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div v-for="f in forms" :key="f.id" class="brutalist-card bg-white">
                <h3 class="font-black text-xl mb-4">{{ f.title }}</h3>
                <button @click="startForm(f)" class="brutalist-btn-primary text-xs">SVAR NÅ</button>
            </div>
            <div v-if="forms.length === 0 && !loading" class="col-span-full py-20 text-center brutalist-card border-dashed">
                <p class="text-gray-400 font-bold uppercase">Ingen aktive undersøkelser.</p>
            </div>
        </div>

        <div v-else class="brutalist-card bg-white max-w-2xl mx-auto">
            <div v-if="submitted" class="py-10 text-center">
                <h3 class="text-4xl font-black text-green-600 mb-4">TAKKEKORT!</h3>
                <p class="font-bold uppercase">Ditt svar er registrert i Chronos.</p>
            </div>
            <div v-else>
                <h3 class="text-3xl font-black uppercase mb-8 border-b-4 border-black pb-4">{{ selectedForm.title }}</h3>
                <div class="space-y-6">
                    <div v-for="field in (selectedForm.schema as any).fields" :key="field.name" class="flex flex-col">
                        <label class="brutalist-label">{{ field.label }}</label>
                        <input 
                            v-if="field.type === 'text'"
                            v-model="answers[field.name]"
                            type="text"
                            class="brutalist-input"
                        />
                        <select 
                            v-if="field.type === 'select'"
                            v-model="answers[field.name]"
                            class="brutalist-input"
                        >
                            <option v-for="opt in field.options" :key="opt" :value="opt">{{ opt }}</option>
                        </select>
                        <textarea 
                            v-if="field.type === 'textarea'"
                            v-model="answers[field.name]"
                            class="brutalist-input h-32"
                        ></textarea>
                    </div>
                </div>
                <div class="mt-12 flex gap-4">
                    <button @click="submit" class="brutalist-btn-primary flex-1">SEND INN</button>
                    <button @click="selectedForm = null" class="brutalist-btn bg-white">AVBRYT</button>
                </div>
            </div>
        </div>
    </div>
</template>
