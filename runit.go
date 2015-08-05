package main

import (
	"bytes"
	"text/template"
)

type Process struct {
	App     string
	Type    string
	Process string
	Cmd     string
	Ulimit  int
}

func (p *Process) SetDefaults() {
	if p.Ulimit == 0 {
		p.Ulimit = 65536
	}
}

func (p *Process) RenderLog() string {
	log := `#!/bin/bash
exec /opt/bin/logger -tag {{.App}}_{{.Type}}_{{.Process}} -syslog $RSYSLOG_HOST
`
	t, _ := template.New("log").Parse(log)
	buffer := bytes.NewBuffer(make([]byte, 0))
	_ = t.Execute(buffer, p)
	return buffer.String()
}

func (p *Process) RenderRun() string {
	log := `#!/bin/bash

if [ -e /opt/env.sh ]; then
  . /opt/env.sh
fi
cd /opt/app
chown app:app /dev/fd/1
exec chpst \
  -o {{.Ulimit}} \
  {{.Cmd}} 2>&1
`
	t, _ := template.New("log").Parse(log)
	buffer := bytes.NewBuffer(make([]byte, 0))
	_ = t.Execute(buffer, p)
	return buffer.String()
}
