# transfer
Easy and fast file sharing from the command-line

fork from https://github.com/dutchcoders/transfer.sh, but remove aws and virustotal features, and add the following features:
```
1. http basic auth to verifid users;
2. transfer read configure file to revify user;
3. add timestamp to the upload file;
4. use http delete method to delete file;
```

### How to build
```
go get github.com/arstercz/transfer
cd $GOPATH/src/github.com/arstercz/transfer
go build -o transfer *.go
```

### Run with conf file
```
./transfer -log /tmp/trans.log -port 8000 -temp /data/transfer/logs --basedir /data/transfer/logs -conf /etc/user.conf
```

### config file example
the following config file set the user and password that the client must use http basic auth to verify wheather the user is valid. the username is `arstercz` or `user2`, you can set multi users. 
```
$ cat /etc/user.conf 
[user1]
pass = pass1
[user2]
pass = pass2
```

### How to transfer and delete file

you can use `curl` to transfer file, such as following command:
```
# curl -u 'user1':'pass1'  --upload-file file http://127.0.0.1:8000/
http://127.0.0.1:8000/2xzy/file-20171222154850

# curl -u 'user1':'pass1'  -X DELETE http://127.0.0.1:8889/2xzy/file-20171222154850       
```

and you add `transh.sh` content to your .zshrc or .bashrc, and then `source .zshrc` or `source .bashrc`, you can change user and pass info by yourself.

#### 1. transfer file


```
$ transfer mysql.3306.txt 
######################################################################## 100.0%
http://127.0.0.1:8000/9b09/mysql.3306.txt-20161202154850
```

#### 2. delete file

```
$ transfer-del http://127.0.0.1:8000/9b09dhj/mysql.3306.txt-20161202154850
delete 9b09/mysql.3306.txt-20161202154850 ok
```

#### 3. transfer with wrong user or pass:

http 401 returned when there is wrong user and pass
```
$  transfer mysql.3306.txt 
Authorized error!
```
