package site

import (
	"context"
	"testing"

	"github.com/spf13/viper"
)

func TestSiteGen(t *testing.T) {
	viper.SetConfigFile("./.env")
	viper.ReadInConfig()

	apiKey := viper.Get("ANTHROPIC_API_KEY").(string)

	flowMachine, err := NewSiteBuildSteps(apiKey)
	if err != nil {
		t.Fatalf("Failed to create site build steps: %v", err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "input", "Create a html page with a story about data magician and a mission")

	err = flowMachine.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute flow machine: \n%v", err)
	}
}
