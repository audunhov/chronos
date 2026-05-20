<template>
  <button 
    :disabled="loading"
    class="inline-flex items-center justify-center font-bold px-6 py-3 border-4 border-black transition-all active:translate-x-[2px] active:translate-y-[2px] active:shadow-none disabled:opacity-50 disabled:cursor-not-allowed"
    :class="[
      variant === 'primary' ? 'bg-yellow-400 hover:bg-yellow-300 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]' : '',
      variant === 'secondary' ? 'bg-white hover:bg-gray-100 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]' : '',
      variant === 'danger' ? 'bg-red-500 text-white hover:bg-red-400 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]' : '',
      variant === 'ghost' ? 'border-2 py-2 px-4 shadow-none hover:bg-gray-50' : '',
      variant === 'success' ? 'bg-green-400 hover:bg-green-300 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]' : '',
      props.class
    ]"
    @click="$emit('click', $event)"
  >
    <span v-if="loading" class="mr-2 animate-spin">⧖</span>
    <slot />
  </button>
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost' | 'success'
  loading?: boolean
  class?: string
}>(), {
  variant: 'primary',
  loading: false
})

defineEmits(['click'])
</script>
