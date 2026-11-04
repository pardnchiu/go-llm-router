package llmrouter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

type DecisionType string

const (
	DecisionNoul   DecisionType = "noul"
	DecisionChoice DecisionType = "choice"
	DecisionScore  DecisionType = "score"
)

type DecisionQuestion struct {
	Type         DecisionType
	Instructions any
	Criteria     any
}

type DecisionAnswer struct {
	Type          DecisionType
	Noul          float64
	Choice        string
	Score         float64
	Confidence    *float64
	Probabilities map[string]float64
	Legend        map[string]any
}

type DecisionResult struct {
	Model   string
	Answers map[string]DecisionAnswer
	Usage   Usage
}

type DecisionAgent interface {
	Decide(ctx context.Context, state any, questions map[string]DecisionQuestion) (*DecisionResult, error)
}

func DecisionBody(model string, state any, questions map[string]DecisionQuestion) (map[string]any, error) {
	if model == "" {
		return nil, fmt.Errorf("[Invalid] decision: model is required")
	}
	if state == nil {
		return nil, fmt.Errorf("[Invalid] decision: state is required")
	}
	if len(questions) == 0 {
		return nil, fmt.Errorf("[Invalid] decision: questions is empty")
	}

	qs := make(map[string]any, len(questions))
	for name, q := range questions {
		switch q.Type {
		case DecisionNoul, DecisionChoice, DecisionScore:
		default:
			return nil, fmt.Errorf("[Invalid] decision question %q: type %q must be noul, choice or score", name, q.Type)
		}
		item := map[string]any{"type": string(q.Type)}
		if q.Instructions != nil {
			item["instructions"] = q.Instructions
		}
		if q.Criteria != nil {
			item["criteria"] = q.Criteria
		}
		qs[name] = item
	}

	return map[string]any{
		"model":     model,
		"state":     state,
		"questions": qs,
	}, nil
}

type DecisionResponse struct {
	Model   string `json:"model"`
	Answers map[string]struct {
		Type          string             `json:"type"`
		Noul          float64            `json:"noul"`
		Choice        string             `json:"choice"`
		Score         float64            `json:"score"`
		Confidence    *float64           `json:"confidence"`
		Probabilities map[string]float64 `json:"probabilities"`
		Legend        map[string]any     `json:"legend"`
	} `json:"answers"`
	Usage struct {
		InputTokens        int `json:"input_tokens"`
		OutputTokens       int `json:"output_tokens"`
		InputTokensDetails struct {
			CachedTokens int `json:"cached_tokens"`
		} `json:"input_tokens_details"`
	} `json:"usage"`
}

func (r DecisionResponse) Result(label string) (*DecisionResult, error) {
	if len(r.Answers) == 0 {
		return nil, fmt.Errorf("[Upstream] %s: response has no answers", label)
	}

	answers := make(map[string]DecisionAnswer, len(r.Answers))
	for name, a := range r.Answers {
		answer := DecisionAnswer{
			Type:          DecisionType(a.Type),
			Noul:          a.Noul,
			Choice:        a.Choice,
			Score:         a.Score,
			Confidence:    a.Confidence,
			Probabilities: a.Probabilities,
			Legend:        a.Legend,
		}
		answers[name] = answer
	}

	return &DecisionResult{
		Model:   r.Model,
		Answers: answers,
		Usage: Usage{
			Input:     r.Usage.InputTokens,
			Output:    r.Usage.OutputTokens,
			CacheRead: r.Usage.InputTokensDetails.CachedTokens,
		},
	}, nil
}

func HTTPErrorLabel(code int, err error) string {
	switch {
	case errors.Is(err, context.Canceled):
		return "Canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return "Timeout"
	}
	switch code {
	case 0:
		return "Unavailable"
	case http.StatusBadRequest, http.StatusUnsupportedMediaType, http.StatusUnprocessableEntity:
		return "Invalid"
	case http.StatusUnauthorized, http.StatusForbidden:
		return "Denied"
	case http.StatusPaymentRequired, http.StatusRequestEntityTooLarge:
		return "Limit"
	case http.StatusNotFound:
		return "Not found"
	case http.StatusConflict:
		return "Conflict"
	case http.StatusRequestTimeout, http.StatusGatewayTimeout, 524:
		return "Timeout"
	case http.StatusTooManyRequests, http.StatusServiceUnavailable, 529:
		return "Unavailable"
	}
	return "Upstream"
}
