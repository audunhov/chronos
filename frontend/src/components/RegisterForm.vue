<script setup lang="ts">
import { ref } from 'vue'
import { api } from '../services/api'

const emit = defineEmits(['registered', 'cancel'])

const form = ref({
  name: '',
  email: '',
  org_id: '',
  birth_year: 2000,
  metadata: {}
})

const loading = ref(false)
const error = ref('')

const handleSubmit = async () => {
  loading.value = true
  error.value = ''
  try {
    const payload = {
      ...form.value,
      metadata: { 
        ...form.value.metadata, 
        birth_year: form.value.birth_year 
      }
    }
    await api.registerMember(payload)
    emit('registered')
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <h3 class="text-lg font-bold leading-6 text-gray-900 mb-4" id="modal-title">Registrer nytt medlem</h3>
    <form @submit.prevent="handleSubmit" class="space-y-4">
      <div>
        <label for="name" class="block text-sm font-medium text-gray-700">Fullt navn</label>
        <input type="text" id="name" v-model="form.name" required class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm p-2 border" />
      </div>
      <div>
        <label for="email" class="block text-sm font-medium text-gray-700">E-postadresse</label>
        <input type="email" id="email" v-model="form.email" required class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm p-2 border" />
      </div>
      <div>
        <label for="birth_year" class="block text-sm font-medium text-gray-700">Fødselsår (for kontingent)</label>
        <input type="number" id="birth_year" v-model="form.birth_year" required class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm p-2 border" />
      </div>
      <div>
        <label for="org_id" class="block text-sm font-medium text-gray-700">Organisasjons-ID (valgfri)</label>
        <input type="text" id="org_id" v-model="form.org_id" placeholder="La stå tom for din egen org" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm p-2 border" />
      </div>
      
      <div v-if="error" class="text-sm text-red-600 font-medium">
        ⚠️ {{ error }}
      </div>

      <div class="mt-6 flex flex-row-reverse gap-3">
        <button type="submit" :disabled="loading" class="inline-flex w-full justify-center rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 sm:w-auto">
          {{ loading ? 'Lagrer...' : 'Registrer medlem' }}
        </button>
        <button type="button" @click="emit('cancel')" class="inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 sm:w-auto">
          Avbryt
        </button>
      </div>
    </form>
  </div>
</template>
