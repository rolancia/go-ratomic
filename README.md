# go-ratomic

## Concept

Maintaining atomicity is very difficult and cumbersome in the concurrent application.

---
Take the case of communication with DBMS for example.

For mutual exclusive whether it's optimistic/distributed or not, we would be with locking.

Most DBMS supports them but in application programmer side it should be more complicated to use.

So I'm trying to give up using DBMS side features once.

Redis can drive the optimistic locking (not meaning Watch), it does not matter kind of DBMS, kind of Language in this way.

Just give up the faster latency to be easy! you might be using cache/lazy update already if your service needs fast latency.

The point is, Redis MSetNX sets key-value pairs atomically. MSetNX will set all locks or not if we want to get multiple locks.


---
## Drivers

[Redis](https://github.com/rolancia/go-ratomic-redis-driver)
