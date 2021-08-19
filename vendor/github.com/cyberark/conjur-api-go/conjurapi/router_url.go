package conjurapi

import (
	"fmt"
	"strings"
)

type routerURL string

func makeRouterURL(components ...string) routerURL {
	return routerURL(strings.Join(components, "/"))
}

func (url routerURL) withFormattedQuery(queryFormat string, queryArgs ...interface{}) routerURL {
	query := fmt.Sprintf(queryFormat, queryArgs...)
	return routerURL(strings.Join([]string{string(url), query}, "?"))
}

func (url routerURL) withQuery(query string) routerURL {
	return routerURL(strings.Join([]string{string(url), query}, "?"))
}

func (url routerURL) String() string {
	return string(url)
}
