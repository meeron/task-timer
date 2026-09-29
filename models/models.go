package models

type Task struct {
	Id          string
	Name        string
	Description string
	StartUnix   int64
	Duration    int64
}
