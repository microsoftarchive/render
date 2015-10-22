**This is deprecated. Please do not use this repo.**


# Render

does a couple of things post provisioning:

* clone repos
* render runit
* render crontab
* render api.json
* render database.yml
* render consul environment

```sh
$ rm -rf tmp
$ PROJECT_PATH="$HOME/src/github.com/wunderlist" go run main.go manifest.go runit.go crontab.go database.go -type web -app rabbitmqwatcher -rev c9f45e8
$ ls tmp
bin crontab env.sh repos sv
```
