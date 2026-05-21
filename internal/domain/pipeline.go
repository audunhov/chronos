package domain

import (
	"fmt"
	"strings"
)

type NodeInputMode string

const (
	InputModeStatic NodeInputMode = "static"
	InputModeRef    NodeInputMode = "ref"
)

type NodeInput struct {
	Mode  NodeInputMode `json:"mode"`
	Value string        `json:"value"`
}

type PipelineNode struct {
	ID     string               `json:"id"`
	Type   string               `json:"type"`
	Inputs map[string]NodeInput `json:"inputs"`
	// UI Metadata (for Vue Flow)
	PositionX float64 `json:"pos_x"`
	PositionY float64 `json:"pos_y"`
}

type PipelineEdge struct {
	ID         string `json:"id"`
	Source     string `json:"source"`      // Node ID
	SourcePort string `json:"source_handle"` // Port ID (e.g., 'true', 'false', 'default')
	Target     string `json:"target"`      // Node ID
	TargetPort string `json:"target_handle"` // Port ID (e.g., input name)
}

type PipelineConfig struct {
	Nodes []PipelineNode `json:"nodes"`
	Edges []PipelineEdge `json:"edges"`
}

type PipelineExecutor struct {
	// Operations map node types to functions
	// An operation takes inputs and returns outputs + chosen Port + error
	Operations map[string]func(inputs map[string]any) (map[string]any, string, error)
}

func (e *PipelineExecutor) Execute(config PipelineConfig, triggerData map[string]any) error {
	// results stores the outputs of each node: map[nodeID]map[portName]value
	results := make(map[string]map[string]any)
	results["trigger"] = triggerData

	// To handle a DAG correctly, we should ideally use a topological sort.
	// But since our 'Edges' explicitly define dependencies, we can use a 
	// simple queue-based walk starting from the trigger.
	
	// Track nodes that have been executed
	executed := make(map[string]bool)
	
	// For simplicity in this implementation, we will walk the Nodes array 
	// but resolve inputs from the Edges mapping.
	// This requires that the Nodes array is ordered correctly (Topological Sort).
	// Vue Flow / Frontend should provide them in a reasonable order, 
	// or we can sort them here.
	
	for _, node := range config.Nodes {
		op, ok := e.Operations[node.Type]
		if !ok {
			return fmt.Errorf("unknown operation type: %s", node.Type)
		}

		// 1. Resolve inputs
		resolvedInputs := make(map[string]any)
		
		// 1a. Start with static values
		for key, input := range node.Inputs {
			if input.Mode == InputModeStatic {
				resolvedInputs[key] = input.Value
			}
		}

		// 1b. Override with connected edge values
		for _, edge := range config.Edges {
			if edge.Target == node.ID {
				sourceResults, ok := results[edge.Source]
				if !ok {
					// Source hasn't executed yet. This means the DAG is either 
					// invalid or out of order.
					return fmt.Errorf("dependency missing: node %s depends on %s which hasn't executed", node.ID, edge.Source)
				}
				
				val, ok := sourceResults[edge.SourcePort]
				if ok {
					resolvedInputs[edge.TargetPort] = val
				}
			}
		}

		// 2. Execute operation
		outputs, chosenPort, err := op(resolvedInputs)
		if err != nil {
			return fmt.Errorf("node %s failed: %w", node.ID, err)
		}

		// 3. Store outputs
		results[node.ID] = outputs
		executed[node.ID] = true
		
		// 4. Branching Logic: If a node returned a specific Port, 
		// we should ideally prune the graph.
		// In a complex DAG, this is done by only queuing children 
		// connected to 'chosenPort'.
		_ = chosenPort // Placeholder for future sophisticated pruning
	}

	return nil
}

// Logic Helper: Evaluate a simple condition (used by If/Then nodes)
func EvaluateCondition(left any, operator string, right any) bool {
	sLeft := fmt.Sprintf("%v", left)
	sRight := fmt.Sprintf("%v", right)

	switch operator {
	case "==":
		return sLeft == sRight
	case "!=":
		return sLeft != sRight
	case "contains":
		return strings.Contains(sLeft, sRight)
	}
	return false
}
