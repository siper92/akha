-- Flow project, one per generation run of a design flow
CREATE TABLE IF NOT EXISTS flow_projects (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  design_flow TEXT NOT NULL,
  clarify_path TEXT NOT NULL DEFAULT '',
  plan_path TEXT NOT NULL DEFAULT '',
  meta TEXT NOT NULL DEFAULT '{}',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Execution of a project
CREATE TABLE IF NOT EXISTS flow_runs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  project_id INTEGER NOT NULL REFERENCES flow_projects(id),
  status TEXT NOT NULL DEFAULT 'executing' CHECK (status IN ('executing', 'completed', 'failed')),
  started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  finished_at TIMESTAMP
);

-- Story of a run, the list of tasks to execute the plan
CREATE TABLE IF NOT EXISTS flow_stories (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id INTEGER NOT NULL REFERENCES flow_runs(id),
  file_path TEXT NOT NULL,
  story_key TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'executing' CHECK (status IN ('executing', 'completed', 'failed')),
  started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  finished_at TIMESTAMP,
  run_time_ms INTEGER NOT NULL DEFAULT 0
);

-- Task of a story
CREATE TABLE IF NOT EXISTS flow_tasks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  story_id INTEGER NOT NULL REFERENCES flow_stories(id),
  position INTEGER NOT NULL DEFAULT 0,
  task_key TEXT NOT NULL DEFAULT '',
  label TEXT NOT NULL,
  content TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'executing' CHECK (status IN ('executing', 'completed', 'failed')),
  error TEXT NOT NULL DEFAULT '',
  attempts INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indices for performance
CREATE INDEX IF NOT EXISTS idx_flow_runs_project_id ON flow_runs(project_id);
CREATE INDEX IF NOT EXISTS idx_flow_stories_run_id ON flow_stories(run_id);
CREATE INDEX IF NOT EXISTS idx_flow_tasks_story_id ON flow_tasks(story_id);
