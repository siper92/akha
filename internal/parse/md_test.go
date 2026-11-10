package parse

import "testing"

const sample = `# @id: tasks
# @idea
execute the task and provide the output

# @tools
get_code_source: get the code(package)
get_input_task: get the input task

# @task1
text text 1
## test1 - section 2
adasda asd asdadsa

# @test2
qwerwer werwer werwerwerwerwer wer
qwerwer werwer werwerwerwerwer wer
qwerwer werwerwerwerwer wer
`

func TestParse(t *testing.T) {
	step, err := Parse(sample)
	if err != nil {
		t.Fatal(err)
	}

	if step.Name() != "tasks" {
		t.Fatalf("name = %q", step.Name())
	}

	wantIDs := []string{"id", "idea", "tools", "task1", "test2"}
	secs := step.Content()
	if len(secs) != len(wantIDs) {
		t.Fatalf("got %d sections, want %d", len(secs), len(wantIDs))
	}
	for i, id := range wantIDs {
		if secs[i].Title().ID() != id {
			t.Errorf("section %d id = %q, want %q", i, secs[i].Title().ID(), id)
		}
	}

	idea, _ := step.Section(IDEA_SECTION_KEY)
	if got := idea.Content().String(); got != "execute the task and provide the output" {
		t.Errorf("idea content = %q", got)
	}

	task1, _ := step.Section("task1")
	wantContent := "text text 1\n## test1 - section 2\nadasda asd asdadsa"
	if got := task1.Content().String(); got != wantContent {
		t.Errorf("task1 content = %q", got)
	}
	if got := task1.String(); got != "# @task1\n"+wantContent {
		t.Errorf("task1 raw = %q", got)
	}

	id, _ := step.Section(ID_SECTION_KEY)
	if id.String() != "# @id: tasks" || id.Content().String() != "tasks" {
		t.Errorf("id section = %q / %q", id.String(), id.Content().String())
	}

	tools, _ := step.Section(TOOLS_SECTION_KEY)
	tm := ParseTools(tools)
	if tm["get_code_source"] != "get the code(package)" || tm["get_input_task"] != "get the input task" {
		t.Errorf("tools = %v", tm)
	}
}

func TestParseMissingID(t *testing.T) {
	if _, err := Parse("# @idea\nfoo\n"); err != ErrMissingID {
		t.Fatalf("err = %v", err)
	}
}
