<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import UserInspector from '../components/UserInspector.vue'
import BButton from '../components/base/BButton.vue'

const route = useRoute()
const router = useRouter()
const userId = route.params.id as string

const goBack = () => {
    router.back()
}
</script>

<template>
  <div class="space-y-6 animate-in fade-in">
    <BButton @click="goBack" variant="secondary" class="text-xs">
        ← TILBAKE
    </BButton>
    <UserInspector :user-id="userId" />
  </div>
</template>
