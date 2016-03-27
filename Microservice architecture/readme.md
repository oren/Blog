# Microservices in Go

Code sample for [this](https://jacobmartins.com/2016/03/14/web-app-using-microservices-in-go-part-1-design) article

![pic](https://www.lucidchart.com/publicSegments/view/cb49c63f-9256-47ae-a21a-18afa85cc4fd/image.png)

## Run

### key-value store (for configuration)
```
cd keyvaluestore
go build
./keyvaluestore
```
It's running on port 3000

### Database (storing tasks)
```
cd Database
go build
./Database 127.0.0.1:3001 127.0.0.1:3000
```
It's running on port 3001

### Storage (for images)
```
cd Storage
go build
./Storage 127.0.0.1:3002 127.0.0.1:3000
```
It's running on port 3002

### Master
```
cd Master
go build
./Master 127.0.0.1:3003 127.0.0.1:3000
```
It's running on port 3003

### workers
```
cd Worker
go build
./Worker 127.0.0.1:3000 3
```

### Frontend
```
cd Frontend
go build
sudo ./Frontend 127.0.0.1:3000
```
It's running on port 80

open the brower at 127.0.0.1 , choose a file and hit 'upload'

## Misc

show key-value store
```
curl localhost:3000/list

databaseAddress : 127.0.0.1:3001
masterAddress : 127.0.0.1:3003
storageAddress : 127.0.0.1:3002
```

show tasks
```
curl localhost:3001/list

0 :  id: 0  state: 0
1 :  id: 1  state: 1
```

States

* 0 – not started
* 1 – in progress
* 2 – finished
