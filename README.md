# go-vsrevisited
Golang implementation of Viewstamped Replication revisited protocol

## ToDo:
 - [X] Add a timeout on client side till which it waits for response
 - [X] If timeout expires client can send the request to all nodes
 - [X] Update server code to only respond to client request if they are the leader
 - [X] Add check to figure out if a node is a leader even after multiple view changes (IF viewNumber % NUMBER_OF_NODES == replicaNumber)
 - [X] Add code to send a commit message after receiving quorum
 - [X] Add a key-value store for server code
 - [X] Add a timeout on replica side to check primary node's liveness
 - [X] Add mechanism to buffer request in case replica node is out of date
 - [X] Add mechanism to catch up if replica node is out of date along with state change
 - [X] Add mechanism for clearing the buffer with state change
 - [ ] Add view change mechanism in case of server timeout
