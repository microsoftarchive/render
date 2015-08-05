package main

import (
	"bytes"
	"fmt"
	"net/url"
	"text/template"
)

type DatabaseUrl struct {
	Url      string
	Host     string
	User     string
	Password string
	Scheme   string
	Database string
	Port     string
}

func (d *DatabaseUrl) RenderYAML() string {
	parsedUrl, err := url.Parse(d.Url)
	if err != nil {
		fmt.Println(err)
	}
	d.Scheme = parsedUrl.Scheme
	d.Database = parsedUrl.Path
	d.Host = parsedUrl.Host
	d.User = parsedUrl.User.Username()
	d.Password, _ = parsedUrl.User.Password()

	databaseyml := `production:
  adapter: {{.Scheme}}
  host: {{.Host}}
  prepared_statements: false
  encoding: unicode
  pool: <%= ENV["DB_POOL"] || 1 %>
  port: {{.Port}}
  database: {{.Database}}
  username: {{.User}}
  password: {{.Password}}
`
	t, _ := template.New("databaseyml").Parse(databaseyml)
	buffer := bytes.NewBuffer(make([]byte, 0))
	_ = t.Execute(buffer, d)
	return buffer.String()
}
