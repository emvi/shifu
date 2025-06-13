package cms

const (
	defaultPageContent = `<!DOCTYPE html>
		<html lang="en">
		<head>
			<base href="/" />
			<meta charset="UTF-8" />
			<meta name="viewport" content="width=device-width, initial-scale=1.0" />
			<meta name="copyright" content="Emvi Software GmbH" />
			<meta name="author" content="Emvi Software GmbH" />
			<meta name="title" content="Default Page" />
			<meta name="description" content="Default Shifu Page." />
			<title>Default Page</title>
			<link rel="stylesheet" type="text/css" href="/shifu-admin/static/admin.css" />
			<link rel="stylesheet" type="text/css" href="/shifu-admin/static/trix/trix.css" />
			<script defer src="/shifu-admin/static/trix/trix.min.js"></script>
			<script defer src="/shifu-admin/static/htmx.min.js"></script>
			<script defer src="/shifu-admin/static/htmx-ext-response-targets.min.js"></script>
			<script defer src="/shifu-admin/static/admin.js"></script>
		</head>
		<body>
			%s
		</body>
		</html>`
)
