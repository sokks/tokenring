# Token ring protocol
Fault tolerant token ring network protocol implementation

## Without active monitor
Each node has timers to detect token lost and node terminations. The first timed out node creates new token. If it's timed out again, previous node is judjed as terminated.  
Each node has also one cicle ticker. If there are more than one message received in one cicle, it drops next message.

## With active monitor (*dev*)
One node is an AM. It controls that there is token in the ring and this token is the only. Also it broadcasts periodically a control packet to know which nodes are active. If AM is terminated, a new AM is elected by id (min).