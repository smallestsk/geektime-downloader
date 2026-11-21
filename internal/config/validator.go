package config

import (
	"fmt"
	"strings"
)

// ValidateConfig validates the application configuration.
func ValidateConfig(cfg *AppConfig) error {
	if err := validateCookies(cfg); err != nil {
		return err
	}
	if err := validateQuality(cfg); err != nil {
		return err
	}
	if err := validateLogLevel(cfg); err != nil {
		return err
	}
	if err := validateQuiet(cfg); err != nil {
		return err
	}
	return nil
}

func validateQuiet(cfg *AppConfig) error {
	if cfg.Quiet {
		if !validateProductType(cfg) || cfg.ProductID == "" {
			return fmt.Errorf("ProductType and ProductId must be specified correctly when quiet mode is enabled")
		}
		articleIDs := cfg.ArticleID
		productIDs := cfg.ProductID
		if articleIDs != "" && len(strings.Split(productIDs, ",")) != 1 {
			return fmt.Errorf("When specifying article IDs, only one product ID can be specified")
		}
	}
	return nil
}

func validateProductType(cfg *AppConfig) bool {
	if cfg.Quiet {
		if cfg.IsEnterprise && cfg.ProductType != 0 {
			return false
		} else if !cfg.IsEnterprise && (cfg.ProductType < 0 || cfg.ProductType > 5) {
			return false
		}

		return true
	}
	return true
}

func validateCookies(cfg *AppConfig) error {
	if cfg.Gcid == "" || cfg.Gcess == "" {
		return fmt.Errorf("arguments 'gcid' and 'gcess' are required and cannot be empty")
	}
	return nil
}

func validateQuality(cfg *AppConfig) error {
	validQuality := []string{"ld", "sd", "hd"}

	isValidQualityFlag := false
	for _, v := range validQuality {
		if cfg.Quality == v {
			isValidQualityFlag = true
			break
		}
	}

	if !isValidQualityFlag {
		return fmt.Errorf("argument 'quality' is not valid, must be one of ld, sd, hd")
	}

	return nil
}

func validateLogLevel(cfg *AppConfig) error {
	validLogLevels := []string{"debug", "info", "warn", "error", "none"}

	isValidLogLevel := false
	for _, v := range validLogLevels {
		if cfg.LogLevel == v {
			isValidLogLevel = true
			break
		}
	}

	if !isValidLogLevel {
		return fmt.Errorf("argument 'log-level' is not valid, must be one of debug, info, warn, error, none")
	}

	return nil
}

