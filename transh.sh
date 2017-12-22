# add the following code to your .zshrc or .bashrc, you can change email and pass by yourself.
transfer() {
    if [ $# -eq 0 ]; then
        echo "No arguments specified. Usage:\ntransfer /tmp/test.md\ncat /tmp/test.md | transfer test.md"
        return 1
    fi 
    tmpfile=$( mktemp -t transferXXX )
    if tty -s; then 
        TS=$(date +%Y%m%d%H%M%S)
        basefile=$(basename "$1" | sed -e 's/[^a-zA-Z0-9._-]/-/g')
        curl -u 'user1':'pass1' \
             --progress-bar --upload-file "$1" "http://127.0.0.1:8000/$basefile-$TS" >> $tmpfile
    else
        curl -u 'user1':'pass1' \
             --progress-bar --upload-file "-" "http://127.0.0.1:8000/$1-$TS" >> $tmpfile
    fi
    cat $tmpfile
    rm -f $tmpfile
}
transfer-del() {
    if [ $# -eq 0 ]; then
        echo "No arguments specified. Usage:\ntransfer-del http://xxxx/tmp/test.md\n"
        return 1
    fi
    if tty -s; then
        curl -u 'user1':'pass1' -X DELETE $1
    fi
}
