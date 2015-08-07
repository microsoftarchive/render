package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type RunDefinition struct {
	Processes map[string]string
	CronJobs  []string `json:"cron_jobs"`
}

type Manifest struct {
	ServerType   string `json:omit`
	Platform     string `json:"platform"`
	BuildCommand string `json:"build"`
	Owner        []string
	Types        map[string]RunDefinition
	Repos        []Repo
}

type Repo struct {
	Url    string
	Folder string
	ShaRaw string `json:"sha"`
}

func (r *Repo) Sha() string {
	if r.ShaRaw == "" {
		return "origin/master"
	}
	return r.ShaRaw
}

func firstNonEmpty(possibilities []string) string {
	for _, v := range possibilities {
		if v != "" {
			return v
		}
	}
	return ""
}

func (m *Manifest) Processes() map[string]string {
	return m.Types[m.ServerType].Processes
}

func (m *Manifest) CronJobs() []string {
	return m.Types[m.ServerType].CronJobs
}

func (m *Manifest) Get(key string) string {
	switch {
	case key == "owner":
		return strings.Join(m.Owner, "\n")
	case key == "processes":
		processes := []string{}
		for k, v := range m.Processes() {
			processes = append(processes, fmt.Sprintf("%s %s", k, v))
		}
		return strings.Join(processes, "\n")
	case key == "cron_jobs":
		return strings.Join(m.CronJobs(), "\n")
	case key == "repos":
		repos := []string{}
		for _, r := range m.Repos {
			repos = append(repos, fmt.Sprintf("%s %s %s", r.Folder, r.Url, r.Sha()))
		}
		return strings.Join(repos, "\n")
	}
	return ""
}

func NewManifest(path string, serverType string) Manifest {
	m := Manifest{}
	m.ServerType = serverType
	content, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = json.Unmarshal(content, &m)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return m
}
