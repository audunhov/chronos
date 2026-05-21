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
import ProfileManager from './components/ProfileManager.vue'
import ReactionManager from './components/ReactionManager.vue'
import OrganListManager from './components/OrganListManager.vue'
import AuditTrail from './components/AuditTrail.vue'
import BButton from './components/base/BButton.vue'
import BCard from './components/base/BCard.vue'
import BBadge from './components/base/BBadge.vue'
import BInput from './components/base/BInput.vue'
import BSelect from './components/base/BSelect.vue'

const members = ref<Member[]>([])
const myMemberships = ref<MyMembership[]>([])
const organizations = ref<{id: string, name: string}[]>([])
const selectedOrg = ref('')
const loading = ref(false)
const error = ref('')
const selectedDate = ref('')
const showModal = ref(false)
const currentView = ref<'admin' | 'me'>('me')
const adminTab = ref<'members' | 'treasury' | 'orgs' | 'stats' | 'forms' | 'pipelines' | 'organs' | 'audit'>('members')
const meTab = ref<'memberships' | 'forms' | 'profile'>('memberships')
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
    organizations.value = await api.getOrganizations()
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

onMounted(() => {
  if (auth.user) {
    fetchMyMemberships()
    fetchOrganizations()
  }
})

// Reager på org-bytte eller dato-bytte i admin
watch([selectedOrg, selectedDate, currentView, adminTab], ([org, date, view, tab]) => {
    if (view === 'admin') {
        if (tab === 'members') fetchMembers()
        if (tab === 'treasury') fetchTreasuryReport()
    }
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
  }
})
</script>

<template>
  <div class="min-h-screen p-4 md:p-8 flex flex-col items-center">
    
    <!-- LOGIN SCREEN -->
    <div v-if="!auth.user" class="flex-1 flex items-center justify-center w-full">
      <LoginForm />
    </div>

    <!-- MAIN APP -->
    <div v-else class="w-full max-w-7xl">
      <header class="flex flex-col md:flex-row justify-between items-start md:items-center mb-12 gap-6 border-b-8 border-black pb-8">
        <div>
          <h1 class="text-7xl font-black uppercase italic tracking-tighter leading-none mb-2">CHRONOS</h1>
          <div class="flex gap-2">
            <BBadge class="bg-black text-white italic">PROD-ENV-ACTIVE</BBadge>
            <BBadge class="bg-yellow-400">IDENT: {{ auth.user.email }}</BBadge>
          </div>
        </div>

        <div class="flex flex-wrap gap-4">
            <BButton 
                @click="currentView = 'me'"
                :variant="currentView === 'me' ? 'primary' : 'secondary'"
                class="text-xs italic"
            >
                MINE SIDER
            </BButton>
            <BButton 
                v-if="auth.user?.role === 'admin'"
                @click="currentView = 'admin'"
                :variant="currentView === 'admin' ? 'primary' : 'secondary'"
                class="text-xs italic"
            >
                ADMIN
            </BButton>
            <BButton 
                @click="handleLogout" 
                variant="danger"
                class="text-xs italic"
            >
                LOGG UT
            </BButton>
        </div>
      </header>

      <div v-if="error" class="bg-red-500 text-white border-4 border-black p-6 mb-8 shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] font-black italic uppercase">
        SYSTEM ERROR :: {{ error }}
      </div>

      <!-- PERSONAL VIEW -->
      <div v-if="currentView === 'me'">
        <div class="flex flex-col xl:flex-row xl:items-end justify-between mb-12 gap-6 border-b-4 border-black pb-8">
            <nav class="flex flex-wrap gap-4">
                <BButton 
                    @click="meTab = 'memberships'"
                    :variant="meTab === 'memberships' ? 'primary' : 'ghost'"
                    class="text-sm uppercase italic"
                >
                    MINE MEDLEMSKAP
                </BButton>
                <BButton 
                    @click="meTab = 'forms'"
                    :variant="meTab === 'forms' ? 'primary' : 'ghost'"
                    class="text-sm uppercase italic"
                >
                    UNDERSØKELSER
                </BButton>
                <BButton 
                    @click="meTab = 'profile'"
                    :variant="meTab === 'profile' ? 'primary' : 'ghost'"
                    class="text-sm uppercase italic"
                >
                    MIN PROFIL
                </BButton>
            </nav>
            <BButton
                v-if="meTab === 'memberships'"
                @click="showModal = true"
                variant="success"
                class="text-xl italic"
            >
                + BLI MEDLEM
            </BButton>
        </div>

        <div v-if="meTab === 'memberships'">
            <div v-if="loading" class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
                <BCard v-for="i in 3" :key="i" class="animate-pulse h-48 bg-gray-100"></BCard>
            </div>
            <div v-else-if="myMemberships.length > 0" class="grid grid-cols-1 gap-8 sm:grid-cols-2 lg:grid-cols-3">
                <BCard v-for="ms in myMemberships" :key="ms.id" class="flex flex-col h-full group hover:bg-orange-50">
                    <div class="flex justify-between items-start mb-6">
                        <div>
                            <h3 class="font-black text-3xl uppercase leading-none mb-2 group-hover:text-blue-600 transition-colors">{{ ms.org_name }}</h3>
                            <BBadge class="bg-gray-100">{{ ms.role }}</BBadge>
                        </div>
                        <BBadge :class="ms.status === 'ACTIVE' ? 'bg-green-400' : 'bg-red-400'">
                            {{ ms.status }}
                        </BBadge>
                    </div>

                    <div class="mt-auto pt-6 border-t-2 border-black border-dashed flex justify-between items-end">
                        <div>
                            <p class="text-[10px] font-black uppercase text-gray-400">Saldo</p>
                            <p class="text-2xl font-black italic" :class="ms.balance < 0 ? 'text-red-500' : 'text-green-600'">
                                {{ ms.balance }} NOK
                            </p>
                        </div>
                        <div class="text-right">
                            <p class="text-[10px] font-black uppercase text-gray-400">Sist oppdatert</p>
                            <p class="text-[10px] font-bold">{{ new Date(ms.updated_at).toLocaleDateString() }}</p>
                        </div>
                    </div>
                </BCard>
            </div>
            <div v-else class="py-32 text-center border-8 border-black border-dashed bg-white">
                <h3 class="text-6xl font-black uppercase italic opacity-20 mb-4 underline">Utenfor</h3>
                <p class="font-bold uppercase tracking-widest text-gray-400">Du er ikke medlem i noen organisasjoner ennå.</p>
            </div>
        </div>

        <div v-if="meTab === 'forms'">
            <FormManager />
        </div>

        <div v-if="meTab === 'profile'">
            <ProfileManager />
        </div>
      </div>

      <!-- ADMIN VIEW -->
      <div v-if="currentView === 'admin'">
        <div class="flex flex-col xl:flex-row xl:items-end justify-between mb-8 gap-6 border-b-4 border-black pb-6">
            <nav class="flex flex-wrap gap-4">
                <BButton 
                    @click="adminTab = 'members'"
                    :variant="adminTab === 'members' ? 'primary' : 'ghost'"
                    class="text-sm uppercase italic"
                >
                    MEDLEMSLISTE
                </BButton>
                <BButton 
                    @click="adminTab = 'treasury'"
                    :variant="adminTab === 'treasury' ? 'primary' : 'ghost'"
                    class="text-sm uppercase italic"
                >
                    FINANS
                </BButton>
                <BButton 
                    @click="adminTab = 'orgs'"
                    :variant="adminTab === 'orgs' ? 'primary' : 'ghost'"
                    class="text-sm uppercase italic"
                >
                    STRUKTUR
                </BButton>
                <BButton 
                    @click="adminTab = 'stats'"
                    :variant="adminTab === 'stats' ? 'primary' : 'ghost'"
                    class="text-sm uppercase italic"
                >
                    STATISTIKK
                </BButton>
                <BButton 
                    @click="adminTab = 'forms'"
                    :variant="adminTab === 'forms' ? 'primary' : 'ghost'"
                    class="text-sm uppercase italic"
                >
                    SKJEMAER
                </BButton>
                <BButton 
                    @click="adminTab = 'pipelines'"
                    :variant="adminTab === 'pipelines' ? 'primary' : 'ghost'"
                    class="text-sm uppercase italic"
                >
                    PIPELINES
                </BButton>
                <BButton 
                    @click="adminTab = 'organs'"
                    :variant="adminTab === 'organs' ? 'primary' : 'ghost'"
                    class="text-sm uppercase italic"
                >
                    ORGANER
                </BButton>
                <BButton 
                    @click="adminTab = 'audit'"
                    :variant="adminTab === 'audit' ? 'primary' : 'ghost'"
                    class="text-sm uppercase italic"
                >
                    AUDIT LOG
                </BButton>
            </nav>
        </div>

        <!-- Members Tab -->
        <div v-if="adminTab === 'members'" class="space-y-8">
            <div class="flex flex-wrap gap-6 items-end bg-black text-white p-6 shadow-[8px_8px_0px_0px_rgba(0,0,0,0.3)]">
                <BSelect v-model="selectedOrg" label="Org-Filter" class="bg-black text-white border-white min-w-[250px]">
                    <option value="">ALLE ORGANISASJONER</option>
                    <option v-for="org in organizations" :key="org.id" :value="org.id">{{ org.name }}</option>
                </BSelect>

                <BInput v-model="selectedDate" type="date" label="Historisk dato (As-Of)" class="bg-black text-white border-white" />
                
                <BButton @click="fetchMembers()" variant="secondary" class="italic">OPPDATER</BButton>
            </div>

            <div class="overflow-x-auto border-4 border-black shadow-[16px_16px_0px_0px_rgba(0,0,0,1)]">
                <table class="brutalist-table">
                <thead>
                    <tr>
                    <th class="brutalist-th">Identitet</th>
                    <th class="brutalist-th">Org</th>
                    <th class="brutalist-th">Status</th>
                    <th class="brutalist-th">Rolle</th>
                    <th class="brutalist-th">Saldo</th>
                    <th class="brutalist-th">Sist endret</th>
                    <th class="brutalist-th">Handlinger</th>
                    </tr>
                </thead>
                <tbody class="bg-white">
                    <tr v-if="loading" v-for="i in 5">
                        <td colspan="7" class="brutalist-td animate-pulse bg-gray-50 h-12"></td>
                    </tr>
                    <tr v-else v-for="m in members" :key="m.id" class="hover:bg-blue-50 transition-colors">
                    <td class="brutalist-td">
                        <div class="font-black uppercase tracking-tighter">{{ m.name }}</div>
                        <div class="text-[10px] font-bold text-gray-400">{{ m.email }}</div>
                    </td>
                    <td class="brutalist-td"><BBadge class="bg-gray-100">{{ m.org_id.split('-')[0] }}</BBadge></td>
                    <td class="brutalist-td">
                        <BBadge :class="m.status === 'ACTIVE' ? 'bg-green-400' : 'bg-red-400'">{{ m.status }}</BBadge>
                    </td>
                    <td class="brutalist-td font-bold uppercase text-xs">{{ m.role }}</td>
                    <td class="brutalist-td font-black" :class="m.balance < 0 ? 'text-red-600' : 'text-green-600'">{{ m.balance }}</td>
                    <td class="brutalist-td text-[10px] font-bold">{{ new Date(m.updated_at).toLocaleString() }}</td>
                    <td class="brutalist-td">
                        <BButton 
                            v-if="m.status !== 'SHREDDED'"
                            @click="shredMember(m.id)" 
                            variant="danger"
                            class="text-[8px] py-1 px-2 uppercase"
                        >
                            Glem (GDPR)
                        </BButton>
                    </td>
                    </tr>
                </tbody>
                </table>
            </div>
        </div>

        <!-- Treasury Tab -->
        <div v-if="adminTab === 'treasury'" class="grid grid-cols-1 gap-8 md:grid-cols-2 lg:grid-cols-3">
            <BCard v-for="t in treasuryReport" :key="t.id" class="flex flex-col group hover:bg-green-50">
                <div class="mb-4">
                    <p class="text-[8px] font-black text-gray-400 tracking-widest">{{ t.path }}</p>
                    <h3 class="text-3xl font-black uppercase italic leading-none">{{ t.name }}</h3>
                </div>
                
                <div class="grid grid-cols-2 gap-4 mt-6">
                    <div class="p-3 border-2 border-black bg-white">
                        <p class="text-[8px] font-black uppercase opacity-50">Lokal Saldo</p>
                        <p class="text-xl font-black italic">{{ t.local_balance }}</p>
                    </div>
                    <div class="p-3 border-2 border-black bg-black text-white">
                        <p class="text-[8px] font-black uppercase opacity-50">Branch Total</p>
                        <p class="text-xl font-black italic">{{ t.total_branch_balance }}</p>
                    </div>
                </div>
            </BCard>
        </div>

        <!-- Org Manager Tab -->
        <div v-if="adminTab === 'orgs'">
            <OrgManager />
        </div>

        <!-- Stats Tab -->
        <div v-if="adminTab === 'stats'">
            <StatsDashboard />
        </div>

        <!-- Forms Admin Tab -->
        <div v-if="adminTab === 'forms'">
            <FormManager :isAdmin="true" />
        </div>

        <!-- Pipelines Admin Tab -->
        <div v-if="adminTab === 'pipelines'">
            <ReactionManager :orgs="organizations" />
        </div>

        <!-- Organs Admin Tab -->
        <div v-if="adminTab === 'organs'">
            <OrganListManager :orgs="organizations" />
        </div>

        <!-- Audit Admin Tab -->
        <div v-if="adminTab === 'audit'">
            <AuditTrail :orgs="organizations" />
        </div>
      </div>
    </div>

    <!-- Register Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm">
      <BCard class="relative w-full max-w-lg">
        <button @click="showModal = false" class="absolute -top-4 -right-4 bg-black text-white w-10 h-10 border-4 border-black font-black flex items-center justify-center hover:bg-gray-800 shadow-[4px_4px_0px_0px_rgba(255,255,255,1)]">X</button>
        <RegisterForm :organizations="organizations" @registered="onMemberRegistered" @cancel="showModal = false" />
      </BCard>
    </div>
  </div>
</template>
