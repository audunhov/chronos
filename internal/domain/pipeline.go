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
}

type PipelineConfig struct {
	Nodes []PipelineNode `json:"nodes"`
}

type PipelineExecutor struct {
	// Operations map node types to functions
	// An operation takes inputs and returns outputs + error
	Operations map[string]func(inputs map[string]any) (map[string]any, error)
}

func (e *PipelineExecutor) Execute(config PipelineConfig, triggerData map[string]any) error {
	results := make(map[string]map[string]any)
	results["trigger"] = triggerData

	for _, node := range config.Nodes {
		op, ok := e.Operations[node.Type]
		if !ok {
			return fmt.Errorf("unknown operation type: %s", node.Type)
		}

		// Resolve inputs
		resolvedInputs := make(map[string]any)
		for key, input := range node.Inputs {
			if input.Mode == InputModeStatic {
				resolvedInputs[key] = input.Value
			} else {
				// Resolve reference like "trigger.user_id" or "node_1.email"
				parts := strings.Split(input.Value, ".")
				if len(parts) != 2 {
					return fmt.Errorf("invalid reference format: %s", input.Value)
				}
				sourceID, field := parts[0], parts[1]
				sourceResults, ok := results[sourceID]
				if !ok {
					return fmt.Errorf("source id not found: %s", sourceID)
				}
				val, ok := sourceResults[field]
				if !ok {
					return fmt.Errorf("field %s not found in source %s", field, sourceID)
				}
				resolvedInputs[key] = val
			}
		}

		// Execute operation
		outputs, err := op(resolvedInputs)
		if err != nil {
			return fmt.Errorf("node %s failed: %w", node.ID, err)
		}

		results[node.ID] = outputs
	}

	return nil
}
