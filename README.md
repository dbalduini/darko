## darko

**Why?**

Adding a rate limiter into your application may cost a huge amount of memory increase.
To avoid this, you may choose to drop the requests. This is also known as *leak bucket rate limiter*.

If your application is allowed to drop the requests, then you don't need darko or any other external rate limiter.

**darko** is a scalable rate limiter that gives you:
- Guaranteed order on the requests
- No data loss
- Vertical scale
- Horizontal scale on HA mode

## Standalone Mode

![darko standalone mode](/static/darko-standalone.png)

## High Availability Mode

![darko high availability mode](/static/darko-ha.png)

### Master

- Act as writer
- Accept writes and creates new jobs

### Followers

- Act as reader
- Read the new jobs for its partition key

## Guarantees

### Order

### No Data Loss

To avoid the loss of data, set the redis db persistence to AOF.

See more about redis persistence [here](https://redis.io/topics/persistence).

## TODO

- Improve logging
- Dynamic scale of followers
- Monitoring (maybe Prometheus)
