./build.sh

./run_manual_allocator_benchmark.sh
./run_standard_allocator_benchmark.sh

python3 generate_benchmark_charts.py manual_bench_results.csv
python3 generate_benchmark_charts.py standard_bench_results.csv
