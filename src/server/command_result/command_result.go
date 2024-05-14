package command_result

type CommandResult struct {
	Error         string `json:"error,omitempty"`
	PID           int    `json:"pid,omitempty"`
	ID            int64  `json:"id,omitempty"`
	Output        string `json:"output,omitempty"`
	CommandErrors string `json:"command_errors,omitempty"`
}
