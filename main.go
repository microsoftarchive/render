package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type ConsulEnv struct {
	Short string
	Key   string
	Value string
}

var debug bool

func init() {
	if os.Getenv("DEBUG") != "" {
		debug = true
	} else {
		debug = false
	}
}

func writeRunit(m Manifest, app string) {
	if debug {
		fmt.Println("Write runit")
	}
	for k, v := range m.Processes() {
		p := Process{App: app, Type: m.ServerType, Cmd: v, Process: k}
		p.SetDefaults()
		err := os.MkdirAll(fmt.Sprintf("tmp/sv/%s/log", p.Process), 0755)
		if err != nil {
			fmt.Println(err)
		}
		err = ioutil.WriteFile(
			fmt.Sprintf("tmp/sv/%s/run", p.Process),
			[]byte(p.RenderRun()),
			0644,
		)
		if err != nil {
			fmt.Println(err)
		}
		err = ioutil.WriteFile(
			fmt.Sprintf("tmp/sv/%s/log/run", p.Process),
			[]byte(p.RenderLog()),
			0644,
		)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func writeCron(m Manifest) {
	if debug {
		fmt.Println("Write crontab")
	}
	c := CronJobs{m.CronJobs()}
	path := "tmp/crontab"
	err := ioutil.WriteFile(path, []byte(c.Render()), 0644)
	if err != nil {
		if debug {
			fmt.Fprintf(os.Stderr, "  Cannot write crontab to %s\n", path)
		}
		os.Exit(1)
	}
}

func cloneRepos(m Manifest) {
	if debug {
		fmt.Println("Clone repos")
	}
	err := os.MkdirAll("tmp/repos", 0755)
	if err != nil {
		fmt.Println(err)
	}
	for _, r := range m.Repos {
		path := fmt.Sprintf("tmp/repos/%s", r.Folder)
		cmd := exec.Command("git", "clone", r.Url, path)
		_, err := cmd.Output()
		if err != nil {
			fmt.Println(err)
		}
		curr, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
		}
		err = os.Chdir(path)
		if err != nil {
			fmt.Println(err)
		}
		cmd = exec.Command("git", "reset", "--hard", r.Sha())
		_, err = cmd.Output()
		if err != nil {
			fmt.Println(err)
		}
		err = os.Chdir(curr)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func writeConsulEnv(app string) {
	if debug {
		fmt.Println("Write consul env")
	}
	path := fmt.Sprintf("%s/env", app)
	cmd := exec.Command("wake", "config", "--raw", path)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}
	envs := make([]ConsulEnv, 0)
	err = json.Unmarshal(out, &envs)
	if err != nil {
		fmt.Println(err)
	}
	envString := ""
	for _, v := range envs {
		envString = envString + fmt.Sprintf("export %s='%s'\n", v.Short, v.Value)
	}
	err = ioutil.WriteFile(
		"tmp/env.sh",
		[]byte(envString),
		0644,
	)
	if err != nil {
		fmt.Println(err)
	}
}

func writeApi() {
	fmt.Println("TODO: write api dependencies.")
}

func writeDatabaseConfig(app string) {
	if debug {
		fmt.Println("Write db config")
	}
	path := fmt.Sprintf("%s/env/database_url", app)
	cmd := exec.Command("wake", "config", "--raw", path)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}
	envs := make([]ConsulEnv, 0)
	err = json.Unmarshal(out, &envs)
	if err != nil {
		fmt.Println(err)
	}
	if len(envs) > 0 {
		d := DatabaseUrl{Url: envs[len(envs)-1].Value}
		err = ioutil.WriteFile(
			"tmp/database.yml",
			[]byte(d.RenderYAML()),
			0644,
		)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func buildExecutable(platform string, app string, rev string) {
	if debug {
		fmt.Println("Build executable")
	}
	err := os.MkdirAll("tmp/bin", 0755)
	if err != nil {
		fmt.Println(err)
	}
	if platform == "golang" {
		for _, p := range strings.Split(os.Getenv("PROJECT_PATH"), ":") {
			project := fmt.Sprintf("%s/%s", p, app)
			f, err := os.Open(project)
			if err == nil {
				if i, _ := f.Stat(); i.IsDir() {

					curr, err := os.Getwd()
					if err != nil {
						fmt.Println(err)
						return
					}
					err = os.Chdir(project)
					if err != nil {
						fmt.Println(err)
						return
					}
					cmd := exec.Command("git", "checkout", rev)
					_, err = cmd.Output()
					if err != nil {
						if debug {
							fmt.Fprintf(os.Stderr, "  Cannot checkout %s: %s\n", rev, err)
						}
						os.Exit(1)
					}
					target := fmt.Sprintf("%s/tmp/bin/%s", curr, app)
					cmd = exec.Command("go", "build", "-o", target)
					cmd.Env = append(os.Environ(), "GOOS=linux")
					cmd.Env = append(cmd.Env, "GOARCH=amd64")
					cmd.Env = append(cmd.Env, "CGO_ENABLED=0")
					_, err = cmd.Output()
					if err != nil {
						fmt.Fprintf(os.Stderr, "  Cannot build %s: %s\n", app, err)
						return
					}
					cmd = exec.Command("git", "checkout", "-")
					_, err = cmd.Output()
					if err != nil {
						fmt.Fprintf(os.Stderr, "  Cannot checkout %s: %s\n", rev, err)
						return
					}
					err = os.Chdir(curr)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
			}
		}
	}
}

func findProjectPath(app string) string {
	for _, p := range strings.Split(os.Getenv("PROJECT_PATH"), ":") {
		path := fmt.Sprintf("%s/%s", p, app)
		f, err := os.Open(path)
		if err == nil {
			if i, _ := f.Stat(); i.IsDir() {
				return path
			}
		}
	}
	if debug {
		fmt.Fprintf(os.Stderr, "Cannot find project %s in $PROJECT_PATH(%s)\n", app, os.Getenv("PROJECT_PATH"))
	}
	os.Exit(1)
	return ""
}

func main() {
	var app = flag.String("app", "", "app name")
	var path = flag.String("path", "", "path to manifest")
	var serverType = flag.String("type", "", "server type")
	var rev = flag.String("rev", "origin/master", "revision")
	flag.Parse()

	projectPath := findProjectPath(*app)

	var manifestPath string
	if *path == "" {
		manifestPath = fmt.Sprintf("%s/manifest.json", projectPath)
	} else {
		manifestPath = *path
	}
	m := NewManifest(manifestPath, *serverType)

	writeRunit(m, *app)
	writeCron(m)
	cloneRepos(m)
	writeConsulEnv(*app)
	writeApi()
	buildExecutable(m.Platform, *app, *rev)
	writeDatabaseConfig(*app)
}
