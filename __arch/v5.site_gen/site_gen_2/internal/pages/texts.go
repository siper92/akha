package pages

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadTexts(ctx context.Context, path string) (Texts, error) {
	if err := ctx.Err(); err != nil {
		return Texts{}, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return Texts{}, fmt.Errorf("%w: %v", ErrTexts, err)
	}
	var t Texts
	if err := yaml.Unmarshal(raw, &t); err != nil {
		return Texts{}, fmt.Errorf("%w: %s: %v", ErrTexts, path, err)
	}
	if t.SiteName == "" {
		return Texts{}, fmt.Errorf("%w: %s: site_name is empty", ErrTexts, path)
	}
	return t, nil
}

func Hash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
