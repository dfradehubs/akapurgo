package api

import (
	"akapurgo/api/v1alpha1"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"regexp"
	"strconv"
	"strings"
)

const (
	RequestPartsPattern   = `REQUEST:([^\}]+)`
	RequestHeaderPattern  = `REQUEST_HEADER:([^\}]+)`
	ResponsePartsPattern  = `RESPONSE:([^\}]+)`
	ResponseHeaderPattern = `RESPONSE_HEADER:([^\}]+)`
)

var (
	//
	RequestPartsPatternCompiled    = regexp.MustCompile(RequestPartsPattern)
	RequestHeadersPatternCompiled  = regexp.MustCompile(RequestHeaderPattern)
	ResponsePartsPatternCompiled   = regexp.MustCompile(ResponsePartsPattern)
	ResponseHeadersPatternCompiled = regexp.MustCompile(ResponseHeaderPattern)
)

// replaceRequestTags replaces the HTTP request tags in the given text
// Tags are expressed as ${REQUEST:<part>}, where <part> can be one of the following:
// scheme, host, port, path, query, method, proto
func replaceRequestTags(req *fasthttp.Request, textToProcess string) (result string) {

	// Replace request parts in the format ${REQUEST:<part>}
	requestTags := map[string]string{
		"scheme":  string(req.URI().Scheme()),
		"host":    string(req.Host()),
		"path":    string(req.URI().Path()),
		"query":   string(req.URI().QueryString()),
		"method":  string(req.Header.Method()),
		"proto":   string(req.Header.Protocol()),
		"referer": string(req.Header.Referer()),
		"body":    string(req.Body()),
	}

	result = RequestPartsPatternCompiled.ReplaceAllStringFunc(textToProcess, func(match string) string {

		variable := strings.ToLower(RequestPartsPatternCompiled.FindStringSubmatch(match)[1])

		if replacement, exists := requestTags[variable]; exists {
			return replacement
		}

		return ""
	})

	return result
}

// replaceRequestHeaderTags replaces the HTTP request headers in the given text
// Tags are expressed as ${REQUEST_HEADER:<header-name>}
func replaceRequestHeaderTags(req *fasthttp.Request, textToProcess string) (result string) {

	result = RequestHeadersPatternCompiled.ReplaceAllStringFunc(textToProcess, func(match string) string {

		variable := strings.ToLower(RequestHeadersPatternCompiled.FindStringSubmatch(match)[1])
		headerValue := string(req.Header.Peek(variable))

		return headerValue
	})

	return result
}

// GetRequestLogFields returns the fields attached to a log message for the given HTTP request
func GetRequestLogFields(req *fasthttp.Request, configurationFields []string) []interface{} {
	var logFields []interface{}

	for _, field := range configurationFields {

		result := replaceRequestTags(req, field)

		result = replaceRequestHeaderTags(req, result)

		// Ignore not expanded fields
		if result == field {
			continue
		}

		// Clean the field name a bit and add it to the fields pool
		field = strings.TrimPrefix(field, "REQUEST:")
		field = strings.TrimPrefix(field, "REQUEST_HEADER:")

		logFields = append(logFields, field, result)
	}

	return logFields
}

// replaceResponseTags replaces the HTTP request tags in the given text
// Tags are expressed as ${REQUEST:<part>}, where <part> can be one of the following:
// scheme, host, port, path, query, method, proto
func replaceResponseTags(resp *fasthttp.Response, textToProcess string) (result string) {

	// Replace request parts in the format RESPONSE:<part>
	responseTags := map[string]string{
		"status": strconv.Itoa(resp.StatusCode()),
		"body":   string(resp.Body()),
		"proto":  string(resp.Header.Protocol()),
	}

	result = ResponsePartsPatternCompiled.ReplaceAllStringFunc(textToProcess, func(match string) string {

		variable := strings.ToLower(ResponsePartsPatternCompiled.FindStringSubmatch(match)[1])

		if replacement, exists := responseTags[variable]; exists {
			return replacement
		}

		return ""
	})

	return result
}

// replaceResponseHeaderTags replaces the HTTP response headers in the given text
// Tags are expressed as ${RESPONSE_HEADER:<header-name>}
func replaceResponseHeaderTags(res *fasthttp.Response, textToProcess string) (result string) {

	result = ResponseHeadersPatternCompiled.ReplaceAllStringFunc(textToProcess, func(match string) string {

		variable := strings.ToLower(ResponseHeadersPatternCompiled.FindStringSubmatch(match)[1])
		headerValue := string(res.Header.Peek(variable))

		return headerValue
	})

	return result
}

// GetResponseLogFields returns the fields attached to a log message for the given HTTP response
func GetResponseLogFields(resp *fasthttp.Response, configurationFields []string) []interface{} {
	var logFields []interface{}

	for _, field := range configurationFields {

		result := replaceResponseTags(resp, field)

		result = replaceResponseHeaderTags(resp, result)

		// Ignore not expanded fields
		if result == field {
			continue
		}

		// Clean the field name a bit and add it to the fields pool
		field = strings.TrimPrefix(field, "RESPONSE:")
		field = strings.TrimPrefix(field, "RESPONSE_HEADER:")

		logFields = append(logFields, field, result)
	}

	return logFields
}

// LogRequest
func LogRequest(c *fiber.Ctx, ctx v1alpha1.Context) {
	if ctx.Config.Logs.ShowAccessLogs {
		logFieldsReq := GetResponseLogFields(c.Response(), ctx.Config.Logs.AccessLogsFields)
		logFieldsResp := GetRequestLogFields(c.Request(), ctx.Config.Logs.AccessLogsFields)
		logFields := append(logFieldsReq, logFieldsResp...)
		ctx.Logger.Infow("request", logFields...)
	}
}
