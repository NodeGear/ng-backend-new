#!/bin/bash

# $1 home location
# $2 process location
# $3 Git URL
# $4 Git Branch
# $5 Use snapshot (0/1)
# $6 Snapshot Location

# Exit codes:
# 0 - OK
# 1 - /home/$1/$2 exists
# 2 - Git not valid
# 3 - Could not find Branch
# 5 - Other error

#source ~/.bashrc

#PATH=$PATH:/usr/local/bin

set -o errexit

printf "#\041/bin/bash\nssh -i ${1}/.ssh/id_rsa \$1 \$2\n" > $1/ssh_wrapper.sh
chmod +x $1/ssh_wrapper.sh

GIT_SSH=$1/ssh_wrapper.sh git clone "$3" $2
if [ $? -ne 0 ]; then
	exit 2
fi

cd $2

git checkout "$4"
if [ $? -ne 0 ]; then
	exit 3
fi

if [ $5 -eq 1 ]; then
	#set +o errexit
	echo $6
	git apply $6
	#set -o errexit

	rm -f $6
fi

exit 0
