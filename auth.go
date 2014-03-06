package grunway

type AuthHandler interface {
	PerformAuth(routePtr *Route, ctx *Context) (authenticationWasSucessful bool, failureToAuthErrorNum int)
	GetSecretKey(publicKey string) (secretKey string, errNum int)
}
