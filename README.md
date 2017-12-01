# Token ring protocol
Token ring network protocol simulation. The implementation can process correctly: 
- token loss
- one node's termination and recovery

## Running task3
To run some example scenario (~ 2,5 minutes of simulation):
```console
$ go get github.com/sokks/tokenring
$ cd `go env GOPATH`/src/github.com/sokks/tokenring/task3/
$ make task3
```

## Implementation
It is a fully distributed implementation. Each node has timers to detect token loss and node terminations. The first timed out node creates new token. If it's timed out more than 3 times, previous node is judjed as terminated. If it receives message from that node, it cancels terimation.  
A node sometimes drops token if it is judged as extra (usually after recovery).  
  
There is no internal control that there is only one terminated node is the ring. If there is one terminated node, others are ignored (judged as token loss).

## Environment
Network interface: *lo*  
Network protocol: *UDP*  
Base token port: *30000*  
Base maintainance port: *40000*  
Logging: *Stdout* TODO: use some logging package  

### Notes
There can be some data races (send first token) but they are not critical because of infinite cycle.  
Message that is not acked is sent again even if ack is lost (not message itself).  