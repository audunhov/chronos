<script setup lang="ts">
import { Handle, Position, type NodeProps, useVueFlow } from '@vue-flow/core'

const props = defineProps<NodeProps<{
    type: 'if' | 'filter' | 'list_filter';
    inputs: string[];
    value1?: string;
    operator?: string;
    value2?: string;
}>>()

const { updateNodeData } = useVueFlow()

const updateData = (key: string, val: string) => {
    updateNodeData(props.id, { [key]: val })
}
</script>

<template>
  <div :class="['bg-yellow-50 border-4 border-black p-4 min-w-[240px] shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] transition-all', selected ? 'ring-8 ring-yellow-400' : '']">
    <div class="mb-4 relative">
        <Handle 
            type="target" 
            :position="Position.Left" 
            id="default"
            class="!w-4 !h-4 !bg-black !border-4 !border-yellow-400 !-left-6"
        />
        <p class="text-[8px] font-black uppercase text-yellow-600">Logikk</p>
        <h4 class="text-sm font-black uppercase italic tracking-tighter">
            {{ data.type === 'if' ? 'IF / THEN' : (data.type === 'list_filter' ? 'FILTER LIST' : 'STOP FILTER') }}
        </h4>
    </div>

    <div class="space-y-4">
        <!-- Configuration Inputs -->
        <div class="space-y-2">
            <input 
                :value="data.value1" 
                @input="(e: any) => updateData('value1', e.target.value)"
                :placeholder="data.type === 'list_filter' ? 'Mål-liste (ref)' : 'Verdi 1 (eller ref)'"
                class="w-full bg-white border-2 border-black p-1 text-[10px] font-mono focus:bg-yellow-100 outline-none"
            />
            
            <select 
                :value="data.operator" 
                @change="(e: any) => updateData('operator', e.target.value)"
                class="w-full bg-black text-white border-2 border-black p-1 text-[10px] font-black"
            >
                <option value="==">ER LIK (==)</option>
                <option value="!=">ER ULIK (!=)</option>
                <option value="contains">INNEHOLDER</option>
                <option value="exists">EKSISTERER</option>
            </select>

            <input 
                :value="data.value2" 
                @input="(e: any) => updateData('value2', e.target.value)"
                :placeholder="data.type === 'list_filter' ? 'Filter-verdi' : 'Verdi 2'"
                class="w-full bg-white border-2 border-black p-1 text-[10px] font-mono focus:bg-yellow-100 outline-none"
            />
        </div>

        <div v-if="data.type === 'if'" class="space-y-3 text-right">
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
                <span class="text-[8px] font-black text-black uppercase italic">{{ data.type === 'list_filter' ? 'Filtrert Liste' : 'Passér' }}</span>
                <Handle 
                    type="source" 
                    :position="Position.Right" 
                    :id="data.type === 'list_filter' ? 'filtered_list' : 'default'" 
                    class="!w-4 !h-4 !bg-black !border-4 !border-yellow-400 !-right-6" 
                />
            </div>
        </div>
    </div>
  </div>
</template>
