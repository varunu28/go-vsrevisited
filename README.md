# go-vsrevisited
Golang implementation of Viewstamped Replication revisited protocol

## ToDo:
 - [X] Add a timeout on client side till which it waits for response
 - [X] If timeout expires client can send the request to all nodes
 - [X] Update server code to only respond to client request if they are the leader
 - [X] Add check to figure out if a node is a leader even after multiple view changes (IF viewNumber % NUMBER_OF_NODES == replicaNumber)
 - [X] Add code to send a commit message after receiving quorum
 - [ ] Add a timeout on replica side to check primary node's liveness
 - [ ] Add a key-value store for server code