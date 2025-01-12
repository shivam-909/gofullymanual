rm -f manual_bench_results.csv

for i in $(seq 1000 20000 1000000)
do
  for run in $(seq 1 100)
  do
    GOGC=off ./manual_test "$i" >> manual_bench_results.csv
  done
done
