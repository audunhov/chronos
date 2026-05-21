<script setup lang="ts">
import { useRoute } from 'vue-router'
import BButton from '../components/base/BButton.vue'

const props = defineProps<{
    orgs: { id: string, name: string }[]
}>()

const route = useRoute()
</script>

<template>
  <div class="space-y-8">
    <div class="flex flex-col xl:flex-row xl:items-end justify-between mb-8 gap-6 border-b-4 border-black pb-6">
        <nav class="flex flex-wrap gap-4">
            <router-link to="/admin/members" custom v-slot="{ navigate, isExactActive }">
                <BButton 
                    @click="navigate"
                    :variant="isExactActive ? 'primary' : 'ghost'"
                    class="text-sm uppercase italic"
                >
                    MEDLEMSLISTE
                </BButton>
            </router-link>
            <router-link to="/admin/treasury" custom v-slot="{ navigate, isExactActive }">
                <BButton 
                    @click="navigate"
                    :variant="isExactActive ? 'primary' : 'ghost'"
                    class="text-sm uppercase italic"
                >
                    FINANS
                </BButton>
            </router-link>
            <router-link to="/admin/orgs" custom v-slot="{ navigate, isExactActive }">
                <BButton 
                    @click="navigate"
                    :variant="isExactActive ? 'primary' : 'ghost'"
                    class="text-sm uppercase italic"
                >
                    STRUKTUR
                </BButton>
            </router-link>
            <router-link to="/admin/stats" custom v-slot="{ navigate, isExactActive }">
                <BButton 
                    @click="navigate"
                    :variant="isExactActive ? 'primary' : 'ghost'"
                    class="text-sm uppercase italic"
                >
                    STATISTIKK
                </BButton>
            </router-link>
            <router-link to="/admin/forms" custom v-slot="{ navigate, isExactActive }">
                <BButton 
                    @click="navigate"
                    :variant="isExactActive ? 'primary' : 'ghost'"
                    class="text-sm uppercase italic"
                >
                    SKJEMAER
                </BButton>
            </router-link>
            <router-link to="/admin/pipelines" custom v-slot="{ navigate, isExactActive }">
                <BButton 
                    @click="navigate"
                    :variant="isExactActive ? 'primary' : 'ghost'"
                    class="text-sm uppercase italic"
                >
                    PIPELINES
                </BButton>
            </router-link>
            <router-link to="/admin/organs" custom v-slot="{ navigate, isExactActive }">
                <BButton 
                    @click="navigate"
                    :variant="isExactActive ? 'primary' : 'ghost'"
                    class="text-sm uppercase italic"
                >
                    ORGANER
                </BButton>
            </router-link>
            <router-link to="/admin/audit" custom v-slot="{ navigate, isExactActive }">
                <BButton 
                    @click="navigate"
                    :variant="isExactActive ? 'primary' : 'ghost'"
                    class="text-sm uppercase italic"
                >
                    AUDIT LOG
                </BButton>
            </router-link>
        </nav>
    </div>

    <router-view :orgs="props.orgs" />
  </div>
</template>
