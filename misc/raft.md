# The RAFT algorithm and Distributed State Properties and Solutions

## Sources

The best source is the [RAFT paper](https://raft.github.io/raft.pdf) itself, which is surprisingly understandable for an academic paper, an details how each step in the process satisfies or excludes specific required/undesired properties. The best way to understand RAFT is as a daisy chain of core principles each of which verified by solving specific timing and sub-problems. 

In other words, it is a hot-mess of asynchronous logical properties, because distributed consistency is as difficult in computers as it is in people. Or so I was told by a facebook fact-checker the last time I made any number of common sensical claims about the state of the world that a fraction of a fraction of the populace strategically proclaim are controversial. Friggin lizard people.  :)

* https://raft.github.io/raft.pdf
* wikpedia
* Eli Bendersky

## Notes:
* CAP / Brewer's theorem: 
    * Important because it contrasts to ACID transactions. One must deeply understand the properties of CAP to ensure ACID.
    * States that a deterministic distributed system may only satisfy two of three properties:
        1) Consistency: every read receives the most recent write or an error
        2) Availability: every request receives a non-error response without the guarantee that it contains the most recent write
        3) Partition tolerance: the system continues to operate despite an arbitrary number of dropped messages between nodes
        
        Gist: when a network partition occurs, it is an implementation decision to favor availability of consistency. When a partition occurs and requests fail, the store as a whole must decide whether to:
            1) cancel the request to favor consistency but concede availability
            2) satisfy the request to favor availability but risk inconsistency
    * RDBMS implementing ACID transactions usually favor consistency over availability
    * NoSQL systems often implement a BASE (basically available, soft state, eventually consistent) philosophy, favoring availability through eventual consistency.
    * These are pure opinions, but are expressed to point out properties that highlight certain design decisions.

## RAFT

RAFT fault tolerance

| Num Servers | Num Failures Tolerated |
|  :---: | :----: | 
| 3 | 1 | 
| 5 | 2 |
| 7 | 3 |
| 9 | 4 |
| ... | ... | 
| 2n + 1 | n |
| 2n + 1 = k | floor(k / 2) |

* RAFT favors consistency at the expense of availability.
* Client requests trigger replication and persistence across replicas before client receives a response.
    * Bad for high frequency db transactions, good for coarse-grained objects like config.

### Defs

1) Followers: replicates state from the leader, always ready to take over leadership if leader goes down.
2) Leader: accepts client requests, adds new entries to the log, replicates these changes to followers.
3) Election timer: every follower implements a randomized timer value that restarts every time it hears a heartbeat from the leader. Expiration triggers an election. Randomization facilitates speedy election.
4) Candidate: the state to which a follower transitions when its timer expires.
5) Term: the election counter; Candidates increment when initiating an election.
6) Log: the state machine input written and replicated by the leader
7) Quorum: when more than half of participating peers agree.

RPCs:
1) RequestVotes: sent by Candidates to peers during elections; responses indicate peer votes.
2) AppendEntries: Sent by Leader to replicate log entries and also as a heartbeat.

RAFT implements a single-leader model, but whereby all servers participate to elect and support the leader. Participants are essentially components of a single distributed leader, as opposed to the pejorative definition of "follower".

### Leader Election
The leader election proceeds as follows. Each Follower implements an election timer for a randomized period, reset when it receives AppendEntries messages from Leader. When expired, the Follower transitions to a Candidate and initiates an Election. The randomization of the election timer improves the odds that a node initiates an election earlier than the others, and thus elections resolve quickly.

Candidate increments the Term and immediately votes for itself and sends out RequestVotes messages. If the Candidate receives messages with a term larger than its own, it immediately stops its election and recognizes the new leader. If a Candidate receives a majority of votes, it becomes the new Leader. If a split vote occurs, i.e. there are an even number of servers and two initiate new elections simulaneously, then the election restarts.

### Log Replication

When a client requests a command be submitted to the replicated StateMachine, the Leader appends the command to its log and issues AppendEntries messages to all Followers. When a quorum of Followers respond with confirmation, the request is considered committed. All previous entries are also considered committed. Followers are likewise notified and commit as well. Obviously there are timing inconsistency considerations, but just ignore as a subproblem. The point is that mutual confirmation is received, and state ratchets forward.

Log inconsistency can occur when the Leader disconnects. Some logs of the Leader may not be fully replicated throughout the cluster. After a new election, the new Leader reconciles the Log by negotiating the last entry on which all of its followers agree. If a server has committed a log to its StateMachine, then no other server may commit a different command for the same entry.

Note that in a partitioned state of odd-numbered clusters, partitions with half the nodes or fewer cannot ratchet forward. They remain Candidates.

### Timing requirements

RAFT requires the following property of fault tolerance:
```
  broadcastTime << electionTimeout << MTBF
```
which are properties of the deployed infrastructure.

* broadcastTime: (0.5-20ms) the total time it takes the leader to send and receive messages from all followers.
* electionTimeout: (10-500ms) specifies the time before a Follower transitions to Candidate and initiates and Election. Obviously this cannot be about equal to or less than the broadcastTime since broadcasts are required for timely elections.
* MTBF: (weeks or months) the MBTF of server nodes. If close to the electionTimeout then hypothetically elections could not be resolved before another potential failure occurred.

### Practical Properties

1) An odd number of nodes is required
2) Latency property required: `broadcastTime << electionTimeout << MTBF`
3) Completeness: if a Leader commits an entry, then the entry is present in all subsequent Leaders.
4) If a server has committed an entry at a given index, then no other server will commit a different entry for that index.
5) Leader append-only: entries are only appended, never deleted or modified.

### Relationship with Kubernetes

The API Server observes changes to distributed state and notifies Controllers of these changes. The `watch` mechanisms of the API server ensure that observers (Controllers, Kubelets, and KubeProxies on the nodes) are bound to the state changes recorded by etcd.
