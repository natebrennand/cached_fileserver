
Nate Brennand

nsb2142

Distributed Systems HW #1



## 1

Contrast TCP and UDP.
Under what circumstances would you choose one over the other?


TCP and UDP are both built on top of IP.
UDP has faster transfer speeds because packets have no guarantees with respect to order or receiving.
TCP is a slower protocal that has built-in guarantees for order of packets as well as guarantees that the packet is received by the recipient.
These guarantees are earned by ordering the packets after recieving them and using acknowledgement messages for each packet.

UDP should be used when throughput is needed and it is not necessary to recieve 100% of the packets.
Streaming music or video is one of these cases.
TCP should generally be used in all other situations, especially when the data must be recived.


## 2

What's the difference between caching and data replication?


Caching is the process of saving a result of a calculation / rendering / query so that it can be used to serve more than one request.
The results are saved because they are computationally expensive or interact with a highly contested resource (like a database).
The danger of caches is that they must be refreshed when the data behind the operation changes, otherwise you risk returning stale data for a query.

Data replication is the act of making multiple copies of data.
This is done for either performance or redundancy.
For instance, in a database that has a high number of reads with few updates, it can be viable to have multiple copies of the same database and split database queries between them.
If redundancy is needed, you may be splitting the data between multiple computers or multiple data centers so that in the case that one of them goes down, you can still access your data and maintain uptime.



## 3

Why do we need an application-level cache to optimize programs, i.e. what are the benefits of application-level cache over hardware or os-level cache?

Hardware and os level caches are great for reducing disk accesses or keeping frequently used data in faster storage than RAM.
Application-level caches are used for situations that involve computation for the result.
For example, a database query that performs several aggregate operations.



