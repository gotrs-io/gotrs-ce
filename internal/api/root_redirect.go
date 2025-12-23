package api

import (
	"os"
	"strings"
)

func resolveRootRedirect() string {
	v := strings.TrimSpace(os.Getenv("ROOT_REDIRECT_PATH"))
	if v == "" {
		flag := strings.ToLower(strings.TrimSpace(os.Getenv("CUSTOMER_FE_ONLY")))
		if flag == "1" || flag == "true" {
			return "/customer"
		}
	}
	if v == "" {
		return "/login"
	}
	if !strings.HasPrefix(v, "/") {
		return "/" + v
	}
	return v
}

func RootRedirectTarget() string {
	return resolveRootRedirect()
}
