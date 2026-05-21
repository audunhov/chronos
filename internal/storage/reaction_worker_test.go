package storage

import (
	"testing"
)

func TestReactionWorker_OpCode(t *testing.T) {
	worker := &ReactionWorker{}
	
	t.Run("Basic JavaScript", func(t *testing.T) {
		inputs := map[string]any{
			"code": "const result = a + b; result;",
			"a":    10,
			"b":    20,
		}
		
		res, port, err := worker.opCode(inputs)
		if err != nil {
			t.Fatalf("opCode failed: %v", err)
		}
		
		if port != "default" {
			t.Errorf("expected port default, got %s", port)
		}
		
		if res["result"] != int64(30) {
			t.Errorf("expected result 30, got %v", res["result"])
		}
	})

	t.Run("Logging", func(t *testing.T) {
		inputs := map[string]any{
			"code": "log('hello world'); 42;",
		}
		
		res, _, _ := worker.opCode(inputs)
		logs, ok := res["_logs"].([]string)
		if !ok || len(logs) == 0 || logs[0] != "hello world" {
			t.Errorf("expected log 'hello world', got %v", res["_logs"])
		}
	})

	t.Run("Variable Export", func(t *testing.T) {
		inputs := map[string]any{
			"code": "var x = 100; var y = 'test';",
		}
		
		res, _, _ := worker.opCode(inputs)
		if res["x"] != int64(100) || res["y"] != "test" {
			t.Errorf("expected exported variables, got %v", res)
		}
	})
}
