package placeholder

import "errors"

// ErrPlaceholderNotFound signals that a Restore call could not locate
// the requested placeholder ID. Callers use errors.Is to distinguish
// legitimate user-deletions (placeholder removed from edited markdown)
// from real Manager-internal failures, which must be propagated.
var ErrPlaceholderNotFound = errors.New("placeholder not found")
