import sys
import re
import matplotlib.pyplot as plt
from collections import defaultdict
import os


def parse_time_to_seconds(timestr: str) -> float:
    timestr = timestr.strip()
    match = re.match(r"([\d\.]+)([a-zµ]+)", timestr)
    if not match:
        return 0.0

    value_str, unit = match.groups()
    value = float(value_str)
    if unit == "ns":
        return value * 1e-9
    elif unit == "µs":
        return value * 1e-6
    elif unit == "ms":
        return value * 1e-3
    elif unit == "s":
        return value
    else:
        return 0.0


def main():
    if len(sys.argv) < 2:
        print(f"Usage: {sys.argv[0]} <csv_file>")
        sys.exit(1)

    csv_file = sys.argv[1]
    if not os.path.isfile(csv_file):
        print(f"Error: file '{csv_file}' not found.")
        sys.exit(1)

    data = defaultdict(list)  # key = OPS, value = list of (total_sec, avg_sec)

    # Read and parse the CSV
    with open(csv_file, "r") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            parts = line.split("||")
            if len(parts) < 4:
                continue

            # Extract OPS
            ops_str = parts[1].strip().split()[0]
            ops = int(ops_str)

            # Extract TOTAL time
            total_str = parts[2].split(":")[1].strip()
            total_sec = parse_time_to_seconds(total_str)

            # Extract AVERAGE time
            avg_str = parts[3].split(":")[1].strip()
            avg_sec = parse_time_to_seconds(avg_str)

            data[ops].append((total_sec, avg_sec))

    # Aggregate (average) times for each OPS
    ops_list = []
    total_avg_list = []
    avg_op_list = []

    for ops, pairs in sorted(data.items()):
        total_sum = sum(t for (t, _) in pairs)
        avg_sum = sum(a for (_, a) in pairs)
        count = len(pairs)

        ops_list.append(ops)
        total_avg_list.append(total_sum / count)
        avg_op_list.append(avg_sum / count)

    plt.figure(figsize=(8, 5))
    plt.plot(ops_list, total_avg_list, marker='o')
    plt.title("Total Time vs. OPS")
    plt.xlabel("OPS (iterations)")
    plt.ylabel("Total Time (seconds)")
    plt.grid(True)
    plt.tight_layout()
    plt.savefig(f"{csv_file}_total.png")
    plt.close()

    plt.figure(figsize=(8, 5))
    plt.plot(ops_list, avg_op_list, marker='o', color='orange')
    plt.title("Average Time per Operation vs. OPS")
    plt.xlabel("OPS (iterations)")
    plt.ylabel("Time per Operation (seconds)")
    plt.grid(True)
    plt.tight_layout()
    plt.savefig(f"{csv_file}_average.png")
    plt.close()


if __name__ == "__main__":
    main()
