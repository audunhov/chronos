<script setup lang="ts">
import { Handle, Position, type NodeProps } from '@vue-flow/core'

const props = defineProps<NodeProps<{
    event: string;
    aggregateId?: string;
    outputs: string[];
    availableEvents?: string[];
    forms?: any[];
    onUpdateEvent?: (ev: string) => void;
    onUpdateAggregate?: (agg: string) => void;
    highlightClass?: string;
}>>()
</script>

<template>
  <div :class="['bg-purple-50 border-4 border-black p-4 min-w-[220px] shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] transition-all', selected ? 'ring-8 ring-purple-400' : '', props.data.highlightClass || '']">
    <div class="mb-4 space-y-2">
        <p class="text-[8px] font-black uppercase text-purple-600">Start (Trigger)</p>
        <select 
            :value="props.data.event" 
            @change="(e: any) => props.data.onUpdateEvent?.(e.target.value)"
            class="w-full bg-white border-4 border-black p-1 font-black uppercase italic text-xs focus:outline-none focus:bg-yellow-400"
        >
            <option v-for="ev in props.data.availableEvents || []" :key="ev" :value="ev">{{ ev }}</option>
        </select>

        <div v-if="props.data.event === 'TimedSchedule'" class="space-y-1 mt-2">
            <p class="text-[8px] font-black uppercase text-gray-500">Intervall (Cron/Kort)</p>
            <select 
                :value="props.data.aggregateId || ''" 
                @change="(e: any) => props.data.onUpdateAggregate?.(e.target.value)"
                class="w-full bg-white border-2 border-black p-1 font-mono text-[10px] focus:outline-none focus:bg-yellow-400"
            >
                <option value="">Velg intervall...</option>
                <option value="daily">Daglig (00:00)</option>
                <option value="weekly">Ukentlig (Man 00:00)</option>
                <option value="monthly">Månedlig (1. hver mnd)</option>
                <option value="yearly">Årlig (1. jan)</option>
            </select>
        </div>

        <div v-if="props.data.event === 'FormResponseSubmitted'" class="space-y-1 mt-2">
            <p class="text-[8px] font-black uppercase text-gray-500">Skjema</p>
            <select 
                :value="props.data.aggregateId || ''" 
                @change="(e: any) => props.data.onUpdateAggregate?.(e.target.value)"
                class="w-full bg-white border-2 border-black p-1 font-mono text-[10px] focus:outline-none focus:bg-yellow-400"
            >
                <option value="">Velg skjema...</option>
                <option v-for="f in props.data.forms || []" :key="f.id" :value="f.id">{{ f.title }}</option>
            </select>
        </div>
    </div>

    <div class="space-y-3">
        <div v-for="out in props.data.outputs" :key="out" class="flex justify-between items-center relative group">
            <div class="flex items-center gap-1">
                <div class="w-1 h-1 bg-purple-400"></div>
                <span class="text-[9px] font-bold font-mono tracking-tighter">{{ out }}</span>
            </div>
            <Handle 
                type="source" 
                :position="Position.Right" 
                :id="out"
                class="!w-4 !h-4 !bg-black !border-4 !border-purple-400 !-right-6"
            />
        </div>
    </div>
  </div>
</template>

<style scoped>
.vue-flow__handle {
    transition: transform 0.2s;
}
.vue-flow__handle:hover {
    transform: scale(1.5);
}
</style>
