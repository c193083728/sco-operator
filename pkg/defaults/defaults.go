package defaults

import "time"

const (
	SyncInterval     = 5 * time.Second
	RetryInterval    = 10 * time.Second
	ConflictInterval = 1 * time.Second

	FinalizerName = "sco1237896.github.com/finalizer"
)
