package reload

// Action describes how a changed config key should be applied.
type Action string

const (
	ActionHotApply Action = "hot-apply"
	ActionRestart  Action = "restart"
)

// DecideAction applies the MVP hybrid policy: safe keys hot-apply, critical keys restart.
func DecideAction(changedKeys []string) Action {
	for _, key := range changedKeys {
		switch key {
		case "gateway.auth", "gateway.port", "sidecar.socket", "channels.telegram.mode":
			return ActionRestart
		}
	}
	return ActionHotApply
}
