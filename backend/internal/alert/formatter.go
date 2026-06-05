package alert

import (
	"fmt"

	"github.com/pingan/hydra/internal/model"
)

// FormatAlertMessage builds a human-readable alert message.
func FormatAlertMessage(rule *model.Rule, host, region string, value float64) string {
	tmpl := rule.Annotations["summary"]
	if tmpl == "" {
		tmpl = fmt.Sprintf("[%s] %s: %s is %.2f (threshold: %.2f)",
			rule.Severity, rule.Name, rule.Metric, value, rule.Threshold)
	}
	return fmt.Sprintf("%s | host=%s region=%s", tmpl, host, region)
}

// FormatResolvedMessage builds a resolution message.
func FormatResolvedMessage(rule *model.Rule, host, region string, value float64) string {
	return fmt.Sprintf("[RESOLVED] %s on %s: %s=%.2f back to normal (threshold: %.2f)",
		rule.Name, host, rule.Metric, value, rule.Threshold)
}
