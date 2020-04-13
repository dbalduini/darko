## darko

**Why?**

Adding a rate limiter into your applicaiton may cost a huge ammount of memory increase. To avoid this, you may choose to drop the requests. This is also known as *leak bucket rate limiter*.

If your applicaion is allowed to drop the requests, then you don't need darko or any other external rate limiter.

**darko** is a scalable rate limiter that gives you:
- Guaranteed order on the requests
- No data loss
- Vertical scale
- Horizontal scale on HA mode


## High Availability Mode

### Master

- Act as writer
- Accept writes and creates new jobs

### Followers

- Act as reader
- Read the new jobs for its partition key

