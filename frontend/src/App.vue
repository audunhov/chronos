<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api } from './services/api'
import { supabase } from './services/supabase'
import RegisterForm from './components/RegisterForm.vue'
import LoginForm from './components/LoginForm.vue'

interface Member {
  ID: string
  OrgID: string
  Name: string
  Email: string
  Status: string
  Metadata: any
  Balance: number
}

const session = ref<any>(null)
const members = ref<Member[]>([])
const loading = ref(false)
const error = ref('')
const selectedDate = ref('')
const showModal = ref(false)

const fetchMembers = async () => {
  if (!session.value) return
  loading.value = true
  error.value = ''
  try {
    if (selectedDate.value) {
      members.value = await api.getMembersAsOf(selectedDate.value)
    } else {
      members.value = await api.getMembers()
    }
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

const handleLogout = async () => {
  await supabase.auth.signOut()
}

const onMemberRegistered = () => {
  showModal.value = false
  fetchMembers()
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

watch(selectedDate, () => {
  fetchMembers()
})

onMounted(() => {
  supabase.auth.getSession().then(({ data }) => {
    session.value = data.session
  })

  supabase.auth.onAuthStateChange((_event, _session) => {
    session.value = _session
    if (_session) fetchMembers()
    else members.value = []
  })
})
</script>

<template>
  <div v-if="!session">
    <LoginForm />
  </div>
  <div v-else class="min-h-screen bg-gray-50 py-8 px-4 sm:px-6 lg:px-8">
    <div class="max-w-6xl mx-auto">
      <div class="flex flex-col md:flex-row md:items-center md:justify-between mb-8">
        <div>
          <h1 class="text-3xl font-bold text-gray-900">Medlemsregister</h1>
          <p class="mt-2 text-sm text-gray-700">Innlogget som: {{ session.user.email }}</p>
        </div>
        <div class="mt-4 md:mt-0 flex flex-wrap gap-3 items-center">
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

      <div class="bg-white shadow overflow-hidden sm:rounded-lg border border-gray-200">
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
              <td colspan="5" class="px-6 py-4 text-center text-sm text-gray-500">Laster data fra uforanderlig logg...</td>
            </tr>
            <tr v-else v-for="member in members" :key="member.ID">
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="text-sm font-medium text-gray-900">{{ member.Name }}</div>
                <div class="text-xs text-gray-400 font-mono">{{ member.ID }}</div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ member.Email }}</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span :class="[
                  'px-2 py-1 text-xs font-semibold rounded-full',
                  member.Status === 'ACTIVE' ? 'bg-green-100 text-green-800' : 
                  member.Status === 'SHREDDED' ? 'bg-gray-100 text-gray-500' : 'bg-red-100 text-red-800'
                ]">
                  {{ member.Status }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 font-semibold">
                {{ (member.Balance / 100).toFixed(2) }} kr
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <button 
                  v-if="member.Status !== 'SHREDDED'"
                  @click="shredMember(member.ID)"
                  class="text-red-600 hover:text-red-900"
                >
                  Glem (GDPR)
                </button>
              </td>
            </tr>
            <tr v-if="!loading && members.length === 0">
              <td colspan="5" class="px-6 py-4 text-center text-sm text-gray-500">Ingen medlemmer i dette hierarkiet.</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Register Modal -->
    <div v-if="showModal" class="fixed inset-0 z-10 overflow-y-auto">
      <div class="flex min-h-screen items-end justify-center px-4 pt-4 pb-20 text-center sm:block sm:p-0">
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75" @click="showModal = false"></div>
        <span class="hidden sm:inline-block sm:h-screen sm:align-middle">&#8203;</span>
        <div class="inline-block transform overflow-hidden rounded-lg bg-white px-4 pt-5 pb-4 text-left align-bottom shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6 sm:align-middle">
          <RegisterForm @registered="onMemberRegistered" @cancel="showModal = false" />
        </div>
      </div>
    </div>
  </div>
</template>
