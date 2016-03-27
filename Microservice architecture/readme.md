# Microservices in Go

This is an example for a Micro-services architecture. 
It's an app with a web interface that accepts a png file and modify it's colors. The user upload the file and the backend services process the image and store it in /tmp folder. It's the code for [this](https://jacobmartins.com/2016/03/14/web-app-using-microservices-in-go-part-1-design) article.

Here is a high level diagram of the different services:  
We use 6 separate executables: Frontend, Master, Task store, Storage, key-value store, and Workers.
![pic](https://www.lucidchart.com/publicSegments/view/cb49c63f-9256-47ae-a21a-18afa85cc4fd/image.png)

## Build
```
./build
```
This command will create 6 executables in the bin folder: config-store, tasks-store, images-store, master, worker, and frontend.

## Run
```
./run
```
open the brower at 127.0.0.1 , choose a png file and hit 'upload'

To verify it's working view the 2 png files: /tmp/working/0.png and /tmp/finished/0.png  
The first one is the original image and the second one is the modified image.

## Stop
```
./stop
```

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
