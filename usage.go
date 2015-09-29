package main

// Changes:
// - removed '[argument...]' at the end of USAGE line
// - changed '[global options]' with '[options]'
// - changed 'GLOBAL OPTIONS' with 'OPTIONS'
// - removed 'COMMANDS' section entirely (was showing only 'help,h' anyway)
// - removed '{{if .Commands}} command [command options]{{end}}' from the line after 'USAGE'
const usageTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.Name}} {{if .Flags}}[options]{{end}}
   {{if .Version}}
VERSION:
   {{.Version}}
   {{end}}{{if len .Authors}}
AUTHOR(S):
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Flags}}
OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
COPYRIGHT:
   {{.Copyright}}
   {{end}}
`
