set title '{{title}}'
set output "{{outfile}}"
set xlabel 'Date'
set timefmt "%Y-%m-%dT%H:%M:%S+02:00"
set key left top
set xdata time
set ytics 500 nomirror
set ylabel 'Lines of Code'
set y2tics 10 nomirror
set y2label '{{y2Label}}' {{y2RangeIfAny}}
set term pngcairo
set terminal png size 1400,700
plot {{plotArgs}}
