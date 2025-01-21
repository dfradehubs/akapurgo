package run

import (
	"akapurgo/api/v1alpha1"
	"akapurgo/internal/api"
	"akapurgo/internal/config"
	"akapurgo/internal/globals"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

const (
	descriptionShort = `Run akapurgo webserver`
	descriptionLong  = `
	Run akapurgo webserver`

	//
	ConfigFlagErrorMessage       = "impossible to get flag --config: %s"
	ConfigNotParsedErrorMessage  = "impossible to parse config file: %s"
	LogLevelFlagErrorMessage     = "impossible to get flag --log-level: %s"
	DisableTraceFlagErrorMessage = "impossible to get flag --disable-trace: %s"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "run",
		DisableFlagsInUseLine: true,
		Short:                 descriptionShort,
		Long:                  strings.ReplaceAll(descriptionLong, "\t", ""),

		Run: RunCommand,
	}

	cmd.Flags().String("config", "config.yaml", "Path to the YAML config file")
	cmd.Flags().String("log-level", "info", "Verbosity level for logs")
	cmd.Flags().Bool("disable-trace", true, "Disable showing traces in logs")

	return cmd
}

func RunCommand(cmd *cobra.Command, args []string) {

	// Init the logger and store the level into the context
	logLevelFlag, err := cmd.Flags().GetString("log-level")
	if err != nil {
		log.Fatalf(LogLevelFlagErrorMessage, err)
	}

	// Check the flags for this command
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		log.Fatalf(ConfigFlagErrorMessage, err)
	}

	disableTraceFlag, err := cmd.Flags().GetBool("disable-trace")
	if err != nil {
		log.Fatalf(DisableTraceFlagErrorMessage, err)
	}

	//
	logger, err := globals.GetLogger(logLevelFlag, disableTraceFlag)
	if err != nil {
		log.Fatal(err)
	}

	// Configure application's context
	ctx := v1alpha1.Context{
		Config: &v1alpha1.ConfigSpec{},
		Logger: logger,
	}

	// Get and parse the config
	configContent, err := config.ReadFile(configPath)
	if err != nil {
		logger.Fatalf(fmt.Sprintf(ConfigNotParsedErrorMessage, err))
	}

	// Set the configuration inside the global context
	ctx.Config = &configContent

	// Load default values
	if ctx.Config.Server.ListenAddress == "" {
		ctx.Config.Server.ListenAddress = defaultListenAddress
	}

	ctx.Logger.Info("Starting Akapurgo webserver in ", ctx.Config.Server.ListenAddress)

	// Create the akamai config file if not exists
	err = config.CreateAkamaiConfigFile(ctx)
	if err != nil {
		ctx.Logger.Fatalf("Error creating Akamai config file: %v", err)
	}

	// Get the base path for the templates and static files
	basePath, err := os.Getwd()
	if err != nil {
		ctx.Logger.Fatalf("error getting base directory: %v", err)
	}

	templatesPath := filepath.Join(basePath, "public", "templates")
	staticPath := filepath.Join(basePath, "public", "static")

	// Create a new Fiber app with the HTML template engine
	engine := html.New(templatesPath, ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// Define the routes

	// Static pages
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{})
	})
	app.Static("/static", staticPath)

	// API
	app.Post("/api/v1/purge", api.PurgeHandler(ctx))

	// Start the webserver
	err = app.Listen(ctx.Config.Server.ListenAddress)
	if err != nil {
		ctx.Logger.Fatalf("Error starting the webserver: %v", err)
	}
}
