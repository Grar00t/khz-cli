package model

// State is a deterministic evidence state.
type State string

const (
	OK   State = "OK"
	Warn State = "WARN"
	Fail State = "FAIL"
	Skip State = "SKIP"
	Info State = "INFO"
)

// Decision is a human-facing decision derived from evidence states.
type Decision string

const (
	Proceed Decision = "PROCEED"
	Review  Decision = "REVIEW"
	Stop    Decision = "STOP"
)

func Decide(states []State) Decision {
	decision := Proceed
	for _, state := range states {
		switch state {
		case Fail:
			return Stop
		case Warn, Skip:
			decision = Review
		}
	}
	return decision
}
