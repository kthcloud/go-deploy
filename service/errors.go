package service

type QuotaExceededError struct {
	reason string
}

func (e QuotaExceededError) Error() string {
	return e.reason
}

func NewQuotaExceededError(reason string) QuotaExceededError {
	return QuotaExceededError{reason: reason}
}
