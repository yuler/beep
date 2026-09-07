package probe

type Status string

const (
	StatusOk       Status = "ok"
	StatusAlerting Status = "alerting"
	StatusError    Status = "error"
)

type Signal struct {
	Status  Status         `json:"status"`
	Title   string         `json:"title"`
	Message string         `json:"message,omitempty"`
	Metrics map[string]any `json:"metrics,omitempty"`
}

func OkSignal(title string, message string, metrics map[string]any) *Signal {
	return &Signal{
		Status:  StatusOk,
		Title:   title,
		Message: message,
		Metrics: metrics,
	}
}

func AlertingSignal(title string, message string, metrics map[string]any) *Signal {
	return &Signal{
		Status:  StatusAlerting,
		Title:   title,
		Message: message,
		Metrics: metrics,
	}
}

func ErrorSignal(title string, message string, metrics map[string]any) *Signal {
	return &Signal{
		Status:  StatusError,
		Title:   title,
		Message: message,
		Metrics: metrics,
	}
}
