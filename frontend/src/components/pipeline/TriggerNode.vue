<script setup lang="ts">
import { Handle, Position, type NodeProps } from '@vue-flow/core'

const props = defineProps<NodeProps<{
    event: string;
    outputs: string[];
}>>()
</script>

<template>
  <div class="bg-purple-100 border-4 border-black p-4 min-w-[200px] shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
    <div class="mb-4">
        <p class="text-[8px] font-black uppercase text-purple-600">Start (Trigger)</p>
        <h4 class="text-sm font-black uppercase italic tracking-tighter">{{ data.event }}</h4>
    </div>

    <div class="space-y-2">
        <div v-for="out in data.outputs" :key="out" class="flex justify-between items-center relative">
            <span class="text-[10px] font-bold font-mono">{{ out }}</span>
            <Handle 
                type="source" 
                :position="Position.Right" 
                :id="out"
                class="!w-3 !h-3 !bg-black !border-none"
            />
        </div>
    </div>
  </div>
</template>
