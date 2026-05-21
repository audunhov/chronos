<script setup lang="ts">
import { Handle, Position, type NodeProps } from '@vue-flow/core'

const props = defineProps<NodeProps<{
    label: string;
    inputs: string[];
    outputs: string[];
}>>()
</script>

<template>
  <div class="bg-white border-4 border-black p-4 min-w-[200px] shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
    <div class="mb-4 relative">
        <Handle 
            type="target" 
            :position="Position.Left" 
            id="default"
            class="!w-3 !h-3 !bg-black !border-none"
        />
        <p class="text-[8px] font-black uppercase text-blue-600 ml-4">Handling</p>
        <h4 class="text-sm font-black uppercase italic tracking-tighter ml-4">{{ data.label }}</h4>
    </div>

    <div class="grid grid-cols-2 gap-4">
        <!-- Inputs Column -->
        <div class="space-y-2">
            <div v-for="inp in data.inputs" :key="inp" class="flex items-center gap-2 relative">
                <Handle 
                    type="target" 
                    :position="Position.Left" 
                    :id="inp"
                    class="!w-2 !h-2 !bg-gray-400 !border-none"
                />
                <span class="text-[8px] font-black text-gray-400">{{ inp }}</span>
            </div>
        </div>

        <!-- Outputs Column -->
        <div class="space-y-2 text-right">
            <div v-for="out in data.outputs" :key="out" class="flex justify-end items-center gap-2 relative">
                <span class="text-[8px] font-black text-green-600">{{ out }}</span>
                <Handle 
                    type="source" 
                    :position="Position.Right" 
                    :id="out"
                    class="!w-3 !h-3 !bg-black !border-none"
                />
            </div>
        </div>
    </div>
  </div>
</template>
