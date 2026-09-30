package models

type Task struct {
	Id          string
	Name        string
	Description string
	StartUnix   int64
	Duration    int64
	// LoggedUnix is when the task was last logged to Jira (Unix seconds); 0 if never.
	LoggedUnix int64
}
