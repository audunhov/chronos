<script setup lang="ts">
import { computed } from 'vue'
import { Handle, Position, type NodeProps, useVueFlow } from '@vue-flow/core'

const props = defineProps<NodeProps<{
    type: 'if' | 'filter' | 'list_filter';
    inputs: string[];
    value1?: string;
    operator?: string;
    value2?: string;
    highlightClass?: string;
}>>()

const { updateNodeData } = useVueFlow()

const updateData = (key: string, val: string) => {
    updateNodeData(props.id, { [key]: val })
}

const isRef1 = computed(() => props.data.value1 && props.data.value1.includes('.'))
const isRef2 = computed(() => props.data.value2 && props.data.value2.includes('.'))
</script>

<template>
  <div :class="['bg-yellow-50 border-4 border-black p-4 min-w-[240px] shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] transition-all', selected ? 'ring-8 ring-yellow-400' : '', props.data.highlightClass || '']">
    <div class="mb-4 relative">
        <Handle 
            type="target" 
            :position="Position.Left" 
            id="default"
            class="!w-4 !h-4 !bg-black !border-4 !border-yellow-400 !-left-6"
        />
        <p class="text-[8px] font-black uppercase text-yellow-600">Logikk</p>
        <h4 class="text-sm font-black uppercase italic tracking-tighter">
            {{ props.data.type === 'if' ? 'IF / THEN' : (props.data.type === 'list_filter' ? 'FILTER LIST' : 'STOP FILTER') }}
        </h4>
    </div>

    <div class="space-y-4">
        <!-- Configuration Inputs -->
        <div class="space-y-2">
            <div class="relative">
                <input 
                    :value="props.data.value1" 
                    @input="(e: any) => updateData('value1', e.target.value)"
                    :placeholder="props.data.type === 'list_filter' ? 'Mål-liste (ref)' : 'Verdi 1 (eller ref)'"
                    :class="['w-full border-2 border-black p-1 text-[10px] font-mono outline-none pr-6', isRef1 ? 'bg-purple-100 text-purple-900 font-bold' : 'bg-white focus:bg-yellow-100']"
                />
                <span v-if="isRef1" class="absolute right-1 top-1 text-[8px] opacity-50" title="Tolkes som referanse">🔗</span>
            </div>
            
            <select 
                :value="props.data.operator" 
                @change="(e: any) => updateData('operator', e.target.value)"
                class="w-full bg-black text-white border-2 border-black p-1 text-[10px] font-black outline-none"
            >
                <option value="==">ER LIK (==)</option>
                <option value="!=">ER ULIK (!=)</option>
                <option value=">">STØRRE ENN (>)</option>
                <option value="<">MINDRE ENN (<)</option>
                <option value="contains">INNEHOLDER</option>
                <option value="exists">EKSISTERER</option>
            </select>

            <div class="relative">
                <input 
                    :value="props.data.value2" 
                    @input="(e: any) => updateData('value2', e.target.value)"
                    :placeholder="props.data.type === 'list_filter' ? 'Filter-verdi' : 'Verdi 2'"
                    :class="['w-full border-2 border-black p-1 text-[10px] font-mono outline-none pr-6', isRef2 ? 'bg-purple-100 text-purple-900 font-bold' : 'bg-white focus:bg-yellow-100']"
                />
                <span v-if="isRef2" class="absolute right-1 top-1 text-[8px] opacity-50" title="Tolkes som referanse">🔗</span>
            </div>
        </div>

        <div v-if="props.data.type === 'if'" class="space-y-3 text-right">
            <div class="flex justify-end items-center gap-2 relative group">
                <span class="text-[10px] font-black text-green-600 uppercase italic">SANN</span>
                <Handle 
                    type="source" 
                    :position="Position.Right" 
                    id="true" 
                    class="!w-4 !h-4 !bg-green-600 !border-4 !border-black !-right-6" 
                />
            </div>
            <div class="flex justify-end items-center gap-2 relative group">
                <span class="text-[10px] font-black text-red-600 uppercase italic">USANN</span>
                <Handle 
                    type="source" 
                    :position="Position.Right" 
                    id="false" 
                    class="!w-4 !h-4 !bg-red-600 !border-4 !border-black !-right-6" 
                />
            </div>
        </div>
        <div v-else class="text-right space-y-3">
            <div class="flex justify-end items-center gap-2 relative group">
                <span class="text-[8px] font-black text-black uppercase italic">{{ props.data.type === 'list_filter' ? 'Filtrert Liste' : 'Passér' }}</span>
                <Handle 
                    type="source" 
                    :position="Position.Right" 
                    :id="props.data.type === 'list_filter' ? 'filtered_list' : 'default'" 
                    class="!w-4 !h-4 !bg-black !border-4 !border-yellow-400 !-right-6" 
                />
            </div>
        </div>
    </div>
  </div>
</template>
