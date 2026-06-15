package collector

// CollectScript 远程采集脚本。
// 设计要点（PRD 5.4）：单次 SSH 连接执行，输出一行 JSON，包含 CPU/内存/负载/磁盘/系统信息。
// 使用 POSIX 工具（awk/df/hostname/grep/nproc），兼容主流 Linux 发行版。
// 每个字段失败均回退默认值，保证最终输出是合法 JSON。
const CollectScript = `hn=$(hostname 2>/dev/null); [ -z "$hn" ] && hn="unknown"
os=$(grep '^PRETTY_NAME=' /etc/os-release 2>/dev/null | head -1 | sed 's/^PRETTY_NAME=//; s/^"//; s/"$//')
[ -z "$os" ] && os=$(uname -sr 2>/dev/null); [ -z "$os" ] && os="unknown"
uptime_s=$(awk '{printf "%d", $1}' /proc/uptime 2>/dev/null); [ -z "$uptime_s" ] && uptime_s=0
cores=$(nproc 2>/dev/null); [ -z "$cores" ] && cores=$(grep -c '^processor' /proc/cpuinfo 2>/dev/null); [ -z "$cores" ] && cores=0
s1=$(awk '/^cpu /{print $2,$3,$4,$5,$6,$7,$8}' /proc/stat 2>/dev/null)
sleep 1
s2=$(awk '/^cpu /{print $2,$3,$4,$5,$6,$7,$8}' /proc/stat 2>/dev/null)
cpu_usage=$(printf '%s %s\n' "$s1" "$s2" | awk '{t1=$1+$2+$3+$4+$5+$6+$7; t2=$8+$9+$10+$11+$12+$13+$14; id1=$4+$5; id2=$11+$12; dt=t2-t1; did=id2-id1; if(dt>0) printf "%.2f",(1-did/dt)*100; else printf "0"}')
[ -z "$cpu_usage" ] && cpu_usage=0
mem=$(awk '/MemTotal/{t=$2}/MemAvailable/{a=$2}END{if(t>0){u=t-a; printf "%d %d %d %.2f",t/1024,u/1024,a/1024,u/t*100}else print "0 0 0 0"}' /proc/meminfo 2>/dev/null)
[ -z "$mem" ] && mem="0 0 0 0"
mem_total=$(echo "$mem"|awk '{print $1}'); mem_used=$(echo "$mem"|awk '{print $2}'); mem_avail=$(echo "$mem"|awk '{print $3}'); mem_usage=$(echo "$mem"|awk '{print $4}')
load=$(awk '{print $1,$2,$3}' /proc/loadavg 2>/dev/null); [ -z "$load" ] && load="0 0 0"
l1=$(echo "$load"|awk '{print $1}'); l5=$(echo "$load"|awk '{print $2}'); l15=$(echo "$load"|awk '{print $3}')
disks=$(df -B1 -P 2>/dev/null | awk 'NR>1 && $2 ~ /^[0-9]+$/ && $1 !~ /^(tmpfs|devtmpfs|overlay|shm|udev|proc|sysfs|cgroup|none)/ {u=$5;gsub(/%/,"",u); if(seen)printf ","; printf "{\"filesystem\":\"%s\",\"mount_point\":\"%s\",\"size_bytes\":%s,\"used_bytes\":%s,\"available_bytes\":%s,\"usage_percent\":%s}",$1,$6,$2,$3,$4,u; seen=1}')
printf '{"hostname":"%s","os":"%s","uptime_seconds":%s,"cpu":{"usage_percent":%s,"cores":%s},"memory":{"total_mb":%s,"used_mb":%s,"available_mb":%s,"usage_percent":%s},"load":{"load1":%s,"load5":%s,"load15":%s},"disks":[%s]}' "$hn" "$os" "$uptime_s" "$cpu_usage" "$cores" "$mem_total" "$mem_used" "$mem_avail" "$mem_usage" "$l1" "$l5" "$l15" "$disks"`

// collectData 采集脚本输出的 JSON 结构（对应 PRD 5.4）
type collectData struct {
	Hostname      string `json:"hostname"`
	OS            string `json:"os"`
	UptimeSeconds int64  `json:"uptime_seconds"`
	CPU           struct {
		UsagePercent float64 `json:"usage_percent"`
		Cores        int     `json:"cores"`
	} `json:"cpu"`
	Memory struct {
		TotalMB      int64   `json:"total_mb"`
		UsedMB       int64   `json:"used_mb"`
		AvailableMB  int64   `json:"available_mb"`
		UsagePercent float64 `json:"usage_percent"`
	} `json:"memory"`
	Load struct {
		Load1  float64 `json:"load1"`
		Load5  float64 `json:"load5"`
		Load15 float64 `json:"load15"`
	} `json:"load"`
	Disks []diskData `json:"disks"`
}

type diskData struct {
	Filesystem     string  `json:"filesystem"`
	MountPoint     string  `json:"mount_point"`
	SizeBytes      int64   `json:"size_bytes"`
	UsedBytes      int64   `json:"used_bytes"`
	AvailableBytes int64   `json:"available_bytes"`
	UsagePercent   float64 `json:"usage_percent"`
}
