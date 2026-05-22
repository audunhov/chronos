<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api } from './services/api'
import { auth, logout } from './services/auth'
import RegisterForm from './components/RegisterForm.vue'
import LoginForm from './components/LoginForm.vue'
import BCard from './components/base/BCard.vue'
import BButton from './components/base/BButton.vue'
import OmniSearch from './components/OmniSearch.vue'

const organizations = ref<{id: string, name: string}[]>([])
const loading = ref(false)
const error = ref('')
const showModal = ref(false)

const fetchOrganizations = async () => {
  if (!auth.user) return
  try {
    const data = await api.getOrganizationHierarchy()
    organizations.value = Array.isArray(data) ? data.map(o => ({ id: o.id, name: o.name })) : []
  } catch (e) {
    console.error('Failed to fetch orgs:', e)
  }
}

const handleLogout = () => {
  logout()
}

const onMemberRegistered = () => {
  showModal.value = false
  // Refresh data by triggering event or letting views handle it
}

// Reager på innlogging/utlogging
watch(() => auth.user, (newUser) => {
  if (newUser) {
    fetchOrganizations()
  } else {
    organizations.value = []
  }
}, { immediate: true })

onMounted(() => {
  if (auth.user) {
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
            <router-link to="/" custom v-slot="{ navigate, isActive }">
                <button 
                    @click="navigate"
                    :class="[
                        'px-4 py-2 text-xs font-black uppercase transition-colors',
                        isActive ? 'bg-yellow-400 text-black' : 'bg-black text-white hover:bg-gray-800'
                    ]"
                >
                    DASHBOARD
                </button>
            </router-link>
            <router-link to="/surveys" custom v-slot="{ navigate, isActive }">
                <button 
                    @click="navigate"
                    :class="[
                        'px-4 py-2 text-xs font-black uppercase transition-colors',
                        isActive ? 'bg-purple-400 text-black' : 'bg-black text-white hover:bg-gray-800'
                    ]"
                >
                    UNDERSØKELSER
                </button>
            </router-link>
            <router-link to="/profile" custom v-slot="{ navigate, isActive }">
                <button 
                    @click="navigate"
                    :class="[
                        'px-4 py-2 text-xs font-black uppercase transition-colors',
                        isActive ? 'bg-blue-400 text-black' : 'bg-black text-white hover:bg-gray-800'
                    ]"
                >
                    PROFIL
                </button>
            </router-link>
            <router-link v-if="auth.user?.role === 'admin'" to="/admin" custom v-slot="{ navigate, isActive }">
                <button 
                    @click="navigate"
                    :class="[
                        'px-4 py-2 text-xs font-black uppercase transition-colors',
                        isActive ? 'bg-green-400 text-black' : 'bg-black text-white hover:bg-gray-800'
                    ]"
                >
                    ADMIN
                </button>
            </router-link>
          </nav>

          <button
            @click="handleLogout"
            class="brutalist-btn bg-white hover:bg-red-500 hover:text-white px-4 py-2 text-xs uppercase"
          >
            LOGG UT
          </button>
        </div>
      </div>

      <!-- MAIN CONTENT -->
      <router-view :orgs="organizations" @open-register="showModal = true" />
    </div>

    <!-- Global Register Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm">
      <BCard class="relative bg-white border-8 border-black p-8 shadow-[16px_16px_0px_0px_rgba(0,0,0,1)] transform transition-all sm:max-w-lg sm:w-full">
        <button @click="showModal = false" class="absolute -top-6 -right-6 w-12 h-12 bg-red-500 border-4 border-black text-white font-black flex items-center justify-center hover:bg-gray-800 shadow-[4px_4px_0px_0px_rgba(255,255,255,1)]">X</button>
        <RegisterForm :organizations="organizations" @registered="onMemberRegistered" @cancel="showModal = false" />
      </BCard>
    </div>

    <!-- Global OmniSearch -->
    <OmniSearch v-if="auth.user && auth.user.role === 'admin'" />
  </div>
</template>
