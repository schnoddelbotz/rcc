package resources

import _ "embed"

//go:embed chart.js
var ChartJS string

//go:embed chart.html
var ChartHTML string

//go:embed gnuplot.tpl
var GnuplotScriptTemplate string
