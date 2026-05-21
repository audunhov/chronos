<script setup lang="ts">
import { Handle, Position, type NodeProps } from '@vue-flow/core'

const props = defineProps<NodeProps<{
    type: 'if' | 'filter';
    inputs: string[];
    condition?: string;
}>>()
</script>

<template>
  <div class="bg-yellow-50 border-4 border-black p-4 min-w-[180px] shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
    <div class="mb-4 relative">
        <Handle 
            type="target" 
            :position="Position.Left" 
            id="default"
            class="!w-3 !h-3 !bg-black !border-none"
        />
        <p class="text-[8px] font-black uppercase text-yellow-600 ml-4">Logikk</p>
        <h4 class="text-sm font-black uppercase italic tracking-tighter ml-4">{{ data.type === 'if' ? 'IF / THEN' : 'FILTER' }}</h4>
    </div>

    <div v-if="data.type === 'if'" class="space-y-4 text-right">
        <div class="flex justify-end items-center gap-2 relative">
            <span class="text-[10px] font-black text-green-600">SANN</span>
            <Handle type="source" :position="Position.Right" id="true" class="!w-3 !h-3 !bg-green-600 !border-none" />
        </div>
        <div class="flex justify-end items-center gap-2 relative">
            <span class="text-[10px] font-black text-red-600">USANN</span>
            <Handle type="source" :position="Position.Right" id="false" class="!w-3 !h-3 !bg-red-600 !border-none" />
        </div>
    </div>
    <div v-else class="text-right">
        <div class="flex justify-end items-center gap-2 relative">
            <span class="text-[10px] font-black text-black">PASSÉR</span>
            <Handle type="source" :position="Position.Right" id="default" class="!w-3 !h-3 !bg-black !border-none" />
        </div>
    </div>
  </div>
</template>
