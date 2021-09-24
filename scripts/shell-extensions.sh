run () {
	for dir in script scripts
	do
		for suffix in "" ".sh" ".go"
		do
			cmd=$(git rev-parse --show-toplevel)/$dir/$1$suffix
			if test -f $cmd
			then
				if [[ "${cmd: -3}" == ".go" ]]
				then
					go run $cmd ${@:2}
				else
					$cmd ${@:2}
				fi
				return $?
			fi
		done
	done
	echo "no command $1 found"
}
