<script setup lang="ts">
import { ref } from 'vue'
import { api } from '../services/api'

const email = ref('')
const password = ref('')
const orgId = ref('')
const loading = ref(false)
const error = ref('')

const handleLogin = async () => {
  loading.value = true
  error.value = ''
  try {
    await api.login({
      email: email.value,
      password: password.value,
    })
  } catch (err: any) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

const handleSignUp = async () => {
  loading.value = true
  error.value = ''
  try {
    await api.signup({
      email: email.value,
      password: password.value,
      org_id: orgId.value
    })
    alert('Bruker opprettet. Du kan nå logge inn.')
  } catch (err: any) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8 p-8 bg-white shadow rounded-lg">
      <div>
        <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">Logg inn i Medlemsregister</h2>
      </div>
      <form class="mt-8 space-y-6" @submit.prevent="handleLogin">
        <div class="rounded-md shadow-sm -space-y-px">
          <div>
            <label for="org-id" class="sr-only">Organisasjons-ID</label>
            <input v-model="orgId" id="org-id" type="text" class="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm" placeholder="Organisasjons-ID (valgfritt)">
          </div>
          <div>
            <label for="email-address" class="sr-only">E-post</label>
            <input v-model="email" id="email-address" type="email" required class="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm" placeholder="E-post">
          </div>
          <div>
            <label for="password" class="sr-only">Passord</label>
            <input v-model="password" id="password" type="password" required class="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-b-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm" placeholder="Passord">
          </div>
        </div>

        <div v-if="error" class="text-red-500 text-sm">{{ error }}</div>

        <div class="flex space-x-4">
          <button type="submit" :disabled="loading" class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
            Logg inn
          </button>
          <button type="button" @click="handleSignUp" :disabled="loading" class="group relative w-full flex justify-center py-2 px-4 border border-indigo-600 text-sm font-medium rounded-md text-indigo-600 bg-white hover:bg-indigo-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
            Registrer
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
