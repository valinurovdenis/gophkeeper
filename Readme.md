# gophkeeper
## Client commands list:

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