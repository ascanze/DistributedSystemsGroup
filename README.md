"# DistributedSystemsGroup" 


a) What are packages in your implementation? What data structure do you use to transmit data and meta-data?

b) Does your implementation use threads or processes? Why is it not realistic to use threads?


------------------------------------------------
c) In case the network changes the order in which messages are delivered, how would you handle message re-ordering?


------------------------------------------
d) In case messages can be delayed or lost, how does your implementation handle message loss?

- Our task 1 implementation doesnt handle delays or message loss cause of the guarentee via channels but 
-------------------------------------------

e) Why is the 3-way handshake important?

- Conformation that the connection is ready before the data transfer begins to prevent data loss
