package provider

type IRequest interface {
	SetToken(token string)
	ShouldIgnore() bool
}
