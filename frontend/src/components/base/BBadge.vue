<template>
  <span 
    class="px-3 py-1 border-2 border-black font-bold text-[10px] uppercase rounded-none tracking-widest inline-block"
    :class="props.class"
  >
    <slot />
  </span>
</template>

<script setup lang="ts">
const props = defineProps<{
  class?: string
}>()
</script>
