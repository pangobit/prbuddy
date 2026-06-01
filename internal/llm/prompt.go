// Package llm defines provider-neutral language model request contracts.
package llm

// Prompt is the text contract sent to an LLM provider.
type Prompt struct {
	// System is the provider-neutral instruction for model behavior.
	System string
	// User is the provider-neutral review request content.
	User string
}
