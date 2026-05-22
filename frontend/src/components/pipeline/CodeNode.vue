<script setup lang="ts">
import { computed, watch, ref } from 'vue'
import { Handle, Position, type NodeProps, useVueFlow } from '@vue-flow/core'

const props = defineProps<NodeProps<{
    code: string;
    inputs: string[];
    outputs: string[];
    connectedInputs?: Set<string>;
    onUpdateData?: (key: string, val: any) => void;
    logs?: string[];
    [key: string]: any;
}>>()

const { updateNodeData } = useVueFlow()

const codeValue = ref(props.data.code || '// Bruk variabler som input_navn\nconst hilsen = "Hei " + navn;\nreturn hilsen;')

// Improved detection for more complex JS syntax
const detectInputs = (code: string) => {
    const inputs: string[] = []
    
    // 1. Simple assignments: const x = ...
    const simpleRegex = /(?:const|let|var)\s+([a-zA-Z0-9_]+)\s*=/g
    const simpleMatches = [...code.matchAll(simpleRegex)]
    inputs.push(...simpleMatches.map(m => m[1] || ''))

    // 2. Destructuring: const { x, y } = ...
    const destructuringRegex = /(?:const|let|var)\s*\{\s*([^}]+)\s*\}\s*=/g
    const dMatches = [...code.matchAll(destructuringRegex)]
    dMatches.forEach(m => {
        const content = m[1]
        if (content) {
            const vars = content.split(',').map(v => {
                const parts = v.trim().split(':')
                return (parts[0] || '').trim()
            })
            inputs.push(...vars)
        }
    })

    return Array.from(new Set(inputs)).filter(name => name !== 'result' && name !== '')
}

watch(codeValue, (newCode) => {
    const newInputs = detectInputs(newCode)
    updateNodeData(props.id, { 
        code: newCode,
        inputs: newInputs
    })
})

const isConnected = (id: string) => props.data.connectedInputs?.has(id)
</script>

<template>
  <div :class="['bg-gray-900 border-4 border-black p-4 min-w-[380px] shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] transition-all', selected ? 'ring-8 ring-yellow-400' : '']">
    <div class="mb-4 flex justify-between items-start">
        <div class="relative">
            <Handle 
                type="target" 
                :position="Position.Left" 
                id="default"
                class="!w-4 !h-4 !bg-yellow-400 !border-4 !border-black !-left-6"
            />
            <p class="text-[8px] font-black uppercase text-yellow-400">Custom Code (Goja JS)</p>
            <h4 class="text-sm font-black uppercase italic tracking-tighter text-white">JavaScript Node</h4>
        </div>
        <div class="flex items-center gap-1 bg-black px-2 py-0.5 border border-gray-700">
            <div class="w-1.5 h-1.5 rounded-full bg-green-500 animate-pulse"></div>
            <span class="text-[7px] font-bold text-gray-400 uppercase">Sandbox Active</span>
        </div>
    </div>

    <div class="space-y-4">
        <!-- Editor -->
        <div class="flex flex-col gap-2 relative group">
            <textarea 
                v-model="codeValue"
                placeholder="Skriv JavaScript her..."
                class="w-full bg-black text-green-400 border-4 border-black p-3 font-mono text-[10px] h-48 resize-none focus:outline-none focus:border-yellow-400"
            ></textarea>
            <div class="absolute right-2 top-2 opacity-0 group-hover:opacity-100 transition-opacity">
                <span class="text-[8px] bg-yellow-400 text-black font-black px-1">JS</span>
            </div>
            <p class="text-[7px] text-gray-500 italic">Variabler definert med const/let blir innganger. Bruk return for resultat.</p>
        </div>

        <!-- Debug Area -->
        <div v-if="data.logs && data.logs.length > 0" class="space-y-1">
            <p class="text-[8px] font-black uppercase text-red-400">Konsoll / Debug</p>
            <div class="bg-black border-2 border-red-900/30 p-2 font-mono text-[9px] text-red-200 max-h-24 overflow-y-auto">
                <div v-for="(log, i) in data.logs" :key="i" class="border-b border-red-900/10 last:border-0 py-0.5">
                    {{ log }}
                </div>
            </div>
        </div>

        <!-- Dynamic Inputs -->
        <div class="grid grid-cols-1 gap-2 border-t border-gray-700 pt-3">
            <p class="text-[8px] font-black uppercase text-gray-400">Innganger ({{ data.inputs?.length || 0 }})</p>
            <div v-for="inp in data.inputs" :key="inp" class="relative">
                <Handle 
                    type="target" 
                    :position="Position.Left" 
                    :id="inp"
                    :class="['!w-3 !h-3 !border-2 !border-black !-left-6', isConnected(inp) ? '!bg-blue-400' : '!bg-gray-600']"
                />
                <div class="flex justify-between items-center bg-gray-800 p-1.5 border border-gray-700">
                    <span class="text-[9px] font-bold font-mono text-blue-300">{{ inp }}</span>
                    <input v-if="!isConnected(inp)"
                        :value="data[inp]"
                        @input="(e: any) => data.onUpdateData?.(inp, e.target.value)"
                        placeholder="Sett verdi..."
                        class="w-32 bg-gray-900 text-white text-[9px] border border-gray-700 px-1 focus:border-yellow-400 outline-none"
                    />
                    <span v-else class="text-[7px] text-blue-400/50 uppercase font-black tracking-widest italic">Linked</span>
                </div>
            </div>
            <div v-if="!data.inputs || data.inputs.length === 0" class="text-[8px] text-gray-600 italic">
                Ingen variabler funnet...
            </div>
        </div>

        <!-- Static Output -->
        <div class="border-t border-gray-700 pt-2 flex justify-end items-center gap-2 relative">
            <div class="flex flex-col text-right">
                <span class="text-[7px] text-gray-500 uppercase font-black">Utdata</span>
                <span class="text-[9px] font-bold font-mono text-green-400 uppercase">result</span>
            </div>
            <Handle 
                type="source" 
                :position="Position.Right" 
                id="result"
                class="!w-4 !h-4 !bg-black !border-4 !border-green-400 !-right-6"
            />
        </div>
    </div>
  </div>
</template>
