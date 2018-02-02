systemctl daemon-reload > /dev/null 2>&1
if [ $1 -ge 1 ] ; then
	systemctl try-restart pouch > /dev/null 2>&1
fi
