package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/Iknite-Space/sqlc-example-api/db/repo"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal("Error:", err)
	}
}

func run() error {
	ctx := context.Background()

	// Load environment
	if _, err := os.Stat(".env"); err == nil {
		err = godotenv.Load()
		if err != nil {
			return fmt.Errorf("failed to load env file: %w", err)
		}
	}

	// Connect to database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable is required")
	}

	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	querier := repo.New(db)

	// Seed dashboard data
	err = seedDashboardData(ctx, querier)
	if err != nil {
		return fmt.Errorf("failed to seed dashboard data: %w", err)
	}

	// Seed editor config
	err = seedEditorConfig(ctx, querier)
	if err != nil {
		return fmt.Errorf("failed to seed editor config: %w", err)
	}

	fmt.Println("Data seeding completed successfully!")
	return nil
}

func seedDashboardData(ctx context.Context, querier *repo.Queries) error {
	// Read dashboard.json
	data, err := ioutil.ReadFile("mydata/dashboard.json")
	if err != nil {
		return fmt.Errorf("failed to read dashboard.json: %w", err)
	}

	var dashboard map[string]interface{}
	err = json.Unmarshal(data, &dashboard)
	if err != nil {
		return fmt.Errorf("failed to parse dashboard.json: %w", err)
	}

	// Seed docker containers
	if containers, ok := dashboard["dockerContainers"].([]interface{}); ok {
		for _, containerData := range containers {
			container := containerData.(map[string]interface{})

			var lastRun *time.Time
			if lastRunStr, ok := container["lastRun"].(string); ok {
				if parsed, err := time.Parse("2006-01-02", lastRunStr); err == nil {
					lastRun = &parsed
				}
			}

			params := repo.CreateDockerContainerParams{
				Name:         container["name"].(string),
				Status:       container["status"].(string),
				LastRun:      lastRun,
				Origin:       getStringPtr(container["origin"]),
				Disk:         getStringPtr(container["disk"]),
				Ram:          getStringPtr(container["ram"]),
				Unused:       getBoolPtr(container["unused"]),
				HighMem:      getBoolPtr(container["highMem"]),
				Root:         getBoolPtr(container["root"]),
				ExposedPorts: getBoolPtr(container["exposedPorts"]),
				UnlimitedMem: getBoolPtr(container["unlimitedMem"]),
			}

			_, err := querier.CreateDockerContainer(ctx, params)
			if err != nil {
				log.Printf("Failed to create container %s: %v", container["name"], err)
			} else {
				fmt.Printf("Created container: %s\n", container["name"])
			}
		}
	}

	// Seed git repositories
	if repos, ok := dashboard["gitRepos"].([]interface{}); ok {
		for _, repoData := range repos {
			repo := repoData.(map[string]interface{})

			params := repo.CreateGitRepoParams{
				Name:             repo["name"].(string),
				Untouched:        getStringPtr(repo["untouched"]),
				Duplicate:        getBoolPtr(repo["duplicate"]),
				ClonedNeverBuilt: getBoolPtr(repo["clonedNeverBuilt"]),
			}

			_, err := querier.CreateGitRepo(ctx, params)
			if err != nil {
				log.Printf("Failed to create repo %s: %v", repo["name"], err)
			} else {
				fmt.Printf("Created repository: %s\n", repo["name"])
			}
		}
	}

	return nil
}

func seedEditorConfig(ctx context.Context, querier *repo.Queries) error {
	// Read editorConfig.json
	data, err := ioutil.ReadFile("mydata/editorConfig.json")
	if err != nil {
		return fmt.Errorf("failed to read editorConfig.json: %w", err)
	}

	params := repo.CreateEditorConfigParams{
		Name:       "Default Editor Config",
		ConfigData: data,
	}

	_, err = querier.CreateEditorConfig(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create editor config: %w", err)
	}

	fmt.Println("Created editor configuration")
	return nil
}

// Helper functions
func getStringPtr(v interface{}) *string {
	if v == nil {
		return nil
	}
	if str, ok := v.(string); ok {
		return &str
	}
	return nil
}

func getBoolPtr(v interface{}) *bool {
	if v == nil {
		return nil
	}
	if b, ok := v.(bool); ok {
		return &b
	}
	return nil
}
