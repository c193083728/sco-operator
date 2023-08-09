package defaults

import "time"

const (
	SyncInterval     = 5 * time.Second
	RetryInterval    = 10 * time.Second
	ConflictInterval = 1 * time.Second

	FinalizerName = "c193083728.github.comfinalizer"
)
