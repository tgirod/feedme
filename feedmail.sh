#!/bin/zsh
FROM=feedme@localhost
TO=your@address

CONTENT=$(feedme fetch)
if [[ $CONTENT != "" ]]
then
	COUNT=$(echo $CONTENT | grep "^ - " | wc -l)
	echo "From:$FROM\nSubject:$COUNT new post(s)\n\n$CONTENT" | sendmail $TO
fi

