package llm

type ChatRequest struct {
	Model       string
	Messages    []Message
	Temperature float64
	MaxTokens   int
}

type Message struct {
	Role    string
	Content string
}

type ChatResponse[T any] struct {
	Content string
	Raw     string
	Result  T
}

const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)
