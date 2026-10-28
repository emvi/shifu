package pkg

import (
	"errors"
	"os"
	"path/filepath"
)

var starterFiles = map[string]string{
	// Configuration and gitignore. ADMIN_PASSWORD is only used to create the user "admin" on the first start
	// (when shifu.db is created), so it must be changed in the admin UI before going live.
	"config.json": `{
    "dev": ${DEV},
    "log_level": "info",
    "server": {
        "host": "${HOST}",
        "port": ${PORT},
        "secure_cookies": ${SECURE_COOKIES},
        "hostname": "${HOSTNAME}"
    },
    "content": {
        "provider": "fs",
        "not_found": {
            "en": "/not-found"
        }
    },
    "sass": {
        "dir": "assets/scss",
        "entrypoint": "main.scss",
        "out": "static/main.css",
        "out_source_map": "static/main.css.map",
        "watch": true,
        "compiler": "sass"
    },
    "js": {
        "dir": "assets/js",
        "entrypoint": "main.js",
        "out": "static/main.js",
        "source_map": true,
        "watch": true
    },
    "ui": {
        "path": "/admin",
        "admin_password": "${ADMIN_PASSWORD}"
    }
}
`,
	".secrets.env": `DEV=true
HOST=localhost
PORT=8080
SECURE_COOKIES=false
HOSTNAME=http://localhost:8080
ADMIN_PASSWORD=admin
`,
	".gitignore": `.secrets.env
shifu.db
static/main.css
static/main.css.map
static/main.js
static/main.js.map
`,

	// templates
	"tpl/layout/head.html": `{{- $title := copy .Page .Content "title" -}}
{{- $description := copy .Page .Content "meta_description" -}}
<!DOCTYPE html>
<html lang="{{ .Page.Language }}">
<head>
    <base href="/" />
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>{{ $title }}</title>
    <meta name="description" content="{{ $description }}" />
    <link rel="canonical" href="{{ .Page.CanonicalLink }}" />
    {{- range $language, $path := .Page.Path }}
    <link rel="alternate" hreflang="{{ $language }}" href="{{ hostname }}{{ $path }}" />
    {{- end }}
    <link rel="stylesheet" href="/static/main.css?v={{ .CMS.LastUpdate }}" />
    <script defer src="/static/main.js?v={{ .CMS.LastUpdate }}"></script>
    {{ adminHead .Page }}
</head>
<body>
`,
	"tpl/layout/end.html": `{{ adminBody .Page }}
</body>
</html>
`,
	"tpl/layout/header.html": `{{- $home := copy .Page .Content "home_label" -}}

<header class="header" data-component="nav" data-shifu-element="{{ .Content.Position }}">
    <div class="header__inner">
        <a class="header__home" href="/">{{ default $home "Home" }}</a>
        <button class="header__toggle" type="button" aria-expanded="false" aria-label="Menu">&#9776;</button>
        <nav class="header__nav">
            {{ .CMS.Render .Args .Page (fmt "%s.nav" .Content.Position) (index .Content.Content "nav") }}

            {{ if loggedIn .Page }}
                <slot name="{{ .Content.Position }}/nav"></slot>
            {{ end }}
        </nav>
    </div>
</header>
`,
	"tpl/layout/footer.html": `{{- $text := copy .Page .Content "text" -}}

<footer class="footer" data-shifu-element="{{ .Content.Position }}">
    <div class="footer__inner">
        {{ with $text }}<div class="footer__text">{{ html . }}</div>{{ end }}
        <small class="footer__copyright">&copy; {{ year }}</small>
    </div>
</footer>
`,
	"tpl/components/nav_item.html": `{{- $label := copy .Page .Content "label" -}}
{{- $href := copy .Page .Content "href" -}}

<a class="nav-item" href="{{ default $href "/" }}" data-shifu-element="{{ .Content.Position }}">{{ default $label "Link" }}</a>
`,
	"tpl/sections/text.html": `{{- $headline := copy .Page .Content "headline" -}}
{{- $text := copy .Page .Content "text" -}}
{{- $background := get .Content "background" -}}

<section class="text{{ with $background }} text--{{ . }}{{ end }}" data-shifu-element="{{ .Content.Position }}">
    <div class="text__inner">
        {{ with $headline }}<h2 class="text__headline">{{ . }}</h2>{{ end }}
        {{ with $text }}<div class="text__body">{{ html . }}</div>{{ end }}
    </div>
</section>
`,

	// admin configuration
	"admin/tpl/layout/head.json": `{
    "label": "_HTML Head",
    "layout": true,
    "copy": {
        "title": { "label": "Title" },
        "meta_description": { "label": "Meta Description" }
    }
}
`,
	"admin/tpl/layout/end.json": `{
    "label": "_HTML End",
    "layout": true
}
`,
	"admin/tpl/layout/header.json": `{
    "label": "_Header",
    "layout": true,
    "content": {
        "nav": {
            "label": "Navigation",
            "tpl_filter": ["nav_item"]
        }
    },
    "copy": {
        "home_label": { "label": "Home Link Label" }
    }
}
`,
	"admin/tpl/layout/footer.json": `{
    "label": "_Footer",
    "layout": true,
    "copy": {
        "text": { "label": "Text", "type": "html" }
    }
}
`,
	"admin/tpl/components/nav_item.json": `{
    "label": "Navigation Item",
    "copy": {
        "label": { "label": "Label" },
        "href": { "label": "Link" }
    }
}
`,
	"admin/tpl/sections/text.json": `{
    "label": "Text",
    "copy": {
        "headline": { "label": "Headline" },
        "text": { "label": "Text", "type": "html" }
    },
    "data": {
        "background": {
            "label": "Background",
            "type": "select",
            "options": {
                "": "Default",
                "light": "Light",
                "dark": "Dark"
            }
        }
    }
}
`,

	// references
	"content/refs/head.json": `{
    "tpl": "head",
    "copy": {
        "en": { "title": "Shifu Website", "meta_description": "A website built with Shifu." }
    }
}
`,
	"content/refs/end.json": `{
    "tpl": "end"
}
`,
	"content/refs/header.json": `{
    "tpl": "header",
    "copy": {
        "en": { "home_label": "Shifu" }
    },
    "content": {
        "nav": [
            {
                "tpl": "nav_item",
                "copy": { "en": { "label": "Home", "href": "/" } }
            }
        ]
    }
}
`,
	"content/refs/footer.json": `{
    "tpl": "footer",
    "copy": {
        "en": { "text": "<div>Built with <a href=\"https://github.com/emvi/shifu\">Shifu</a>.</div>" }
    }
}
`,

	// pages
	"content/pages/home.json": `{
    "display_name": "Home",
    "path": {
        "en": "/"
    },
    "sitemap": { "priority": "1.0" },
    "content": {
        "content": [
            {
                "ref": "head",
                "copy": { "en": { "title": "Shifu Website" } }
            },
            { "ref": "header" },
            {
                "tpl": "text",
                "copy": {
                    "en": {
                        "headline": "Welcome to Shifu!",
                        "text": "<div>Sign in at <a href=\"/admin\">/admin</a> to edit this page.</div>"
                    }
                },
                "data": { "background": "light" }
            },
            { "ref": "footer" },
            { "ref": "end" }
        ]
    }
}
`,
	"content/pages/not-found.json": `{
    "display_name": "Not Found",
    "path": {
        "en": "/not-found"
    },
    "content": {
        "content": [
            {
                "ref": "head",
                "copy": { "en": { "title": "Page not found" } }
            },
            { "ref": "header" },
            {
                "tpl": "text",
                "copy": {
                    "en": {
                        "headline": "Page not found",
                        "text": "<div>The page you are looking for does not exist. <a href=\"/\">Go to the home page</a>.</div>"
                    }
                }
            },
            { "ref": "footer" },
            { "ref": "end" }
        ]
    }
}
`,

	// styles
	"assets/scss/main.scss": `@use "variables";
@use "base";

@use "header";
@use "nav_item";
@use "footer";
@use "text";
`,
	"assets/scss/_variables.scss": `$color-text: #1d1d1f;
$color-muted: #6e6e73;
$color-bg: #ffffff;
$color-light: #f5f5f7;
$color-dark: #1d1d1f;
$color-accent: #0a64d6;
$max-width: 960px;
$breakpoint-md: 768px;

@mixin md {
    @media (min-width: $breakpoint-md) {
        @content;
    }
}

@mixin inner {
    max-width: $max-width;
    margin: 0 auto;
    padding: 0 20px;
}
`,
	"assets/scss/_base.scss": `@use "variables" as *;

*,
*::before,
*::after {
    box-sizing: border-box;
}

body {
    margin: 0;
    font-family: system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
    line-height: 1.6;
    color: $color-text;
    background: $color-bg;
}

a {
    color: $color-accent;
}

img {
    max-width: 100%;
    height: auto;
}
`,
	"assets/scss/_header.scss": `@use "variables" as *;

.header {
    border-bottom: 1px solid $color-light;

    &__inner {
        @include inner;
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        justify-content: space-between;
        min-height: 64px;
    }

    &__home {
        font-weight: 600;
        color: $color-text;
        text-decoration: none;
    }

    &__toggle {
        padding: 4px 8px;
        font-size: 20px;
        background: none;
        border: 0;
        cursor: pointer;

        @include md {
            display: none;
        }
    }

    &__nav {
        display: none;
        flex-basis: 100%;
        flex-direction: column;
        gap: 8px;
        padding: 0 0 16px;

        @include md {
            display: flex;
            flex-basis: auto;
            flex-direction: row;
            gap: 24px;
            padding: 0;
        }
    }

    &--open &__nav {
        display: flex;
    }
}
`,
	"assets/scss/_nav_item.scss": `@use "variables" as *;

.nav-item {
    color: $color-text;
    text-decoration: none;

    &:hover {
        color: $color-accent;
    }
}
`,
	"assets/scss/_footer.scss": `@use "variables" as *;

.footer {
    padding: 32px 0;
    color: $color-muted;
    border-top: 1px solid $color-light;

    &__inner {
        @include inner;
    }
}
`,
	"assets/scss/_text.scss": `@use "variables" as *;

.text {
    padding: 64px 0;

    &--light {
        background: $color-light;
    }

    &--dark {
        color: $color-bg;
        background: $color-dark;

        a {
            color: inherit;
        }
    }

    &__inner {
        @include inner;
    }

    &__headline {
        margin: 0 0 16px;
        font-size: 32px;
        line-height: 1.2;
    }

    &__body {
        p, ul, ol, blockquote, pre {
            margin: 0 0 16px;
        }
    }
}
`,

	// scripts
	"assets/js/main.js": `import { initNav } from "./nav.js";

const components = {
    nav: initNav,
};

function init(root = document) {
    root.querySelectorAll("[data-component]").forEach((el) => {
        if (el.dataset.componentReady) {
            return;
        }

        const fn = components[el.dataset.component];

        if (fn) {
            fn(el);
            el.dataset.componentReady = "true";
        }
    });
}

document.addEventListener("DOMContentLoaded", () => {
    init();
    // the admin UI swaps element HTML in place after edits
    new MutationObserver(() => init()).observe(document.body, { childList: true, subtree: true });
});
`,
	"assets/js/nav.js": `export function initNav(el) {
    const toggle = el.querySelector(".header__toggle");

    if (!toggle) {
        return;
    }

    toggle.addEventListener("click", () => {
        const open = el.classList.toggle("header--open");
        toggle.setAttribute("aria-expanded", String(open));
    });
}
`,

	// directories that must exist but start empty
	"static/media/.gitkeep": "",
	"tpl/partials/.gitkeep": "",
}

// Init creates a new Shifu project directory structure in given directory.
// Existing files are not overwritten.
func Init(path string) error {
	s, err := os.Stat(path)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	if s != nil && !s.IsDir() {
		return errors.New("target path is not a directory")
	}

	for name, content := range starterFiles {
		file := filepath.Join(path, filepath.FromSlash(name))

		if _, err := os.Stat(file); err == nil {
			continue
		}

		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}
