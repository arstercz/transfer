# transfer
Easy and fast file sharing from the command-line

fork from https://github.com/dutchcoders/transfer.sh, but remove aws and virustotal features, and add the following features:
```
1. use custume http header to verifid users;
2. transfer read configure file to revify user;
3. add timestamp to the upload file;
4. use http delete method to delete file;
```

### How to build
```
go get github.com/chenzhe07/transfer
cd $GOPATH/src/github.com/chenzhe07/transfer
go build -o transfer *.go
```

### Run with conf file
```
./transfer -log /tmp/trans.log -port 8000 -temp /data/transfer/logs --basedir /data/transfer/logs -conf /etc/user.conf
```

### config file example
the following config file is equal to curl -H "X-Alter-Email: chenzhe07@gmail.com" -H "X-Alter-Pass: 827(ISJhs7s" https://xxxx/xxxx`
```
$ cat /etc/user.conf 
[chenzhe07@gmail.com]
pass = 827(ISJhs7s
```

### How to transfer and delete file

add `transh.sh` content to your .zshrc or .bashrc, and then `source .zshrc` or `source .bashrc`

#### 1. transfer file

```
$ transfer mysql.3306.txt 
######################################################################## 100.0%
http://127.0.0.1:8000/9b09dhj/mysql.3306.txt-20161202154850
```

#### 2. delete file

```
$ transfer-del http://127.0.0.1:8000/9b09dhj/mysql.3306.txt-20161202154850
delete 9b09dhj/mysql.3306.txt-20161202154850 ok
```

#### 3. transfer with wrong http header pass:

http 500 returned when there is wrong email and pass
```
$  transfer mysql.3306.txt 
                                                                           0.0%
Verify user and pass error
```
