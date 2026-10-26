package data_sources

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	tu "github.com/siper92/akha/lang/tests_utils"
)

type meta struct {
	Name string `json:"name"`
}

// --- folder files read by extension handler

func TestFolderRead(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("a.json", `{"name":"games"}`)
	write("b.csv", "gameType,score\nchess,10\n")
	write("c.txt", "hello")

	folder, err := NewFolder(dir, Handlers{
		"json": JSONHandler[meta](),
		"csv":  CSVHandler(true),
		"txt":  NewTextFile,
	})
	if err != nil {
		t.Fatal(err)
	}
	files, err := folder.Files(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	cases := []tu.Case[int, any]{
		{
			Name:     "json_to_struct",
			Input:    0,
			Expected: meta{Name: "games"},
		},
		{
			Name:     "csv_with_header",
			Input:    1,
			Expected: []Row{{"gameType": "chess", "score": "10"}},
		},
		{
			Name:     "text_content",
			Input:    2,
			Expected: "hello",
		},
	}

	tu.Run(tu.New(t), cases, func(i int) (any, error) {
		switch i {
		case 0:
			return Read[meta](context.Background(), files[i])
		case 1:
			return Read[[]Row](context.Background(), files[i])
		default:
			return Read[string](context.Background(), files[i])
		}
	}, nil)

	if _, err := Read[string](context.Background(), files[0]); err == nil {
		t.Fatalf("expected %v", ErrFileType)
	}
}
