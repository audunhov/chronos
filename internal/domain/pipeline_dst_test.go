package domain

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestPipelineDST(t *testing.T) {
	// Standard DST pattern: run multiple simulations with different seeds
	for seed := int64(0); seed < 100; seed++ {
		t.Run(fmt.Sprintf("Seed_%d", seed), func(t *testing.T) {
			simulatePipeline(t, seed)
		})
	}
}

func simulatePipeline(t *testing.T, seed int64) {
	rng := rand.New(rand.NewSource(seed))

	// 1. Setup Mock Operations
	executor := &PipelineExecutor{
		Operations: map[string]func(inputs map[string]any) (map[string]any, error){
			"FindOrg": func(inputs map[string]any) (map[string]any, error) {
				orgID := inputs["start_org_id"].(string)
				return map[string]any{
					"org_id":   orgID,
					"org_name": "Mock Org " + orgID,
				}, nil
			},
			"FindRole": func(inputs map[string]any) (map[string]any, error) {
				role := inputs["role_type"].(string)
				return map[string]any{
					"user_id": "u_" + role,
					"name":    "Leader Name",
					"email":   "leader@example.com",
				}, nil
			},
			"Template": func(inputs map[string]any) (map[string]any, error) {
				return map[string]any{
					"subject": "Velkommen",
					"body":    fmt.Sprintf("Hei %s, %s har meldt seg inn.", inputs["leader_name"], inputs["user_name"]),
				}, nil
			},
			"SendEmail": func(inputs map[string]any) (map[string]any, error) {
				// Simulate internal state change or external call
				return map[string]any{"sent": true}, nil
			},
		},
	}

	// 2. Define complex Pipeline (DAG)
	config := PipelineConfig{
		Nodes: []PipelineNode{
			{
				ID:   "node_1",
				Type: "FindOrg",
				Inputs: map[string]NodeInput{
					"start_org_id": {Mode: InputModeRef, Value: "trigger.org_id"},
				},
			},
			{
				ID:   "node_2",
				Type: "FindRole",
				Inputs: map[string]NodeInput{
					"target_id": {Mode: InputModeRef, Value: "node_1.org_id"},
					"role_type": {Mode: InputModeStatic, Value: "leader"},
				},
			},
			{
				ID:   "node_3",
				Type: "Template",
				Inputs: map[string]NodeInput{
					"user_name":   {Mode: InputModeRef, Value: "trigger.user_name"},
					"leader_name": {Mode: InputModeRef, Value: "node_2.name"},
				},
			},
			{
				ID:   "node_4",
				Type: "SendEmail",
				Inputs: map[string]NodeInput{
					"to_email": {Mode: InputModeRef, Value: "node_2.email"},
					"subject":  {Mode: InputModeRef, Value: "node_3.subject"},
					"body":     {Mode: InputModeRef, Value: "node_3.body"},
				},
			},
		},
	}

	// 3. Simulation inputs
	triggerData := map[string]any{
		"user_id":   fmt.Sprintf("user_%d", rng.Intn(1000)),
		"user_name": "Nymedlem",
		"org_id":    "org_123",
	}

	// 4. Execute
	err := executor.Execute(config, triggerData)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// 5. Verify determinism (Optional, but here we check basic logic)
	// In a full DST we might check a shared state log
}
