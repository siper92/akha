package flow_v2

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	ds "github.com/siper92/akha/internal/data_sources"
	tu "github.com/siper92/akha/lang/tests_utils"
)

var errBadRace = errors.New("bad race")

type Game struct {
	GameType   string
	Player     string
	Difficulty string
	Score      int
	Race       string
}

type Player struct {
	Name  string `json:"name"`
	Race  string `json:"race"`
	Score int    `json:"score"`
}

type GamePop struct {
	GameType   string   `json:"gameType"`
	Players    []Player `json:"players"`
	Difficulty string   `json:"difficulty"`
}

func readExt[T any](ext string) Loader[T] {
	return func(ctx context.Context, rt Runtime) (T, error) {
		var zero T
		dir, err := rt.Services().Dir(ctx, ".")
		if err != nil {
			return zero, err
		}
		files, err := dir.Files(ctx)
		if err != nil {
			return zero, err
		}
		for _, f := range files {
			if f.Ext() == ext {
				return ds.Read[T](ctx, f)
			}
		}
		return zero, os.ErrNotExist
	}
}

func lastResult[R any](id string) Loader[R] {
	return func(_ context.Context, rt Runtime) (R, error) {
		var zero R
		r, err := Results[R](rt.State(), id)
		if err != nil || len(r) == 0 {
			return zero, err
		}
		return r[len(r)-1], nil
	}
}

func gameFlow(root string) (*Machine, error) {
	reads := 0
	readText := NewStep("readText", readExt[string]("txt"), func([]string) Outcome {
		if reads++; reads < 2 {
			return OutcomeFollow
		}
		return OutcomeOK
	})
	parse := NewStep[string, []string]("parse", lastResult[string]("readText"), nil)
	readCSV := NewStep[[]ds.Row, []Game]("readCSV", readExt[[]ds.Row]("csv"), nil)
	buildJSON := NewStep[[]Game, []GamePop]("buildJSON", lastResult[[]Game]("readCSV"), nil)

	plan := NewGraph("readText").
		Edge("readText", OutcomeOK, "parse").
		Edge("readText", OutcomeFollow, "readText").
		Edge("parse", OutcomeOK, "readCSV").
		Edge("readCSV", OutcomeOK, "buildJSON").
		Edge("buildJSON", OutcomeOK, End)

	inits := []error{
		readText.Init([]Task[string, string]{NewTask("trim", nil, func(_ context.Context, s string) (string, error) {
			return strings.TrimSpace(s), nil
		})}, plan),
		parse.Init([]Task[string, []string]{NewTask("lines", nil, func(_ context.Context, s string) ([]string, error) {
			return strings.Split(s, "\n"), nil
		})}, plan),
		readCSV.Init([]Task[[]ds.Row, []Game]{NewTask("games", validateRaces, toGames)}, plan),
		buildJSON.Init([]Task[[]Game, []GamePop]{NewTask("pop", nil, toGamePops)}, plan),
	}
	if err := errors.Join(inits...); err != nil {
		return nil, err
	}

	m := NewMachine(10)
	sp := ds.NewFolderProvider(root, ds.Handlers{"txt": ds.NewTextFile, "csv": ds.CSVHandler(true)})
	return m, m.Init(plan, []Step{readText, parse, readCSV, buildJSON}, sp)
}

func validateRaces(rows []ds.Row) error {
	for _, r := range rows {
		if !slices.Contains([]string{"human", "elf", "orc", "alien"}, r["race"]) {
			return errBadRace
		}
	}
	return nil
}

func toGames(_ context.Context, rows []ds.Row) ([]Game, error) {
	games := make([]Game, 0, len(rows))
	for _, r := range rows {
		score, err := strconv.Atoi(r["score"])
		if err != nil {
			return nil, err
		}
		games = append(games, Game{r["gameType"], r["players"], r["difficulty"], score, r["race"]})
	}
	return games, nil
}

func toGamePops(_ context.Context, games []Game) ([]GamePop, error) {
	var pops []GamePop
	idx := map[string]int{}
	for _, g := range games {
		i, ok := idx[g.GameType]
		if !ok {
			i = len(pops)
			idx[g.GameType] = i
			pops = append(pops, GamePop{GameType: g.GameType, Difficulty: g.Difficulty})
		}
		pops[i].Players = append(pops[i].Players, Player{g.Player, g.Race, g.Score})
	}
	return pops, nil
}

// --- game flow from text, csv to GamePop

func TestGameFlow(t *testing.T) {
	csv := "gameType,players,difficulty,score,race\n" +
		"chess,ann,hard,10,human\n" +
		"chess,bob,hard,7,elf\n" +
		"quest,zug,easy,3,orc\n"

	cases := []tu.Case[string, []GamePop]{
		{
			Name:  "csv_grouped_by_game_type",
			Input: csv,
			Expected: []GamePop{
				{GameType: "chess", Difficulty: "hard", Players: []Player{{"ann", "human", 10}, {"bob", "elf", 7}}},
				{GameType: "quest", Difficulty: "easy", Players: []Player{{"zug", "orc", 3}}},
			},
		},
		{
			Name:  "unknown_race_fails",
			Input: "gameType,players,difficulty,score,race\nchess,ann,hard,10,dwarf\n",
			Err:   errBadRace,
		},
	}

	tu.Run(tu.New(t), cases, func(body string) ([]GamePop, error) {
		root := t.TempDir()
		_ = os.WriteFile(filepath.Join(root, "plan.txt"), []byte("readText\nparse\nreadCSV\nbuildJSON\n"), 0o644)
		_ = os.WriteFile(filepath.Join(root, "games.csv"), []byte(body), 0o644)
		m, err := gameFlow(root)
		if err != nil {
			return nil, err
		}
		if err := m.Run(context.Background()); err != nil {
			return nil, err
		}
		return lastResult[[]GamePop]("buildJSON")(context.Background(), m)
	}, nil)
}
