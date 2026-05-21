<script setup lang="ts">
import { computed } from 'vue'
import { Handle, Position, type NodeProps } from '@vue-flow/core'

const props = defineProps<NodeProps<{
    label: string;
    inputs: string[];
    outputs: string[];
    connectedInputs?: Set<string>;
    referenceData?: {
        organizations: any[];
        organs: any[];
        roles: string[];
    };
    onUpdateData?: (key: string, val: any) => void;
    // Dynamic values stored in data
    [key: string]: any;
}>>()

const isConnected = (id: string) => props.data.connectedInputs?.has(id)

// Helper to determine if an input should show a dropdown
const getOptionsForInput = (inp: string) => {
    if (inp === 'org_id' || inp === 'start_org_id') return props.data.referenceData?.organizations.map(o => ({ label: o.name, value: o.id }))
    if (inp === 'organ_id' || inp === 'organ_name') return props.data.referenceData?.organs.map(o => ({ label: o.name, value: o.id }))
    if (inp === 'role_name' || inp === 'role_type') return props.data.referenceData?.roles.map(r => ({ label: r, value: r }))
    return null
}
</script>

<template>
  <div :class="['bg-white border-4 border-black p-4 min-w-[280px] shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] transition-all', selected ? 'ring-8 ring-blue-400' : '']">
    <div class="mb-4 relative">
        <Handle 
            type="target" 
            :position="Position.Left" 
            id="default"
            class="!w-4 !h-4 !bg-black !border-4 !border-blue-400 !-left-6"
        />
        <p class="text-[8px] font-black uppercase text-blue-600">Handling</p>
        <h4 class="text-sm font-black uppercase italic tracking-tighter">{{ data.label }}</h4>
    </div>

    <div class="space-y-4">
        <!-- Inputs Column -->
        <div class="space-y-3">
            <div v-for="inp in data.inputs" :key="inp" class="relative">
                <Handle 
                    type="target" 
                    :position="Position.Left" 
                    :id="inp"
                    :class="['!w-3 !h-3 !border-2 !border-black !-left-6', isConnected(inp) ? '!bg-blue-400' : '!bg-gray-200']"
                />
                
                <div class="flex flex-col gap-1">
                    <span class="text-[8px] font-black text-gray-500 uppercase">{{ inp }}</span>
                    
                    <!-- Show input field only if NOT connected -->
                    <template v-if="!isConnected(inp)">
                        <select v-if="getOptionsForInput(inp)"
                            :value="data[inp]"
                            @change="(e: any) => data.onUpdateData?.(inp, e.target.value)"
                            class="w-full bg-blue-50 border-2 border-black p-1 text-[10px] font-bold focus:outline-none"
                        >
                            <option value="">Velg {{ inp }}...</option>
                            <option v-for="opt in getOptionsForInput(inp)" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
                        </select>
                        <textarea v-else-if="inp === 'template' || inp === 'body'"
                            :value="data[inp]"
                            @input="(e: any) => data.onUpdateData?.(inp, e.target.value)"
                            placeholder="Skriv tekst her..."
                            class="w-full bg-blue-50 border-2 border-black p-1 text-[10px] font-mono h-16 resize-none focus:outline-none"
                        ></textarea>
                        <input v-else
                            :value="data[inp]"
                            @input="(e: any) => data.onUpdateData?.(inp, e.target.value)"
                            :placeholder="`Skriv inn ${inp}...`"
                            class="w-full bg-blue-50 border-2 border-black p-1 text-[10px] font-mono focus:outline-none"
                        />
                    </template>
                    <div v-else class="bg-gray-100 border-2 border-dashed border-gray-300 p-1 text-[9px] italic text-gray-400">
                        Koblet til variabel
                    </div>
                </div>
            </div>
        </div>

        <div class="border-t-2 border-black pt-2 space-y-2">
            <p class="text-[8px] font-black uppercase text-green-600 text-right">Resultater</p>
            <div v-for="out in data.outputs" :key="out" class="flex justify-end items-center gap-2 relative">
                <span class="text-[9px] font-bold font-mono tracking-tighter">{{ out }}</span>
                <Handle 
                    type="source" 
                    :position="Position.Right" 
                    :id="out"
                    class="!w-4 !h-4 !bg-black !border-4 !border-green-400 !-right-6"
                />
            </div>
        </div>
    </div>
  </div>
</template>
