package run

import (
	"akapurgo/api/v1alpha1"
	"akapurgo/internal/api"
	"akapurgo/internal/config"
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

	return cmd
}

func RunCommand(cmd *cobra.Command, args []string) {

	// Check the flags for this command
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		log.Fatalf("Error getting configuration file path: %v", err)
	}

	// Configure application's context
	ctx := v1alpha1.Context{
		Config: &v1alpha1.ConfigSpec{},
	}

	// Get and parse the config
	configContent, err := config.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error parsing configuration file: %v", err)
	}

	// Set the configuration inside the global context
	ctx.Config = &configContent

	// Load default values
	if ctx.Config.Server.ListenAddress == "" {
		ctx.Config.Server.ListenAddress = defaultListenAddress
	}

	log.Println("Starting Akapurgo webserver in ", ctx.Config.Server.ListenAddress)

	// Create the akamai config file if not exists
	err = config.CreateAkamaiConfigFile(ctx)
	if err != nil {
		log.Fatalf("Error creating Akamai config file: %v", err)
	}

	// Get the base path for the templates and static files
	basePath, err := os.Getwd()
	if err != nil {
		log.Fatalf("error getting base directory: %v", err)
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
		log.Fatalf("Error starting the webserver: %v", err)
	}
}
