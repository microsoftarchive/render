package main

import (
	"fmt"
	"regexp"
)

type CronJobs struct {
	Jobs []string
}

func (c *CronJobs) Render() string {
	re := regexp.MustCompile(`(([\w,\/*-]+\s){5})(.*)`)
	result := ""
	if len(c.Jobs) > 0 {
		result = "SHELL=/bin/bash\n"
		for _, v := range c.Jobs {
			matches := re.FindStringSubmatch(v)
			result = result + fmt.Sprintf("%ssource /opt/env.sh && %s\n", matches[1], matches[3])
		}
	}
	return result
}
