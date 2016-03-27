# Microservices in Go

Code sample for [this](https://jacobmartins.com/2016/03/14/web-app-using-microservices-in-go-part-1-design) article

![pic](https://www.lucidchart.com/publicSegments/view/cb49c63f-9256-47ae-a21a-18afa85cc4fd/image.png)

## Build
```
./build
```

## Run
```
./run
```
open the brower at 127.0.0.1 , choose a file and hit 'upload'

I see the following error:
```
panic: runtime error: invalid memory address or nil pointer dereference
[signal 0xb code=0x1 addr=0x28 pc=0x402468]

goroutine 21 [running]:
panic(0x764860, 0xc82000e140)
        /usr/local/go/src/runtime/panic.go:464 +0x3e6
main.doWorkOnImage(0x0, 0x0, 0x0, 0x0)
        /home/oren/p/go/src/github.com/oren/Blog/Microservice architecture/Worker/worker.go:151 +0x48
main.main.func1()
        /home/oren/p/go/src/github.com/oren/Blog/Microservice architecture/Worker/worker.go:94 +0x61b
created by main.main
        /home/oren/p/go/src/github.com/oren/Blog/Microservice architecture/Worker/worker.go:112 +0xd8b

```

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