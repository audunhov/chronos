package domain

import (
	"fmt"
	"math/rand"
	"reflect"
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
		Operations: map[string]func(inputs map[string]any) (map[string]any, error){
			"Concat": func(inputs map[string]any) (map[string]any, error) {
				v1, _ := inputs["v1"].(string)
				v2, _ := inputs["v2"].(string)
				return map[string]any{"out": v1 + v2}, nil
			},
			"Identity": func(inputs map[string]any) (map[string]any, error) {
				return map[string]any{"out": inputs["in"]}, nil
			},
			"Const": func(inputs map[string]any) (map[string]any, error) {
				return map[string]any{"out": inputs["val"]}, nil
			},
		},
	}

	// 2. Generate Random Pipeline (DAG)
	nodeCount := 2 + rng.Intn(10) // 2 to 12 nodes
	nodes := make([]PipelineNode, nodeCount)
	
	// Track expected results for each node
	expectedResults := make(map[string]map[string]any)
	triggerData := map[string]any{
		"init": fmt.Sprintf("trigger_%d", seed),
	}
	expectedResults["trigger"] = triggerData

	availableSources := []string{"trigger"}

	for i := 0; i < nodeCount; i++ {
		nodeID := fmt.Sprintf("node_%d", i)
		nodeType := ""
		inputs := make(map[string]NodeInput)
		expected := make(map[string]any)

		// Choose a random operation
		r := rng.Intn(3)
		switch r {
		case 0: // Const
			nodeType = "Const"
			val := fmt.Sprintf("const_%d", rng.Intn(100))
			inputs["val"] = NodeInput{Mode: InputModeStatic, Value: val}
			expected["out"] = val
		case 1: // Identity
			nodeType = "Identity"
			source := availableSources[rng.Intn(len(availableSources))]
			// We know all our mocks output "out" or use trigger.init
			field := "out"
			if source == "trigger" { field = "init" }
			
			inputs["in"] = NodeInput{Mode: InputModeRef, Value: source + "." + field}
			expected["out"] = expectedResults[source][field]
		case 2: // Concat
			nodeType = "Concat"
			s1 := availableSources[rng.Intn(len(availableSources))]
			s2 := availableSources[rng.Intn(len(availableSources))]
			
			f1 := "out"; if s1 == "trigger" { f1 = "init" }
			f2 := "out"; if s2 == "trigger" { f2 = "init" }

			inputs["v1"] = NodeInput{Mode: InputModeRef, Value: s1 + "." + f1}
			inputs["v2"] = NodeInput{Mode: InputModeRef, Value: s2 + "." + f2}
			
			v1 := expectedResults[s1][f1].(string)
			v2 := expectedResults[s2][f2].(string)
			expected["out"] = v1 + v2
		}

		nodes[i] = PipelineNode{
			ID:     nodeID,
			Type:   nodeType,
			Inputs: inputs,
		}
		expectedResults[nodeID] = expected
		availableSources = append(availableSources, nodeID)
	}

	config := PipelineConfig{Nodes: nodes}

	// 3. Execution Engine Capture
	// We need a way to verify internal state after execution.
	// Let's modify Execute to return the full results map for testing, or just rely on side effects.
	// Since Execute currently returns error, let's use a "Collector" operation to verify.
	
	finalNodeID := "final_verifier"
	collectorResults := make(map[string]any)
	executor.Operations["Collector"] = func(inputs map[string]any) (map[string]any, error) {
		for k, v := range inputs {
			collectorResults[k] = v
		}
		return nil, nil
	}

	collectorInputs := make(map[string]NodeInput)
	expectedCollector := make(map[string]any)
	for _, source := range availableSources {
		if source == "trigger" { continue }
		field := "out"
		collectorInputs[source] = NodeInput{Mode: InputModeRef, Value: source + "." + field}
		expectedCollector[source] = expectedResults[source][field]
	}

	config.Nodes = append(config.Nodes, PipelineNode{
		ID:     finalNodeID,
		Type:   collectorResultsOpType(), // helper
		Inputs: collectorInputs,
	})

	// Use a closure or wrapper to access collectorResults in the operation
	executor.Operations["Collector"] = func(inputs map[string]any) (map[string]any, error) {
		for k, v := range inputs {
			collectorResults[k] = v
		}
		return nil, nil
	}

	// 4. Run
	err := executor.Execute(config, triggerData)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// 5. Verify entire graph state
	if !reflect.DeepEqual(collectorResults, expectedCollector) {
		t.Errorf("Simulation results mismatch!\nGot: %v\nExp: %v", collectorResults, expectedCollector)
	}
}

func collectorResultsOpType() string { return "Collector" }
