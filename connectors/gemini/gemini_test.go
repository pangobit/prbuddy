package gemini

import (
	"context"
	"testing"
)

func TestNewClientAllowsCustomURLWithoutRealAPIKey(t *testing.T) {
	t.Parallel()

	client, err := NewClient(context.Background(), "", "", "https://ai")
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if client.model != defaultModel {
		t.Fatalf("model = %q", client.model)
	}
}
