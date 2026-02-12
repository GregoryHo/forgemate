package reload

import "testing"

func TestDecideActionHotApplyByDefault(t *testing.T) {
	action := DecideAction([]string{"providers.openai.apiKey", "memory.embeddingModel"})
	if action != ActionHotApply {
		t.Fatalf("expected %q, got %q", ActionHotApply, action)
	}
}

func TestDecideActionRestartForCriticalKeys(t *testing.T) {
	action := DecideAction([]string{"memory.embeddingModel", "sidecar.socket"})
	if action != ActionRestart {
		t.Fatalf("expected %q, got %q", ActionRestart, action)
	}
}

func TestDecideActionEmptyKeysHotApply(t *testing.T) {
	action := DecideAction(nil)
	if action != ActionHotApply {
		t.Fatalf("expected %q, got %q", ActionHotApply, action)
	}
}
