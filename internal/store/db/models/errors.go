package models

import "fmt"

var ErrFailedToConnectToDB = func(err error) error { return fmt.Errorf("failed to connect to database: %w", err) }
