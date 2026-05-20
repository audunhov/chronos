<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api } from './services/api'
import { auth, logout } from './services/auth'
import type { Member, MyMembership, TreasuryItem } from './api'
import RegisterForm from './components/RegisterForm.vue'
import LoginForm from './components/LoginForm.vue'
import OrgManager from './components/OrgManager.vue'
import StatsDashboard from './components/StatsDashboard.vue'
import FormManager from './components/FormManager.vue'

const members = ref<Member[]>([])
const myMemberships = ref<MyMembership[]>([])
const organizations = ref<string[]>([])
const selectedOrg = ref('')
const loading = ref(false)
const error = ref('')
const selectedDate = ref('')
const showModal = ref(false)
const currentView = ref<'admin' | 'me'>('me')
const adminTab = ref<'members' | 'treasury' | 'orgs' | 'stats'>('members')
const meTab = ref<'memberships' | 'forms'>('memberships')
const treasuryReport = ref<TreasuryItem[]>([])

const fetchMembers = async (silent = false) => {
  if (!auth.user || currentView.value !== 'admin') {
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

const fetchMyMemberships = async () => {
  if (!auth.user) return
  loading.value = true
  try {
    const data = await api.getMyMemberships()
    myMemberships.value = Array.isArray(data) ? data : []
  } catch (e) {
    console.error('Failed to fetch my memberships:', e)
  } finally {
    loading.value = false
  }
}

const fetchTreasuryReport = async () => {
    loading.value = true
    try {
        const data = await api.getTreasuryReport()
        treasuryReport.value = Array.isArray(data) ? data : []
    } catch (e) {
        console.error('Failed to fetch treasury report:', e)
    } finally {
        loading.value = false
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
  if (currentView.value === 'admin') fetchMembers()
  if (currentView.value === 'me') fetchMyMemberships()
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
  if (currentView.value === 'admin' && adminTab.value === 'members') fetchMembers()
})

watch(currentView, (view) => {
    if (view === 'admin') {
        if (adminTab.value === 'members') fetchMembers()
        if (adminTab.value === 'treasury') fetchTreasuryReport()
    }
    if (view === 'me') fetchMyMemberships()
})

watch(adminTab, (tab) => {
    if (tab === 'members') fetchMembers()
    if (tab === 'treasury') fetchTreasuryReport()
})

// Reager på innlogging/utlogging
watch(() => auth.user, (newUser) => {
  if (newUser) {
    if (currentView.value === 'admin') {
        if (adminTab.value === 'members') fetchMembers()
        if (adminTab.value === 'treasury') fetchTreasuryReport()
    }
    if (currentView.value === 'me') fetchMyMemberships()
    fetchOrganizations()
  } else {
    members.value = []
    myMemberships.value = []
    organizations.value = []
    treasuryReport.value = []
  }
}, { immediate: true })

onMounted(() => {
  if (auth.user) {
    if (currentView.value === 'admin') {
        if (adminTab.value === 'members') fetchMembers()
        if (adminTab.value === 'treasury') fetchTreasuryReport()
    }
    if (currentView.value === 'me') fetchMyMemberships()
    fetchOrganizations()
  }
})
</script>

<template>
  <div v-if="!auth.user" class="min-h-screen flex items-center justify-center p-4">
    <div class="max-w-md w-full">
      <LoginForm />
    </div>
  </div>
  <div v-else class="min-h-screen py-8 px-4 sm:px-6 lg:px-8">
    <div class="max-w-7xl mx-auto">
      <!-- Header -->
      <div class="flex flex-col lg:flex-row lg:items-center lg:justify-between mb-12 gap-6 bg-white border-4 border-black p-8 shadow-[8px_8px_0px_0px_rgba(0,0,0,1)]">
        <div>
          <h1 class="text-5xl font-black uppercase tracking-tighter italic">CHRONOS</h1>
          <p class="mt-2 text-sm font-bold bg-blue-400 inline-block px-2 border-2 border-black" v-if="auth.user">IDENT: {{ auth.user.email }}</p>
        </div>
        <div class="flex flex-wrap gap-4 items-center">
          <!-- View Switcher -->
          <nav class="flex border-4 border-black bg-black p-1">
            <button 
                @click="currentView = 'me'"
                :class="[
                    'px-4 py-2 text-xs font-black uppercase transition-colors',
                    currentView === 'me' ? 'bg-yellow-400 text-black' : 'bg-black text-white hover:bg-gray-800'
                ]"
            >
                MIN PROFIL
            </button>
            <button 
                @click="currentView = 'admin'"
                :class="[
                    'px-4 py-2 text-xs font-black uppercase transition-colors',
                    currentView === 'admin' ? 'bg-green-400 text-black' : 'bg-black text-white hover:bg-gray-800'
                ]"
            >
                ADMIN
            </button>
          </nav>

          <button
            @click="handleLogout"
            class="brutalist-btn bg-white hover:bg-red-500 hover:text-white px-4 py-2 text-xs uppercase"
          >
            LOGG UT
          </button>
        </div>
      </div>

      <div v-if="error" class="bg-red-500 text-white border-4 border-black p-4 mb-8 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] font-bold">
        ERROR :: {{ error }}
      </div>

      <!-- PERSONAL VIEW -->
      <div v-if="currentView === 'me'">
        <div class="flex flex-col xl:flex-row xl:items-end justify-between mb-8 gap-6 border-b-4 border-black pb-6">
            <nav class="flex gap-4">
                <button 
                    @click="meTab = 'memberships'"
                    :class="[
                        'brutalist-btn text-sm uppercase',
                        meTab === 'memberships' ? 'bg-yellow-400 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]' : 'bg-white hover:bg-gray-100'
                    ]"
                >
                    MINE MEDLEMSKAP
                </button>
                <button 
                    @click="meTab = 'forms'"
                    :class="[
                        'brutalist-btn text-sm uppercase',
                        meTab === 'forms' ? 'bg-purple-400 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]' : 'bg-white hover:bg-gray-100'
                    ]"
                >
                    UNDERSØKELSER
                </button>
            </nav>
            <button
                v-if="meTab === 'memberships'"
                @click="showModal = true"
                class="brutalist-btn-primary"
            >
                + BLI MEDLEM
            </button>
        </div>

        <div v-if="meTab === 'memberships'">
            <div v-if="loading" class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
                <div v-for="i in 3" :key="i" class="brutalist-card animate-pulse h-48 bg-gray-100"></div>
            </div>
            <div v-else-if="myMemberships.length > 0" class="grid grid-cols-1 gap-8 sm:grid-cols-2 lg:grid-cols-3">
                <div v-for="ms in myMemberships" :key="ms.id" class="brutalist-card bg-white group">
                    <div class="flex justify-between items-start mb-6">
                        <div>
                            <h3 class="font-black text-2xl uppercase leading-none mb-1 group-hover:text-blue-600 transition-colors">{{ ms.org_name }}</h3>
                            <span class="brutalist-badge bg-gray-200">{{ ms.role }}</span>
                        </div>
                        <span :class="[
                            'brutalist-badge',
                            ms.status === 'ACTIVE' ? 'bg-green-400' : 'bg-red-400'
                        ]">
                            {{ ms.status }}
                        </span>
                    </div>
                    <div class="mt-8 pt-6 border-t-4 border-black flex justify-between items-end">
                        <div>
                            <p class="text-[10px] font-black uppercase text-gray-500 mb-1">Ditt utestående:</p>
                            <p :class="['text-3xl font-black italic', (ms.balance || 0) < 0 ? 'text-red-600' : 'text-black']">
                                {{ ((ms.balance || 0) / 100).toFixed(2) }} <small class="text-sm not-italic">NOK</small>
                            </p>
                        </div>
                        <button class="brutalist-btn bg-black text-white text-[10px] px-3 py-1 hover:bg-yellow-400 hover:text-black">BETAL</button>
                    </div>
                </div>
            </div>
            <div v-else class="brutalist-card bg-white border-dashed border-gray-400 py-20 text-center">
                <p class="text-xl font-bold uppercase text-gray-400">Ingen aktive medlemskap funnet.</p>
            </div>
        </div>

        <div v-if="meTab === 'forms'">
            <FormManager />
        </div>
      </div>

      <!-- ADMIN VIEW -->
      <div v-if="currentView === 'admin'">
        <div class="flex flex-col xl:flex-row xl:items-end justify-between mb-8 gap-6 border-b-4 border-black pb-6">
            <nav class="flex flex-wrap gap-4">
                <button 
                    @click="adminTab = 'members'"
                    :class="[
                        'brutalist-btn text-sm uppercase',
                        adminTab === 'members' ? 'bg-blue-400 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]' : 'bg-white hover:bg-gray-100'
                    ]"
                >
                    MEDLEMSLISTE
                </button>
                <button 
                    @click="adminTab = 'treasury'"
                    :class="[
                        'brutalist-btn text-sm uppercase',
                        adminTab === 'treasury' ? 'bg-indigo-400 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]' : 'bg-white hover:bg-gray-100'
                    ]"
                >
                    FINANS
                </button>
                <button 
                    @click="adminTab = 'orgs'"
                    :class="[
                        'brutalist-btn text-sm uppercase',
                        adminTab === 'orgs' ? 'bg-orange-400 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]' : 'bg-white hover:bg-gray-100'
                    ]"
                >
                    STRUKTUR
                </button>
                <button 
                    @click="adminTab = 'stats'"
                    :class="[
                        'brutalist-btn text-sm uppercase',
                        adminTab === 'stats' ? 'bg-yellow-400 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]' : 'bg-white hover:bg-gray-100'
                    ]"
                >
                    STATISTIKK
                </button>
            </nav>

            <div v-if="adminTab === 'members'" class="flex flex-wrap gap-4 items-end">
                <div class="flex flex-col">
                    <label class="brutalist-label">Org-Filter</label>
                    <select 
                        v-model="selectedOrg"
                        class="brutalist-input text-xs"
                    >
                        <option value="">ALLE (GLOBAL)</option>
                        <option v-for="org in organizations" :key="org" :value="org">{{ org }}</option>
                    </select>
                </div>

                <div class="flex flex-col">
                    <label class="brutalist-label">Tidsmaskin</label>
                    <div class="flex gap-2">
                        <input 
                            type="date" 
                            v-model="selectedDate"
                            class="brutalist-input text-xs py-2"
                        />
                        <button 
                            v-if="selectedDate" 
                            @click="selectedDate = ''"
                            class="brutalist-btn bg-red-500 text-white px-3 py-2"
                        >✕</button>
                    </div>
                </div>
                
                <button @click="fetchMembers()" class="brutalist-btn bg-yellow-400 p-3" title="Refresh">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                        <path fill-rule="evenodd" d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z" clip-rule="evenodd" />
                    </svg>
                </button>
            </div>
        </div>

        <!-- Members Tab Content -->
        <div v-if="adminTab === 'members'" class="overflow-x-auto shadow-[12px_12px_0px_0px_rgba(0,0,0,1)] mb-12">
            <table class="brutalist-table bg-white">
            <thead>
                <tr>
                <th class="brutalist-th">Identitet</th>
                <th class="brutalist-th">Kontakt</th>
                <th class="brutalist-th">Status</th>
                <th class="brutalist-th">Saldo</th>
                <th class="brutalist-th text-right">Faresone</th>
                </tr>
            </thead>
            <tbody>
                <tr v-if="loading" class="animate-pulse">
                <td colspan="5" class="brutalist-td text-center py-10 uppercase font-black italic">Henter rader...</td>
                </tr>
                <template v-else-if="members && members.length > 0">
                <tr v-for="member in members" :key="member.id" class="hover:bg-yellow-50 transition-colors">
                    <td class="brutalist-td">
                        <div class="font-black text-lg uppercase">{{ member.name }}</div>
                        <div class="text-[10px] text-gray-400 font-mono tracking-tighter">{{ member.id }}</div>
                        <div class="mt-2"><span class="bg-indigo-100 text-indigo-800 text-[10px] px-2 font-black border border-black uppercase">{{ member.org_id }}</span></div>
                    </td>
                    <td class="brutalist-td font-bold italic">{{ member.email }}</td>
                    <td class="brutalist-td">
                    <span :class="[
                        'brutalist-badge',
                        member.status === 'ACTIVE' ? 'bg-green-400' : 
                        member.status === 'SHREDDED' ? 'bg-gray-300' : 'bg-red-400'
                    ]">
                        {{ member.status }}
                    </span>
                    </td>
                    <td class="brutalist-td font-black">
                        {{ ((member.balance || 0) / 100).toFixed(2) }} kr
                    </td>
                    <td class="brutalist-td text-right">
                    <button 
                        v-if="member.status !== 'SHREDDED'"
                        @click="shredMember(member.id!)"
                        class="brutalist-btn bg-white hover:bg-black hover:text-white px-3 py-1 text-[10px]"
                    >
                        SHRED (GDPR)
                    </button>
                    </td>
                </tr>
                </template>
                <tr v-else-if="!loading">
                <td colspan="5" class="brutalist-td text-center py-10 font-bold uppercase">Listen er tom.</td>
                </tr>
            </tbody>
            </table>
        </div>

        <!-- Treasury Tab Content -->
        <div v-if="adminTab === 'treasury'" class="overflow-x-auto shadow-[12px_12px_0px_0px_rgba(0,0,0,1)] mb-12">
             <table class="brutalist-table bg-white">
                <thead>
                    <tr>
                        <th class="brutalist-th">Org-Enhet</th>
                        <th class="brutalist-th">LTREE Path</th>
                        <th class="brutalist-th text-right">Lokal Saldo</th>
                        <th class="brutalist-th text-right bg-black text-white">Branch Sum</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="item in treasuryReport" :key="item.id" class="hover:bg-indigo-50 transition-colors">
                        <td class="brutalist-td">
                            <div class="text-lg font-black uppercase">{{ item.name }}</div>
                            <div class="text-[10px] text-gray-400 font-mono">{{ item.id }}</div>
                        </td>
                        <td class="brutalist-td font-bold text-xs">
                            {{ item.path }}
                        </td>
                        <td class="brutalist-td text-right">
                            <span :class="['font-bold', (item.local_balance || 0) < 0 ? 'text-red-600' : 'text-black']">
                                {{ ((item.local_balance || 0) / 100).toFixed(2) }} kr
                            </span>
                        </td>
                        <td class="brutalist-td text-right font-black bg-gray-100 border-l-4">
                            <span :class="(item.total_branch_balance || 0) < 0 ? 'text-red-600' : 'text-black'">
                                {{ ((item.total_branch_balance || 0) / 100).toFixed(2) }} kr
                            </span>
                        </td>
                    </tr>
                </tbody>
             </table>
        </div>

        <!-- Org Manager Tab -->
        <div v-if="adminTab === 'orgs'">
            <OrgManager />
        </div>

        <!-- Stats Tab -->
        <div v-if="adminTab === 'stats'">
            <StatsDashboard />
        </div>
      </div>
    </div>

    <!-- Register Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black bg-opacity-80 transition-opacity" @click="showModal = false"></div>
      <div class="relative bg-white border-8 border-black p-8 shadow-[16px_16px_0px_0px_rgba(0,0,0,1)] transform transition-all sm:max-w-lg sm:w-full">
        <RegisterForm @registered="onMemberRegistered" @cancel="showModal = false" />
      </div>
    </div>
  </div>
</template>
