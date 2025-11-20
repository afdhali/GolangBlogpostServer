package security

import "github.com/microcosm-cc/bluemonday"

type Sanitizer interface {
    SanitizeHTML(input string) string
    StrictSanitize(input string) string
}

type sanitizer struct {
    policy       *bluemonday.Policy
    strictPolicy *bluemonday.Policy
}

func NewSanitizer() Sanitizer {
    policy := bluemonday.UGCPolicy()
    strictPolicy := bluemonday.StrictPolicy()

    return &sanitizer{
        policy:       policy,
        strictPolicy: strictPolicy,
    }
}

func (s *sanitizer) SanitizeHTML(input string) string {
    return s.policy.Sanitize(input)
}

func (s *sanitizer) StrictSanitize(input string) string {
    return s.strictPolicy.Sanitize(input)
}