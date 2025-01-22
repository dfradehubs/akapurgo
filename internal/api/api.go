package api

import (
	"akapurgo/api/v1alpha1"
	"akapurgo/internal/commons"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v9/pkg/edgegrid"
	"github.com/gofiber/fiber/v2"
	"net/http"
)

var (
	akamaiResp v1alpha1.AkamaiResponse
	req        v1alpha1.PurgeRequest
	purgeURL   string
	resp       *http.Response
)

func PurgeHandler(ctx v1alpha1.Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		// Log the request when the function ends
		defer LogRequest(c, ctx)

		// Parse the JSON body from the request and validate the body
		if err := c.BodyParser(&req); err != nil {
			ctx.Logger.Errorf("Failed to parse request: %v\n", err)
			return c.Status(fiber.StatusBadRequest).JSON(map[string]string{
				"error": "Invalid request payload",
			})
		}

		// Determine the Akamai API URL
		if req.PurgeType == "urls" {
			purgeURL = fmt.Sprintf("%s/ccu/v3/%s/url/%s", ctx.Config.Akamai.Host, req.ActionType, req.Environment)
		} else if req.PurgeType == "cache-tags" {
			purgeURL = fmt.Sprintf("%s/ccu/v3/%s/tag/%s", ctx.Config.Akamai.Host, req.ActionType, req.Environment)
		} else {
			ctx.Logger.Error("Invalid purge type")
			return c.Status(fiber.StatusBadRequest).JSON(map[string]string{
				"error": "Invalid purge type",
			})
		}

		// Create the payload for Akamai
		akamaiPayload := map[string]interface{}{
			"objects": req.Paths,
		}

		// Marshal the payload to JSON
		payloadBytes, err := json.Marshal(akamaiPayload)
		if err != nil {
			ctx.Logger.Errorf("Failed to marshal payload: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{
				"error": "Failed to encode payload",
			})
		}

		// Create the HTTP request to Akamai
		client := &http.Client{}
		apiRequest, err := http.NewRequest("POST", purgeURL, bytes.NewReader(payloadBytes))
		if err != nil {
			ctx.Logger.Errorf("Failed to create HTTP request: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{
				"error": "Failed to create request",
			})
		}

		// Generate the Authorization header with the edgerc Akamai library and the configuration file
		// generated previously or loaded from the environment
		// https://github.com/akamai/AkamaiOPEN-edgegrid-golang
		edgerc, err := edgegrid.New(edgegrid.WithFile(commons.AkamaiConfigPath))
		if err != nil {
			ctx.Logger.Errorf("Failed to sign the request with given credentials: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{
				"error": "Failed to sign the request with given credentials",
			})
		}
		edgerc.SignRequest(apiRequest)

		// Set required headers
		apiRequest.Header.Set("Content-Type", "application/json")

		// Send the request to Akamai
		resp, err = client.Do(apiRequest)
		if err != nil {
			ctx.Logger.Errorf("Failed to send request to Akamai: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{
				"error": "Failed to communicate with Akamai",
			})
		}

		defer resp.Body.Close()

		// Decode the Akamai response
		if err := json.NewDecoder(resp.Body).Decode(&akamaiResp); err != nil {
			ctx.Logger.Errorf("Failed to decode Akamai response: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{
				"error": "Failed to decode Akamai response",
			})
		}

		// Forward the Akamai response to the client
		ctx.Logger.Infof(`akamai-response,detail='%s',status=%d`, akamaiResp.Detail, akamaiResp.HTTPStatus)
		return c.Status(resp.StatusCode).JSON(akamaiResp)
	}
}
