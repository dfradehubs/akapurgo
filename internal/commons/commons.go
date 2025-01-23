package commons

import (
	"akapurgo/api/v1alpha1"
	"encoding/base64"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	AkamaiConfigPath = "/tmp/.edgerc"

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
func GetRequestLogFields(req *fasthttp.Request, configurationFields []string, ctx v1alpha1.Context) []interface{} {
	var logFields []interface{}

	if ctx.Config.Logs.JwtUser.Enabled {
		logFields = addJwtUser(ctx, logFields, req)
	}

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
func GetResponseLogFields(resp *fasthttp.Response, configurationFields []string, duration time.Duration) []interface{} {
	var logFields []interface{}

	logFields = append(logFields, "duration", duration)

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

// addJwtUser
func addJwtUser(ctx v1alpha1.Context, logFields []interface{}, req *fasthttp.Request) []interface{} {
	cookie := string(req.Header.Peek(ctx.Config.Logs.JwtUser.Header))
	if cookie != "" {
		jwtPayload := strings.Split(cookie, ".")
		if len(jwtPayload) != 3 {
			ctx.Logger.Errorf("Invalid JWT format: expected 3 parts but got %d\n", len(jwtPayload))
			return logFields
		}

		jwtPart := strings.TrimSpace(jwtPayload[1])
		jwtPart = strings.ReplaceAll(jwtPart, "\n", "")
		jwtPart = strings.ReplaceAll(jwtPart, "\r", "")
		jwtPart = strings.ReplaceAll(jwtPart, " ", "")

		jwtDecoded, err := base64.RawURLEncoding.DecodeString(jwtPart)
		if err != nil {
			ctx.Logger.Errorf("Failed to decode JWT payload: %v\n", err)
			return logFields
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(jwtDecoded, &payload); err != nil {
			ctx.Logger.Errorf("Failed to parse JWT payload: %v\n", err)
			return logFields
		}

		user, ok := payload[ctx.Config.Logs.JwtUser.JwtField].(string)
		if ok {
			logFields = append(logFields, "jwt_user", user)
			return logFields
		}
	}

	return logFields
}

// LogRequest logs the request and response of a given HTTP request
func LogRequest(ctx v1alpha1.Context) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get initial time
		start := time.Now()

		// Process the request
		err := c.Next()

		// Get the duration of the request
		duration := time.Since(start)

		// Log the request
		if ctx.Config.Logs.ShowAccessLogs {
			logFieldsReq := GetResponseLogFields(c.Response(), ctx.Config.Logs.AccessLogsFields, duration)
			logFieldsResp := GetRequestLogFields(c.Request(), ctx.Config.Logs.AccessLogsFields, ctx)
			logFields := append(logFieldsReq, logFieldsResp...)
			ctx.Logger.Infow("request", logFields...)
		}

		return err
	}
}
