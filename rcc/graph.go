package rcc

func createGraphGnuplot(graphData any, outfile string) error {
	return nil
}

/*

function plot_data() {
        gnuplot <<EOF
        set title 'Zwo LoC / Coverage'
        set xlabel 'Date'
        set timefmt "%Y-%m-%dT%H:%M:%S+02:00"
        set xdata time
        set ytics 500 nomirror
        set ylabel 'LoC'
        set y2tics 10 nomirror
        set y2label 'Coverage'
        set y2range [0:100]
        set term pngcairo
        set terminal png size 1024,768
        set output "code_stats.png"
        plot 'code_stats.dat' using 1:2 t 'Go, including tests' with linespoints, \
             'code_stats.dat' using 1:3 t 'Go, excluding tests' with linespoints, \
             'code_stats.dat' using 1:4 t 'HTML' with linespoints, \
             'code_stats.dat' using 1:5 t 'JS' with linespoints, \
             'code_stats.dat' using 1:6 t 'PS1' with linespoints, \
             'code_stats.dat' using 1:7 t 'Coverage' axis x1y2 with linespoints
EOF
}

2025-07-10T13:54:29+02:00 4204 3856 868 910 1590 10.9
2025-07-10T13:53:26+02:00 4205 3857 868 910 1590 12.4
2025-07-10T13:48:30+02:00 4205 3857 868 910 1585 12.4
2025-07-10T13:10:02+02:00 4218 3870 868 910 1587 10.9
2025-07-10T11:43:40+02:00 4218 3870 868 910 1618 10.9
2025-07-10T11:28:25+02:00 4215 3867 868 910 1618 12.4
2025-07-10T11:25:32+02:00 4216 3868 868 910 1618 10.9

*/
