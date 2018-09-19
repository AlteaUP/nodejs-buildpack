package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	libbuildpackV3 "github.com/buildpack/libbuildpack"
	"nodejs/v3/detect"
	"os"
)

func main() {
	detector, err := libbuildpackV3.DefaultDetect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create default detector: %s", err)
		os.Exit(100)
	}

	err = detect.UpdateBuildPlan(&detector)
	if err != nil {
		detector.Logger.Debug("failed nodejs detection: %s", err)
		detector.Fail()
	}

	encoder := toml.NewEncoder(os.Stdout)
	err = encoder.Encode(detector.BuildPlan)
	if err != nil {
		detector.Logger.Debug("failed to write build plan: %s", err)
		detector.Fail()
	}
}
