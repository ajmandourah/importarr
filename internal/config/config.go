package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"

	"importarr/internal/models"
)

func Load() ([]models.Instance, error) {
	instances := loadYAML()
	if len(instances) == 0 {
		instances = loadEnv()
	}
	return instances, nil
}

func loadEnv() []models.Instance {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("$HOME/.importarr/.env")

	prefixes := []string{"SONARR", "RADARR"}
	var instances []models.Instance

	for _, prefix := range prefixes {
		url := os.Getenv(prefix + "_URL")
		key := os.Getenv(prefix + "_API_KEY")

		if url == "" || key == "" {
			continue
		}

		instances = append(instances, models.Instance{
			Name:   strings.ToLower(prefix),
			Type:   strings.ToLower(prefix),
			URL:    url,
			APIKey: key,
		})
	}
	return instances
}

func loadYAML() []models.Instance {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil
	}

	var cfg struct {
		Instances []struct {
			Name   string `yaml:"name"`
			Type   string `yaml:"type"`
			URL    string `yaml:"url"`
			APIKey string `yaml:"api_key"`
		} `yaml:"instances"`
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil
	}

	var instances []models.Instance
	for _, inst := range cfg.Instances {
		if inst.URL == "" || inst.APIKey == "" {
			continue
		}
		instances = append(instances, models.Instance{
			Name:   inst.Name,
			Type:   inst.Type,
			URL:    inst.URL,
			APIKey: inst.APIKey,
		})
	}
	return instances
}

func FilterInstances(instances []models.Instance, name string, all bool) []models.Instance {
	if all {
		return instances
	}
	if name != "" {
		for _, inst := range instances {
			if inst.Name == name {
				return []models.Instance{inst}
			}
		}
		fmt.Fprintf(os.Stderr, "instance not found: %s\n", name)
		os.Exit(1)
	}
	return instances
}
