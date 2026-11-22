package site

import (
	"context"
	"testing"

	"github.com/siper92/akha/projects/site/internal/flow"
	"github.com/spf13/viper"
)

func TestSiteGen(t *testing.T) {
	viper.SetConfigFile("./.env")
	viper.ReadInConfig()

	apiKey := viper.Get("ANTHROPIC_API_KEY").(string)

	flowMachine, err := flow.NewSiteBuildSteps(apiKey, "_env/.cache")
	if err != nil {
		t.Fatalf("Failed to create site build steps: %v", err)
	}

	ctx := context.Background()
	ctx = context.WithValue(
		ctx, "input", "write a story about vaice, power and the unholy path to light",
	)

	err = flowMachine.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute flow machine: \n%v", err)
	}
}
