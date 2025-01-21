package middleware

import "net/http"

type ApplyMiddlewareLayer func(handler http.Handler) http.Handler

func Apply(
	target http.Handler,
	middleware ...ApplyMiddlewareLayer,
) http.Handler {

	for _, m := range middleware {
		target = m(target)
	}

	return target

}
