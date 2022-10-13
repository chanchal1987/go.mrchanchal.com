package main

import (
	"context"
	_ "embed"
	"html/template"
	"os"
	"path"

	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

//go:embed package.tpl.html
var pkgRaw string

//go:embed index.tpl.html
var idxRaw string

func main() {
	var conf struct {
		Domain     string
		Index      string
		GithubUser string `yaml:"github-user"`
		Target     string
		Packages   []*struct {
			Name string
		}
	}

	// read config
	if data, err := os.ReadFile("config.yaml"); err != nil {
		panic(err)
	} else {
		if err := yaml.Unmarshal(data, &conf); err != nil {
			panic(err)
		}
	}

	// create target
	if err := os.MkdirAll(conf.Target, 0o750); err != nil {
		panic(err)
	}

	tpl := template.Must(template.New("pkg").Parse(pkgRaw))

	type tplData struct {
		Domain     string
		GithubUser string
		Package    string
	}

	errGrp, _ := errgroup.WithContext(context.Background())

	for _, p := range conf.Packages {
		d := tplData{
			Domain:     conf.Domain,
			GithubUser: conf.GithubUser,
			Package:    p.Name,
		}

		errGrp.Go(func() error {
			dir := path.Join(conf.Target, d.Package)

			// create pkg
			err := os.MkdirAll(dir, 0o750)
			if err != nil {
				return err
			}

			// write idx
			f, err := os.Create(path.Join(dir, conf.Index))
			if err != nil {
				return err
			}

			defer f.Close()

			return tpl.Execute(f, d)
		})

	}

	if err := errGrp.Wait(); err != nil {
		panic(err)
	}

	tpl = template.Must(template.New("index").Parse(idxRaw))
	f, err := os.Create(path.Join(conf.Target, conf.Index))
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if err := tpl.Execute(f, conf.Packages); err != nil {
		panic(err)
	}
}
