<template>
  <div class="flex flex-col">
    <label v-if="label" :for="id" class="block text-sm font-black uppercase tracking-tighter mb-1">
      {{ label }}
    </label>
    <input 
      :id="id"
      :type="type"
      :placeholder="placeholder"
      :value="modelValue"
      :required="required"
      class="border-4 border-black p-3 font-mono focus:outline-none focus:bg-blue-50 focus:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] transition-all placeholder:text-gray-400"
      :class="props.class"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  modelValue: string | number
  label?: string
  id?: string
  type?: string
  placeholder?: string
  required?: boolean
  class?: string
}>(), {
  type: 'text',
  required: false
})

defineEmits(['update:modelValue'])
</script>
