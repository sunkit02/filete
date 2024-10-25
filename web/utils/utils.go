package utils

import (
	"fmt"

	"github.com/google/uuid"
)

func WithId(id uuid.UUID, format string, a ...any) string {
	args := make([]any, 0, len(a)+1)
	args = append(args, id)
	args = append(args, a...)

	return fmt.Sprintf("reqId(%s) "+format, args...)
}
