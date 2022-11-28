package statuspages

import "html/template"

var (
	// Begin is the start of status pages, End is the end.
	Begin        *template.Template
	serviceEntry *template.Template

	// For base service:
	mainPageTemplate    *template.Template
	servicePageTemplate *template.Template
)

func init() {
	var err error
	if Begin, err = template.New("Begin").Parse(`<!DOCTYPE html>
<html>
<head>
	<title>{{.}} status page</title>
	<style>
		.expandable {
			display: none;
		}
	</style>
</head>
<body>
	<p><a href='./status'>Main status page</a></p>
    <h1>{{.}} Status Page</h1>`); err != nil {
		panic("Failed to parse status page Begin template")
	}

	if serviceEntry, err = template.New("ServiceEntry").Parse(
		"<h1><a href=\"./status?service={{.}}\">{{.}}</a></h1>\n"); err != nil {
		panic("Failed to parse status page serviceEntry template")
	}

	mainPageTemplate, err = template.New("BaseServiceMainPage").Parse(`
<ul>
	<li>Process name: {{.ProcessName}}</li>
	<li>Current time: {{.Now.Format "15:04:05 02/01/2006"}}</li>
	<li>Start time: {{.StartTime.Format "15:04:05 02/01/2006"}}, {{.StartTimeAgo}}</li>
</ul>
`)
	if err != nil {
		panic("Failed to parse status page base service mainPage template")
	}
	servicePageTemplate, err = template.New("BaseServiceServicePage").Parse(`
<h1>Command line:</h1>
<code>
	{{range .CommandLine}}{{.}} {{end}}
</code>

<h1>Environment:</h1>
<ul>
	{{range .Env}}
		<li>{{.}}</li>
	{{end}}
</ul>
`)
	if err != nil {
		panic("Failed to parse status page base service mainPage template")
	}

}

const End = `<script>
document.getElementsByClassName("expandButton").forEach(func(item) {
	item.addEventListener("click", function () {
		var content = this.nextElementSibling;
		if (content.style.display === "block") {
			content.style.display = "none";
		} else {
			content.style.display = "block";
		}
	});
</script>
</body>
</html>
`
