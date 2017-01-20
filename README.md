# Elevator Cluster

## Brainstorming general algorithm
- All external button-presses gets broadcasted as an `Order`
- All peers acknowledges each `Order` broadcast with an `OrderStatus`:
  - `"RECEIVED"`
    - A simple acknowledgement, no further action is taken by the peer.
  - `"ASSIGNED"`
    - The peer have added the order to its internal queue and will eventually expedite. (Maybe it should be cancelable?)
  - `"STARTED"`
   - The order is know first in the elevators internal queue, and the carriage is on its way.
  - `"COMPLETE"`
   - The carriage have successfully expedited the order ie. reached the desired floor and "picked up" the people ordering it.
- Whenever a peer receive an Order it computes the the cost for each known elevator in the cluster to expedite the order, and adds a list of all elevators to the order in its own state. The list is sorted ascendingly by cost.
  - If the peer finds itself on top of this list, ie. have the lowest cost, then the order gets added to its internal queue and broadcast `"ASSIGNED"` or `"STARTED"` (depending on whether there are other elements already in the queue)
  - If the peer does not have lower cost, it will listen for acknowledgements from the elevators above it on the list. If these fail to acknowledge the order after some time, the peer will assign it to its own queue anyway.
- All elevators broadcast their state as part of their heartbeat:
 - Last floor, and whether it currently is in a floor
 - Motor direction: `"STOP"`, `"UP"` or `"DOWN"`
 - Something more...?

***
### Small stuff
- Set an small but random delay to acknowledgements, in order to mitigate congestion.
