package resources

import _ "embed"

//go:embed chart.js
var ChartJS string

//go:embed chart.html
var ChartHTML string

//go:embed gnuplot.tpl
var GnuplotScriptTemplate string

//go:embed chartjs-adapter-moment.js
var ChartJSAdapterMoment string

//go:embed moment.js
var MomentJS string
