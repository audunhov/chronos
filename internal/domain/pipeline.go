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
	ID       string               `json:"id"`
	Type     string               `json:"type"`
	Inputs   map[string]NodeInput `json:"inputs"`
	Data     map[string]any       `json:"data"` // For custom node state
	Position struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"position"`
}

type PipelineEdge struct {
	ID         string `json:"id"`
	Source     string `json:"source"`        // Node ID
	SourcePort string `json:"sourceHandle"`  // Port ID (e.g., 'true', 'false', 'default')
	Target     string `json:"target"`        // Node ID
	TargetPort string `json:"targetHandle"`  // Port ID (e.g., input name)
}

type PipelineConfig struct {
	Nodes []PipelineNode `json:"nodes"`
	Edges []PipelineEdge `json:"edges"`
}

type PipelineExecutor struct {
	// Operations map node types to functions
	// An operation takes inputs and returns one or more output maps + chosen Port + error
	Operations map[string]func(inputs map[string]any) ([]map[string]any, string, error)
}

func (e *PipelineExecutor) Execute(config PipelineConfig, triggerData map[string]any) ([]map[string]any, error) {
	results := make(map[string]map[string]any)
	results["trigger"] = triggerData

	var trace []map[string]any

	type ExecutionTask struct {
		NodeID      string
		InputValues map[string]any
	}

	queue := []ExecutionTask{}

	for _, edge := range config.Edges {
		if edge.Source == "trigger" {
			queue = append(queue, ExecutionTask{NodeID: edge.Target, InputValues: triggerData})
		}
	}

	maxExecutions := 2000
	executionCount := 0

	for len(queue) > 0 {
		if executionCount > maxExecutions {
			err := fmt.Errorf("pipeline exceeded max executions (%d)", maxExecutions)
			trace = append(trace, map[string]any{"error": err.Error()})
			return trace, err
		}
		executionCount++

		task := queue[0]
		queue = queue[1:]

		var node *PipelineNode
		for i := range config.Nodes {
			if config.Nodes[i].ID == task.NodeID {
				node = &config.Nodes[i]
				break
			}
		}
		if node == nil { continue }

		op, ok := e.Operations[node.Type]
		if !ok { 
			err := fmt.Errorf("unknown operation: %s", node.Type)
			trace = append(trace, map[string]any{"node_id": task.NodeID, "error": err.Error()})
			return trace, err 
		}

		// Resolve inputs
		resolvedInputs := make(map[string]any)
		for key, input := range node.Inputs {
			if input.Mode == InputModeStatic {
				resolvedInputs[key] = input.Value
			}
		}
		for key, val := range node.Data {
			if _, ok := resolvedInputs[key]; !ok {
				resolvedInputs[key] = val
			}
		}
		// Special: merge task input values (allows ForEach to pass 'item')
		for k, v := range task.InputValues {
			resolvedInputs[k] = v
		}

		for _, edge := range config.Edges {
			if edge.Target == node.ID {
				sourceResults, ok := results[edge.Source]
				if ok {
					val, ok := sourceResults[edge.SourcePort]
					if ok { resolvedInputs[edge.TargetPort] = val }
				}
			}
		}

		// Execute
		outputList, chosenPort, err := op(resolvedInputs)
		
		traceEntry := map[string]any{
			"node_id": task.NodeID,
			"inputs": resolvedInputs,
		}

		if err != nil {
			traceEntry["error"] = err.Error()
			trace = append(trace, traceEntry)
			if err.Error() == "filtered" { continue }
			return trace, fmt.Errorf("node %s failed: %w", node.ID, err)
		}

		traceEntry["outputs"] = outputList
		traceEntry["port"] = chosenPort
		trace = append(trace, traceEntry)

		// Store last result for referencing
		if len(outputList) > 0 {
			results[node.ID] = outputList[len(outputList)-1]
		}

		// Queue next nodes
		for _, outputs := range outputList {
			for _, edge := range config.Edges {
				if edge.Source == node.ID {
					if chosenPort == "default" || edge.SourcePort == chosenPort {
						queue = append(queue, ExecutionTask{
							NodeID:      edge.Target,
							InputValues: outputs,
						})
					}
				}
			}
		}
	}

	return trace, nil
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
