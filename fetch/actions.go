package fetch

import (
	"encoding/json"
	"fmt"
)

// TypedAction represents an action to be taken on a page
type TypedAction interface {
	GetType() string
}

// BaseAction contains common fields for all actions
type BaseAction struct {
	Type string `json:"type"`
}

func (a BaseAction) GetType() string {
	return a.Type
}

// ScreenshotAction triggers a screenshot of the page
type ScreenshotAction struct {
	BaseAction
	FullPage bool `json:"full_page,omitempty"`
}

// PDFAction generates a PDF of the page
type PDFAction struct {
	BaseAction
	Format string `json:"format,omitempty"` // A4, Letter, Legal, etc.
}

// WaitAction waits for a condition or time
type WaitAction struct {
	BaseAction
	Selector string `json:"selector,omitempty"` // Wait for element to appear
	Duration int    `json:"duration,omitempty"` // Wait for specific duration in milliseconds
}

// Action is used for JSON marshaling/unmarshaling of polymorphic actions
type Action struct {
	Action TypedAction
}

// UnmarshalJSON implements custom unmarshaling for polymorphic actions
func (a *Action) UnmarshalJSON(data []byte) error {
	// Unmarshal just the type field
	var typeOnly struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &typeOnly); err != nil {
		return err
	}
	switch typeOnly.Type {
	case "screenshot":
		var action ScreenshotAction
		action.Type = typeOnly.Type
		if err := json.Unmarshal(data, &action); err != nil {
			return err
		}
		a.Action = &action
	case "pdf":
		var action PDFAction
		action.Type = typeOnly.Type
		if err := json.Unmarshal(data, &action); err != nil {
			return err
		}
		a.Action = &action
	case "wait":
		var action WaitAction
		action.Type = typeOnly.Type
		if err := json.Unmarshal(data, &action); err != nil {
			return err
		}
		a.Action = &action
	default:
		return fmt.Errorf("unknown action type: %s", typeOnly.Type)
	}
	return nil
}

// MarshalJSON implements custom marshaling for polymorphic actions
func (a *Action) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Action)
}

// ScreenshotActionOptions represents the options for a screenshot action
type ScreenshotActionOptions struct {
	FullPage bool `json:"full_page,omitempty"`
}

// NewScreenshotAction creates a new screenshot action
func NewScreenshotAction(options ScreenshotActionOptions) Action {
	return Action{
		Action: &ScreenshotAction{
			BaseAction: BaseAction{Type: "screenshot"},
			FullPage:   options.FullPage,
		},
	}
}

// PDFActionOptions represents the options for a PDF action
type PDFActionOptions struct {
	Format string `json:"format,omitempty"` // A4, Letter, Legal, etc.
}

// NewPDFAction creates a new PDF action
func NewPDFAction(options PDFActionOptions) Action {
	return Action{
		Action: &PDFAction{
			BaseAction: BaseAction{Type: "pdf"},
			Format:     options.Format,
		},
	}
}

// WaitActionOptions represents the options for a wait action
type WaitActionOptions struct {
	Selector string `json:"selector,omitempty"` // Wait for element to appear
	Duration int    `json:"duration,omitempty"` // Wait for specific duration in milliseconds
}

// NewWaitAction creates a new wait action
func NewWaitAction(options WaitActionOptions) Action {
	return Action{
		Action: &WaitAction{
			BaseAction: BaseAction{Type: "wait"},
			Selector:   options.Selector,
			Duration:   options.Duration,
		},
	}
}
