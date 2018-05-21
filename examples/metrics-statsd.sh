cpu () {
	cpu=$(ps axo %cpu | awk '{ sum+=$1 } END { printf "%.1f\n", sum }' | tail -n 1)
}

disk_io () {
	iostat=($(iostat -K -d -c 1 disk0))
	disk_io="${iostat[6]}"
}

disk_usage () {
	disk_usage=$(df | awk -v disk_regexp="^/dev/disk1" '$0 ~ disk_regexp { printf "%d", $5/10 }')
}

heartbeat () {
	heartbeat=1
}

memory () {
	readonly __memory_os_memsize=$(sysctl -n hw.memsize)
	memory=$(vm_stat | awk -v total_memory=$__memory_os_memsize \
            'BEGIN { FS="   *"; pages=0 }
            /Pages (free|inactive|speculative)/ { pages+=$2 }
            END { printf "%.1f", 100 - (pages * 4096) / total_memory * 100.0 }')
}

network_io () {
	sample=$(netstat -b -I en0 | awk '{ print $7" "$10 }' | tail -n 1)
	calc_kBps () {
		echo $1 $2 | awk -v divisor=1024 '{ printf "%.2f", ($1 - $2) / divisor }'
	}
	network_in=$(calc_kBps $(echo $sample | awk '{print $1}') $(echo $previous_sample | awk '{print $1}'))
  network_out=$(calc_kBps $(echo $sample | awk '{print $2}') $(echo $previous_sample | awk '{print $2}'))
}

ping_ok () {
	ping -c 1 $PING_REMOTE_HOST > /dev/null 2>&1
  	if [ $? -eq 0 ]; then
    	ping_ok=1
  	else
    	ping_ok=0
  	fi
}

cpu
disk_io
disk_usage
heartbeat
memory
network_io
ping_ok

echo $(hostname).cpu:$cpu\|c | nc -u -w0 127.0.0.1 8125
echo $(hostname).disk_io:$disk_io\|c | nc -u -w0 127.0.0.1 8125
echo $(hostname).disk_usage:$disk_usage\|c | nc -u -w0 127.0.0.1 8125
echo $(hostname).heartbeat:$heartbeat\|c | nc -u -w0 127.0.0.1 8125
echo $(hostname).memory:$memory\|c | nc -u -w0 127.0.0.1 8125
echo $(hostname).network_in:$network_in\|c | nc -u -w0 127.0.0.1 8125
echo $(hostname).network_out:$network_out\|c | nc -u -w0 127.0.0.1 8125
echo $(hostname).ping_ok:$ping_ok\|c | nc -u -w0 127.0.0.1 8125
