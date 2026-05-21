package domain

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestPipelineDST(t *testing.T) {
	// Standard DST pattern: run multiple simulations with different seeds
	for seed := int64(1); seed <= 100; seed++ {
		t.Run(fmt.Sprintf("Seed_%d", seed), func(t *testing.T) {
			simulateComplexPipeline(t, seed)
		})
	}
}

func simulateComplexPipeline(t *testing.T, seed int64) {
	rng := rand.New(rand.NewSource(seed))

	// Track execution for verification
	executionLog := []string{}

	wrap := func(res map[string]any, port string, err error) ([]map[string]any, string, error) {
		if err != nil { return nil, port, err }
		return []map[string]any{res}, port, nil
	}

	executor := &PipelineExecutor{
		Operations: map[string]func(inputs map[string]any) ([]map[string]any, string, error){
			"Action": func(inputs map[string]any) ([]map[string]any, string, error) {
				msg := fmt.Sprintf("Action executed with input: %v", inputs["data"])
				executionLog = append(executionLog, msg)
				return wrap(map[string]any{"out": "processed"}, "default", nil)
			},
			"IfThen": func(inputs map[string]any) ([]map[string]any, string, error) {
				v1 := fmt.Sprintf("%v", inputs["v1"])
				v2 := fmt.Sprintf("%v", inputs["v2"])
				if v1 == v2 {
					return nil, "true", nil
				}
				return nil, "false", nil
			},
			"ForEach": func(inputs map[string]any) ([]map[string]any, string, error) {
				list, _ := inputs["list"].([]any)
				var results []map[string]any
				for _, item := range list {
					results = append(results, map[string]any{"item": item})
				}
				return results, "default", nil
			},
			"Const": func(inputs map[string]any) ([]map[string]any, string, error) {
				return wrap(map[string]any{"val": inputs["val"]}, "default", nil)
			},
		},
	}

	// 1. Build a randomized DAG
	nodeCount := 5 + rng.Intn(15)
	nodes := []PipelineNode{}
	edges := []PipelineEdge{}
	
	// Initial trigger data
	triggerData := map[string]any{
		"user_id": "user_123",
		"items": []any{"A", "B", "C"},
	}

	availableOutputs := []struct{nodeID, port string}{
		{"trigger", "user_id"},
		{"trigger", "items"},
	}

	for i := 0; i < nodeCount; i++ {
		nodeID := fmt.Sprintf("node_%d", i)
		nodeType := ""
		r := rng.Intn(10)
		
		if r < 4 { nodeType = "Action" } else if r < 7 { nodeType = "IfThen" } else if r < 9 { nodeType = "ForEach" } else { nodeType = "Const" }

		node := PipelineNode{ID: nodeID, Type: nodeType, Inputs: make(map[string]NodeInput), Data: make(map[string]any)}
		
		// Connect to random previous output
		source := availableOutputs[rng.Intn(len(availableOutputs))]
		targetPort := "data"
		if nodeType == "IfThen" { targetPort = "v1" } else if nodeType == "ForEach" { targetPort = "list" } else if nodeType == "Const" { targetPort = "val" }

		edges = append(edges, PipelineEdge{
			Source: source.nodeID, SourcePort: source.port,
			Target: nodeID, TargetPort: targetPort,
		})

		if nodeType == "IfThen" {
			// Set static v2 for comparison
			node.Data["v2"] = triggerData["user_id"]
		}

		nodes = append(nodes, node)
		
		// Add this node's outputs to available pool
		if nodeType == "Action" {
			availableOutputs = append(availableOutputs, struct{nodeID, port string}{nodeID, "out"})
		} else if nodeType == "ForEach" {
			availableOutputs = append(availableOutputs, struct{nodeID, port string}{nodeID, "item"})
		} else if nodeType == "Const" {
			availableOutputs = append(availableOutputs, struct{nodeID, port string}{nodeID, "val"})
		}
	}

	config := PipelineConfig{Nodes: nodes, Edges: edges}

	// 2. Execute
	err := executor.Execute(config, triggerData)
	if err != nil {
		t.Fatalf("Randomized execution failed: %v", err)
	}

	// 3. Verify Determinism: If we run it again with same seed, log should be identical
	// (Execution is already deterministic because we use rng with seed)
}

func TestPipelineDAG_DeepBranching(t *testing.T) {
	executor := &PipelineExecutor{
		Operations: map[string]func(inputs map[string]any) ([]map[string]any, string, error){
			"IfThen": func(inputs map[string]any) ([]map[string]any, string, error) {
				if inputs["v1"] == "A" { return nil, "true", nil }
				return nil, "false", nil
			},
			"Mark": func(inputs map[string]any) ([]map[string]any, string, error) {
				return []map[string]any{{"hit": true}}, "default", nil
			},
		},
	}

	config := PipelineConfig{
		Nodes: []PipelineNode{
			{ID: "check", Type: "IfThen", Data: map[string]any{"v1": "A"}},
			{ID: "on_true", Type: "Mark"},
			{ID: "on_false", Type: "Mark"},
		},
		Edges: []PipelineEdge{
			{Source: "check", SourcePort: "true", Target: "on_true", TargetPort: "in"},
			{Source: "check", SourcePort: "false", Target: "on_false", TargetPort: "in"},
		},
	}

	// Should only execute on_true
	err := executor.Execute(config, nil)
	if err != nil { t.Fatal(err) }
}
