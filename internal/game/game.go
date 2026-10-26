package game

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	ds "github.com/siper92/akha/internal/data_sources"
	flow "github.com/siper92/akha/internal/flow_v2"
)

var (
	ErrNoStory       = errors.New("no story files")
	ErrNoTitle       = errors.New("story without title")
	ErrUnknownPlayer = errors.New("unknown player")
)

const (
	StoryPrefix = "story_"
	TitlePrefix = "# title:"
	PlayersFile = "players.csv"
	ScoresFile  = "scores.csv"
)

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

type Content struct {
	Text  string
	Read  int
	Total int
}

type Story struct {
	Title string
	Text  string
}

type Tables struct {
	Players []ds.Row
	Scores  []ds.Row
}

type Entry struct {
	GameType   string
	Name       string
	Race       string
	Difficulty string
	Score      int
}

func Handlers() ds.Handlers {
	return ds.Handlers{
		"txt":  ds.NewTextFile,
		"csv":  ds.CSVHandler(true),
		"json": ds.JSONHandler[[]GamePop](),
	}
}

func Steps(plan flow.ExecutionPlan) ([]flow.Step, error) {
	readGame := flow.NewStep[Content, Content]("readGame", readStory, more)
	parse := flow.NewStep[Content, []Story]("parse", last[Content]("readGame"), nil)
	readCSV := flow.NewStep[Tables, []Entry]("readCSV", readTables, nil)
	buildJSON := flow.NewStep[[]Entry, []GamePop]("buildJSON", last[[]Entry]("readCSV"), nil)

	err := errors.Join(
		readGame.Init([]flow.Task[Content, Content]{flow.NewTask("aggregate", nil, func(_ context.Context, c Content) (Content, error) {
			return c, nil
		})}, plan),
		parse.Init([]flow.Task[Content, []Story]{flow.NewTask("stories", nil, parseStories)}, plan),
		readCSV.Init([]flow.Task[Tables, []Entry]{flow.NewTask("join", nil, joinTables)}, plan),
		buildJSON.Init([]flow.Task[[]Entry, []GamePop]{flow.NewTask("pop", nil, toGamePops)}, plan),
	)
	return []flow.Step{readGame, parse, readCSV, buildJSON}, err
}

func last[R any](id string) flow.Loader[R] {
	return func(_ context.Context, rt flow.Runtime) (R, error) {
		var zero R
		rs, err := flow.Results[R](rt.State(), id)
		if err != nil {
			return zero, err
		}
		if len(rs) == 0 {
			return zero, fmt.Errorf("%w: %s", flow.ErrNoResult, id)
		}
		return rs[len(rs)-1], nil
	}
}

func files(ctx context.Context, rt flow.Runtime, keep func(ds.File) bool) ([]ds.File, error) {
	dir, err := rt.Services().Dir(ctx, ".")
	if err != nil {
		return nil, err
	}
	all, err := dir.Files(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]ds.File, 0, len(all))
	for _, f := range all {
		if keep(f) {
			out = append(out, f)
		}
	}
	return out, nil
}

func isStory(f ds.File) bool {
	return f.Ext() == "txt" && strings.HasPrefix(filepath.Base(f.Path()), StoryPrefix)
}

func readStory(ctx context.Context, rt flow.Runtime) (Content, error) {
	stories, err := files(ctx, rt, isStory)
	if err != nil {
		return Content{}, err
	}
	if len(stories) == 0 {
		return Content{}, ErrNoStory
	}
	prev, err := last[Content]("readGame")(ctx, rt)
	if err != nil && !errors.Is(err, flow.ErrNoResult) {
		return Content{}, err
	}
	if prev.Read >= len(stories) {
		return prev, nil
	}
	text, err := ds.Read[string](ctx, stories[prev.Read])
	if err != nil {
		return Content{}, err
	}
	if prev.Text != "" {
		text = prev.Text + "\n" + text
	}
	return Content{Text: text, Read: prev.Read + 1, Total: len(stories)}, nil
}

func more(rs []Content) flow.Outcome {
	if len(rs) > 0 && rs[len(rs)-1].Read < rs[len(rs)-1].Total {
		return flow.OutcomeFollow
	}
	return flow.OutcomeOK
}

func parseStories(_ context.Context, c Content) ([]Story, error) {
	var stories []Story
	for _, line := range strings.Split(c.Text, "\n") {
		if title, ok := strings.CutPrefix(line, TitlePrefix); ok {
			stories = append(stories, Story{Title: strings.TrimSpace(title)})
			continue
		}
		if len(stories) == 0 {
			if strings.TrimSpace(line) != "" {
				return nil, ErrNoTitle
			}
			continue
		}
		stories[len(stories)-1].Text += line + "\n"
	}
	for i := range stories {
		stories[i].Text = strings.TrimSpace(stories[i].Text)
	}
	return stories, nil
}

func readTables(ctx context.Context, rt flow.Runtime) (Tables, error) {
	csvs, err := files(ctx, rt, func(f ds.File) bool { return f.Ext() == "csv" })
	if err != nil {
		return Tables{}, err
	}
	var t Tables
	for _, f := range csvs {
		rows, err := ds.Read[[]ds.Row](ctx, f)
		if err != nil {
			return Tables{}, err
		}
		switch filepath.Base(f.Path()) {
		case PlayersFile:
			t.Players = rows
		case ScoresFile:
			t.Scores = rows
		}
	}
	return t, nil
}

func joinTables(_ context.Context, t Tables) ([]Entry, error) {
	players := make(map[string]ds.Row, len(t.Players))
	for _, p := range t.Players {
		players[p["playerID"]] = p
	}
	entries := make([]Entry, 0, len(t.Scores))
	for _, s := range t.Scores {
		p, ok := players[s["playerID"]]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnknownPlayer, s["playerID"])
		}
		score, err := strconv.Atoi(s["score"])
		if err != nil {
			return nil, err
		}
		entries = append(entries, Entry{
			GameType:   s["game_type"],
			Name:       p["name"],
			Race:       p["race"],
			Difficulty: s["difficulty"],
			Score:      score,
		})
	}
	return entries, nil
}

func toGamePops(_ context.Context, entries []Entry) ([]GamePop, error) {
	var pops []GamePop
	idx := map[string]int{}
	for _, e := range entries {
		i, ok := idx[e.GameType]
		if !ok {
			i = len(pops)
			idx[e.GameType] = i
			pops = append(pops, GamePop{GameType: e.GameType, Difficulty: e.Difficulty})
		}
		pops[i].Players = append(pops[i].Players, Player{e.Name, e.Race, e.Score})
	}
	return pops, nil
}
