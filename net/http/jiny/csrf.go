/**------------------------------------------------------------**
 * @filename jiny/csrf.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-24 10:11
 * @desc     go.jd100.com - jiny - csrf header validate
 **------------------------------------------------------------**/
package jiny

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"go.jd100.com/medusa/log"
)

var (
	_allowHosts = []string{
		".jianjiaolian.com",
		".jd100.com",
	}
	_allowPatterns = []string{
		`^http(?:s)?://([\w\d]+\.)?servicewechat.com/(wx7564fd5313d24844|wx618ca8c24bf06c33)`,
	}

	validations = []func(*url.URL) bool{}
)

func matchHostSuffix(suffix string) func(*url.URL) bool {
	return func(uri *url.URL) bool {
		return strings.HasSuffix(strings.ToLower(uri.Host), suffix)
	}
}

func matchPattern(pattern *regexp.Regexp) func(*url.URL) bool {
	return func(uri *url.URL) bool {
		return pattern.MatchString(strings.ToLower(uri.String()))
	}
}

// addHostSuffix add host suffix into validations
func addHostSuffix(suffix string) {
	validations = append(validations, matchHostSuffix(suffix))
}

// addPattern add referer pattern into validations
func addPattern(pattern string) {
	validations = append(validations, matchPattern(regexp.MustCompile(pattern)))
}

func init() {
	for _, r := range _allowHosts {
		addHostSuffix(r)
	}
	for _, p := range _allowPatterns {
		addPattern(p)
	}
}

func CSRF() HandlerFn {
	return func(c *Context) {
		referer := c.Request.Header.Get("Referer")
		params := c.Request.Form
		cross := (params.Get("callback") != "" && params.Get("jsonp") == "jsonp") || (params.Get("cross_domain") != "")
		if referer == "" {
			if !cross {
				return
			}
			log.Info("The request's Referer header is empty.")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		illegal := true
		if uri, err := url.Parse(referer); err == nil && uri.Host != "" {
			for _, validate := range validations {
				if validate(uri) {
					illegal = false
					break
				}
			}
		}
		if illegal {
			log.Infof("The request's Referer header `%s` does not match any of allowed referers.", referer)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
	}
}