if [ $1 -eq 0 ] ; then
	# Package removal
	systemctl --no-reload disable pouch > /dev/null 2>&1
	systemctl stop pouch > /dev/null 2>&1
fi
