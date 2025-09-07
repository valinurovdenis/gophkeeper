# gophkeeper
Client-server application for storing files in cloud. Files are stored and transmitted in encrypted form. S3 and postgresql are used to store files data and user infos.

## Server
cd gophkeeper/server

go build .

./server --help

### Start example
./server -x "postgresql://localhost/shortener?user={username}&password={password}" -a "minioadmin" -s "minioadmin"

## Client commands list:
cd gophkeeper/client

go build -o gophkeeper

./gophkeeper --help

export SERVER_ADDRESS="localhost:8080"

### User register and login:
./gophkeeper register --login {login} --password {password}

./gophkeeper login --login {login} --password {password}

### List all user files:
./gophkeeper list-files

### Upload file from local path to storage:
./gophkeeper upload --path {path} --comment {optional.comment} --name {optional.name}

### Download file with given id from storage to local path:
./gophkeeper download --path {path} --id {id}

### Delete file with given id
./gophkeeper delete --id {id}

### Get current build version:
./gophkeer version