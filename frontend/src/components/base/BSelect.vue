<template>
  <div class="flex flex-col">
    <label v-if="label" :for="id" class="block text-sm font-black uppercase tracking-tighter mb-1">
      {{ label }}
    </label>
    <select 
      :id="id"
      :value="modelValue"
      :required="required"
      class="border-4 border-black p-3 font-mono focus:outline-none focus:bg-blue-50 focus:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] transition-all bg-white"
      :class="props.class"
      @change="$emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
    >
      <slot />
    </select>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  modelValue: string | number
  label?: string
  id?: string
  required?: boolean
  class?: string
}>(), {
  required: false
})

defineEmits(['update:modelValue'])
</script>
