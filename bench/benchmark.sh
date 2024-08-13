#!/bin/bash
set -euo pipefail

go test -bench=. -run='^$' -v -count=10 ./... | tee benchmarks.out
benchstat -col /impl benchmarks.out | tee benchstat.out

mac_hardware() {
  JS=$(system_profiler -json SPHardwareDataType)
  local hw=$(echo $JS | jq -r .SPHardwareDataType[0].machine_name)
  local id=$(echo $JS | jq -r .SPHardwareDataType[0].machine_model)
  local mn=$(echo $JS | jq -r .SPHardwareDataType[0].model_number)
  local cpu=$(echo $JS | jq -r .SPHardwareDataType[0].chip_type)
  local cores=$(echo $JS | jq -r .SPHardwareDataType[0].number_processors)
  local mem=$(echo $JS | jq -r .SPHardwareDataType[0].physical_memory)
  cat << EOF
| Hardware |   ID  | Model # | CPU    | Cores    | Memory |
|----------|-------|---------|--------|----------|--------|
| ${hw}    | ${id} | ${mn}   | ${cpu} | ${cores} | ${mem} |

EOF
}

hardware() {
  if command -v system_profiler > /dev/null; then
    mac_hardware
  fi
}

{
  cat template/README.head.md
  echo '## Hardware'
  echo ''
  hardware

  echo '## Results'
  echo ''
  echo '```'
  cat benchstat.out
  echo '```'
} > README.md



