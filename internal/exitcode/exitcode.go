package exitcode

// Process exit codes for the khz binary. Child process codes are recorded in
// receipts and never used as KHZ usage codes.
const (
	OK    = 0
	Fail  = 1
	Usage = 2
)
