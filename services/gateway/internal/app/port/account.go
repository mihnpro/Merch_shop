package port

type Verifier interface {
	VerifyAccess(token string) error
}
