package pkg

import (
	"encoding/json"
	"errors"
	"github.com/emvi/shifu/pkg/cfg"
	"os"
	"path/filepath"
)

const (
	mainScss = `body {
	max-width: 800px;
	margin: 40px;
}
`
	mainJs = `console.log("Hello from Shifu!");
`
	homeJson = `{
    "path": {
        "en": "/"
    },
    "sitemap": {
        "priority": "1.0"
    },
    "content": {
        "content": [
            {
                "tpl": "main",
                "copy": {
                    "en": {
						"title": "Shifu Website",
                        "headline": "Welcome to Shifu!"
                    }
                }
            }
        ]
    }
}
`
	mainHtml = `<!DOCTYPE html>
<html lang="en">
<head>
    <base href="/" />
    <meta charset="UTF-8" />
    <link rel="stylesheet" type="text/css" href="static/main.css" />
    <script type="text/javascript" src="static/main.min.js"></script>
    <title>{{copy .Page .Content "title"}}</title>
</head>
<body>
    <h1>{{copy .Page .Content "headline"}}</h1>
</body>
</html>
`
	gitIgnore = `.secrets.env
shifu.db
`
)

var (
	dirs = []string{
		"assets",
		"assets/scss",
		"assets/js",
		"content",
		"static",
		"tpl",
	}
)

// Init creates a new Shifu project directory structure in given directory.
func Init(path string) error {
	s, err := os.Stat(path)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0744); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	if s != nil && !s.IsDir() {
		return errors.New("target path is not a directory")
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(path, dir), 0744); err != nil {
			return err
		}
	}

	cfgFile, err := os.OpenFile(filepath.Join(path, "config.json"), os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	defer cfgFile.Close()
	configJson, err := json.MarshalIndent(cfg.Config{
		Dev: true,
		Server: cfg.Server{
			Host:            "localhost",
			Port:            8080,
			ShutdownTimeout: 30,
			WriteTimeout:    5,
			ReadTimeout:     5,
		},
		Sass: cfg.Sass{
			Entrypoint:   "main.scss",
			Dir:          "assets/scss",
			Watch:        true,
			Out:          "static/main.css",
			OutSourceMap: "static/main.css.map",
		},
		JS: cfg.JS{
			Entrypoint: "main.js",
			Dir:        "assets/js",
			Watch:      true,
			Out:        "static/main.min.js",
			SourceMap:  true,
		},
	}, "", "    ")

	if err != nil {
		return err
	}

	if _, err := cfgFile.Write(configJson); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(path, "assets", "scss", "main.scss"), []byte(mainScss), 0644); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(path, "assets", "js", "main.js"), []byte(mainJs), 0644); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(path, "content", "home.json"), []byte(homeJson), 0644); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(path, "tpl", "main.html"), []byte(mainHtml), 0644); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(path, ".secrets.env"), []byte{}, 0644); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(path, ".gitignore"), []byte(gitIgnore), 0644); err != nil {
		return err
	}

	return nil
}
