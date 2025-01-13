package domain

/**
 * Increment
 */

type Increment struct {
	Handle          IncrementHandle
	PreviousHandle  IncrementHandle
	DocumentChanges []DocumentChange
}

func NewIncrement(
	handle IncrementHandle,
	previousHandle IncrementHandle,
	documentChanges []DocumentChange,
) Increment {
	return Increment{
		Handle:          handle,
		PreviousHandle:  previousHandle,
		DocumentChanges: documentChanges,
	}
}

/**
 * IncrementHandle
 */

type IncrementHandle struct {
	Source string
	Value  string
}

func NewIncrementHandle(source, value string) IncrementHandle {
	return IncrementHandle{
		Source: source,
		Value:  value,
	}
}
