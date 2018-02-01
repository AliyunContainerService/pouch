if [ $1 -ge 0 ] ; then
	# check if pouch is running before upgrade
	if [ -f /var/lib/rpm-state/pouch-is-active ] ; then
		systemctl start pouch > /dev/null 2>&1
		rm -f /var/lib/rpm-state/pouch-is-active > /dev/null 2>&1
	fi
fi

