package domain

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestPipelineDST(t *testing.T) {
	// Standard DST pattern: run multiple simulations with different seeds
	for seed := int64(1); seed <= 50; seed++ {
		t.Run(fmt.Sprintf("Seed_%d", seed), func(t *testing.T) {
			simulateRandomPipeline(t, seed)
		})
	}
}

func simulateRandomPipeline(t *testing.T, seed int64) {
	rng := rand.New(rand.NewSource(seed))

	// 1. Setup Mock Operations with predictable behaviors
	executor := &PipelineExecutor{
		Operations: map[string]func(inputs map[string]any) (map[string]any, string, error){
			"Concat": func(inputs map[string]any) (map[string]any, string, error) {
				v1, _ := inputs["v1"].(string)
				v2, _ := inputs["v2"].(string)
				return map[string]any{"out": v1 + v2}, "default", nil
			},
			"Identity": func(inputs map[string]any) (map[string]any, string, error) {
				return map[string]any{"out": inputs["in"]}, "default", nil
			},
			"Const": func(inputs map[string]any) (map[string]any, string, error) {
				return map[string]any{"out": inputs["val"]}, "default", nil
			},
			"IfThen": func(inputs map[string]any) (map[string]any, string, error) {
				v1 := inputs["v1"]
				v2 := inputs["v2"]
				if v1 == v2 {
					return nil, "true", nil
				}
				return nil, "false", nil
			},
		},
	}

	// 2. Generate Random Pipeline (DAG)
	nodeCount := 2 + rng.Intn(10)
	nodes := make([]PipelineNode, 0)
	edges := make([]PipelineEdge, 0)
	
	triggerData := map[string]any{
		"init": fmt.Sprintf("trigger_%d", seed),
	}

	availableOutputs := []struct{nodeID, port string}{
		{"trigger", "init"},
	}

	for i := 0; i < nodeCount; i++ {
		nodeID := fmt.Sprintf("node_%d", i)
		
		r := rng.Intn(4)
		switch r {
		case 0: // Const
			nodes = append(nodes, PipelineNode{
				ID: nodeID, Type: "Const",
				Inputs: map[string]NodeInput{"val": {Mode: InputModeStatic, Value: "fixed"}},
			})
			availableOutputs = append(availableOutputs, struct{nodeID, port string}{nodeID, "out"})
		case 1: // Identity
			source := availableOutputs[rng.Intn(len(availableOutputs))]
			nodes = append(nodes, PipelineNode{
				ID: nodeID, Type: "Identity",
			})
			edges = append(edges, PipelineEdge{
				Source: source.nodeID, SourcePort: source.port,
				Target: nodeID, TargetPort: "in",
			})
			availableOutputs = append(availableOutputs, struct{nodeID, port string}{nodeID, "out"})
		case 2: // Concat
			s1 := availableOutputs[rng.Intn(len(availableOutputs))]
			s2 := availableOutputs[rng.Intn(len(availableOutputs))]
			nodes = append(nodes, PipelineNode{
				ID: nodeID, Type: "Concat",
			})
			edges = append(edges, PipelineEdge{
				Source: s1.nodeID, SourcePort: s1.port,
				Target: nodeID, TargetPort: "v1",
			})
			edges = append(edges, PipelineEdge{
				Source: s2.nodeID, SourcePort: s2.port,
				Target: nodeID, TargetPort: "v2",
			})
			availableOutputs = append(availableOutputs, struct{nodeID, port string}{nodeID, "out"})
		case 3: // IfThen
			s1 := availableOutputs[rng.Intn(len(availableOutputs))]
			s2 := availableOutputs[rng.Intn(len(availableOutputs))]
			nodes = append(nodes, PipelineNode{
				ID: nodeID, Type: "IfThen",
			})
			edges = append(edges, PipelineEdge{
				Source: s1.nodeID, SourcePort: s1.port,
				Target: nodeID, TargetPort: "v1",
			})
			edges = append(edges, PipelineEdge{
				Source: s2.nodeID, SourcePort: s2.port,
				Target: nodeID, TargetPort: "v2",
			})
			// IfThen has no outputs, but has true/false ports (not yet used for execution path pruning in engine)
		}
	}

	config := PipelineConfig{Nodes: nodes, Edges: edges}

	// 4. Run
	err := executor.Execute(config, triggerData)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}
}

func TestPipelineDAG_BranchingLogic(t *testing.T) {
	executor := &PipelineExecutor{
		Operations: map[string]func(inputs map[string]any) (map[string]any, string, error){
			"IfThen": func(inputs map[string]any) (map[string]any, string, error) {
				if inputs["v1"] == inputs["v2"] {
					return map[string]any{"res": "match"}, "true", nil
				}
				return map[string]any{"res": "no-match"}, "false", nil
			},
			"Collector": func(inputs map[string]any) (map[string]any, string, error) {
				return inputs, "default", nil
			},
		},
	}

	config := PipelineConfig{
		Nodes: []PipelineNode{
			{ID: "check", Type: "IfThen", Inputs: map[string]NodeInput{
				"v1": {Mode: InputModeStatic, Value: "A"},
				"v2": {Mode: InputModeStatic, Value: "A"},
			}},
			{ID: "result", Type: "Collector"},
		},
		Edges: []PipelineEdge{
			{Source: "check", SourcePort: "res", Target: "result", TargetPort: "data"},
		},
	}

	err := executor.Execute(config, nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}
