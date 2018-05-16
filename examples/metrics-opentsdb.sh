cpu () {
	echo $(hostname).cpu $(date +%s) $(ps axo %cpu | awk '{ sum+=$1 } END { printf "%.1f\n", sum }' | tail -n 1) host=$(hostname)
}

disk_io () {
	iostat=($(iostat -K -d -c 1 disk0))
	echo $(hostname).disk_io $(date +%s) "${iostat[6]}" host=$(hostname)
}

disk_usage () {
	echo $(hostname).disk_usage $(date +%s) $(df | awk -v disk_regexp="^/dev/disk1" \
                                '$0 ~ disk_regexp { printf "%d", $5/10 }') host=$(hostname)
}

heartbeat () {
	echo $(hostname).heartbeat $(date +%s) 1 host=$(hostname)
}

memory () {
	readonly __memory_os_memsize=$(sysctl -n hw.memsize)
	echo $(hostname).memory $(date +%s) $(vm_stat | awk -v total_memory=$__memory_os_memsize \
            'BEGIN { FS="   *"; pages=0 }
            /Pages (free|inactive|speculative)/ { pages+=$2 }
            END { printf "%.1f", 100 - (pages * 4096) / total_memory * 100.0 }') host=$(hostname)
}

network_io () {
	sample=$(netstat -b -I en0 | awk '{ print $7" "$10 }' | tail -n 1)
	calc_kBps () {
		echo $1 $2 | awk -v divisor=1024 '{ printf "%.2f", ($1 - $2) / divisor }'
	}
	echo $(hostname).network_in $(date +%s) $(calc_kBps $(echo $sample | awk '{print $1}') $(echo $previous_sample | awk '{print $1}')) host=$(hostname)
    echo $(hostname).network_out $(date +%s) $(calc_kBps $(echo $sample | awk '{print $2}') $(echo $previous_sample | awk '{print $2}')) host=$(hostname)
}

ping_ok () {
	ping -c 1 $PING_REMOTE_HOST > /dev/null 2>&1
  	if [ $? -eq 0 ]; then
    	echo $(hostname).ping_ok $(date +%s) 1 host=$(hostname)
  	else
    	echo $(hostname).ping_ok $(date +%s) 0 host=$(hostname)
  	fi
}

cpu
disk_io
disk_usage
heartbeat
memory
network_io
ping_ok
