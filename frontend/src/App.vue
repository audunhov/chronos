<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api } from './services/api'
import { auth, logout } from './services/auth'
import type { Member } from './api'
import RegisterForm from './components/RegisterForm.vue'
import LoginForm from './components/LoginForm.vue'

const members = ref<Member[]>([])
const organizations = ref<string[]>([])
const selectedOrg = ref('')
const loading = ref(false)
const error = ref('')
const selectedDate = ref('')
const showModal = ref(false)

const fetchMembers = async (silent = false) => {
  if (!auth.user) {
    members.value = []
    return
  }
  if (!silent) loading.value = true
  error.value = ''
  try {
    let data: any
    if (selectedDate.value) {
      data = await api.getMembersAsOf(selectedDate.value, selectedOrg.value)
    } else {
      data = await api.getMembers(selectedOrg.value)
    }
    members.value = Array.isArray(data) ? data : []
  } catch (e: any) {
    if (!silent) {
      console.error('Fetch members failed:', e)
      error.value = e.message
    }
  } finally {
    if (!silent) loading.value = false
  }
}

const fetchOrganizations = async () => {
  if (!auth.user) return
  try {
    const data = await api.getOrganizations()
    organizations.value = Array.isArray(data) ? data : []
  } catch (e) {
    console.error('Failed to fetch orgs:', e)
  }
}

const handleLogout = () => {
  logout()
}

const onMemberRegistered = () => {
  showModal.value = false
  fetchMembers()
  fetchOrganizations()
}

const shredMember = async (id: string) => {
  if (!confirm('Er du sikker på at du vil slette personopplysningene til dette medlemmet? Dette kan ikke angres (Crypto-shredding).')) return
  try {
    await api.shredMember(id)
    fetchMembers()
  } catch (e: any) {
    alert(e.message)
  }
}

// Reager på endringer i filtere
watch([selectedDate, selectedOrg], () => {
  fetchMembers()
})

// Reager på innlogging/utlogging
watch(() => auth.user, (newUser) => {
  if (newUser) {
    fetchMembers()
    fetchOrganizations()
  } else {
    members.value = []
    organizations.value = []
  }
}, { immediate: true })

onMounted(() => {
  if (auth.user) {
    fetchMembers()
    fetchOrganizations()
  }
})
</script>

<template>
  <div v-if="!auth.user">
    <LoginForm />
  </div>
  <div v-else class="min-h-screen bg-gray-50 py-8 px-4 sm:px-6 lg:px-8">
    <div class="max-w-6xl mx-auto">
      <div class="flex flex-col md:flex-row md:items-center md:justify-between mb-8">
        <div>
          <h1 class="text-3xl font-bold text-gray-900">Medlemsregister</h1>
          <p class="mt-2 text-sm text-gray-700" v-if="auth.user">Innlogget som: {{ auth.user.email }}</p>
        </div>
        <div class="mt-4 md:mt-0 flex flex-wrap gap-3 items-center">
          <button
            @click="fetchMembers()"
            class="p-2 text-gray-400 hover:text-indigo-600 transition-colors"
            title="Oppdater liste"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z" clip-rule="evenodd" />
            </svg>
          </button>

          <div class="flex items-center space-x-2 bg-white p-2 rounded-md shadow-sm border border-gray-200">
            <label for="org-select" class="text-xs font-semibold text-gray-500 uppercase">Organisasjon</label>
            <select 
              id="org-select"
              v-model="selectedOrg"
              class="block rounded border-gray-300 text-sm focus:ring-indigo-500"
            >
              <option value="">Alle (Global)</option>
              <option v-for="org in organizations" :key="org" :value="org">{{ org }}</option>
            </select>
          </div>

          <div class="flex items-center space-x-2 bg-white p-2 rounded-md shadow-sm border border-gray-200">
            <label for="date" class="text-xs font-semibold text-gray-500 uppercase">Tidsmaskin</label>
            <input 
              id="date"
              type="date" 
              v-model="selectedDate"
              class="block rounded border-gray-300 text-sm focus:ring-indigo-500"
            />
            <button 
              v-if="selectedDate" 
              @click="selectedDate = ''"
              class="text-xs text-red-600 font-bold"
            >
              ✕
            </button>
          </div>
          <button
            @click="showModal = true"
            class="inline-flex items-center justify-center rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-700"
          >
            Nytt medlem
          </button>
          <button
            @click="handleLogout"
            class="inline-flex items-center justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50"
          >
            Logg ut
          </button>
        </div>
      </div>

      <div v-if="error" class="bg-red-50 border-l-4 border-red-400 p-4 mb-6">
        <p class="text-sm text-red-700">{{ error }}</p>
      </div>

      <div v-if="members" class="bg-white shadow overflow-hidden sm:rounded-lg border border-gray-200">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Navn</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">E-post</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Saldo</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Handlinger</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-if="loading" class="animate-pulse">
              <td colspan="5" class="px-6 py-4 text-center text-sm text-gray-500">Laster data...</td>
            </tr>
            <template v-else-if="members && members.length > 0">
              <tr v-for="member in members" :key="member.id">
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm font-medium text-gray-900">{{ member.name }}</div>
                  <div class="text-xs text-gray-400 font-mono">{{ member.id }}</div>
                  <div class="text-xs text-indigo-500 font-semibold">{{ member.org_id }}</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ member.email }}</td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span :class="[
                    'px-2 py-1 text-xs font-semibold rounded-full',
                    member.status === 'ACTIVE' ? 'bg-green-100 text-green-800' : 
                    member.status === 'SHREDDED' ? 'bg-gray-100 text-gray-500' : 'bg-red-100 text-red-800'
                  ]">
                    {{ member.status }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 font-semibold">
                  {{ ((member.balance || 0) / 100).toFixed(2) }} kr
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  <button 
                    v-if="member.status !== 'SHREDDED'"
                    @click="shredMember(member.id)"
                    class="text-red-600 hover:text-red-900"
                  >
                    Glem (GDPR)
                  </button>
                </td>
              </tr>
            </template>
            <tr v-else-if="!loading">
              <td colspan="5" class="px-6 py-4 text-center text-sm text-gray-500">Ingen medlemmer funnet.</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Register Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-gray-500 bg-opacity-75 transition-opacity" @click="showModal = false"></div>
      <div class="relative bg-white rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:max-w-lg sm:w-full sm:p-6">
        <RegisterForm @registered="onMemberRegistered" @cancel="showModal = false" />
      </div>
    </div>
  </div>
</template>
