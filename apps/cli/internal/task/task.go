package task

type ResultStatus string

const (
	StatusOk       ResultStatus = "ok"
	StatusAlerting ResultStatus = "alerting"
	StatusError    ResultStatus = "error"
)

type Task struct {
	ID             string         `json:"id"`
	JobID          string         `json:"job_id"`
	JobSlug        string         `json:"job_slug"`
	Name           string         `json:"name"`
	Config         map[string]any `json:"config"`
	ScheduledFor   string         `json:"scheduled_for"`
	TimeoutSeconds int            `json:"timeout_seconds"`
	LogURL         string         `json:"log_url"`
	ResultURL      string         `json:"result_url"`
}

type Result struct {
	Status  ResultStatus   `json:"status"`
	Title   string         `json:"title"`
	Message string         `json:"message,omitempty"`
	Metrics map[string]any `json:"metrics,omitempty"`
}

func Ok(title, message string, metrics map[string]any) *Result {
	return &Result{Status: StatusOk, Title: title, Message: message, Metrics: metrics}
}

func Alerting(title, message string, metrics map[string]any) *Result {
	return &Result{Status: StatusAlerting, Title: title, Message: message, Metrics: metrics}
}

func Error(title, message string, metrics map[string]any) *Result {
	return &Result{Status: StatusError, Title: title, Message: message, Metrics: metrics}
}
