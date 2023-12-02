# MBuffer for Hoster

This binary simply throttles down a ZFS send/receive stream.
I don't recommend using it as a standalone program, because it was designed to be integrated with `Hoster`.

It works by taking a pipe (`|`) input from `zfs send`, which is then throttled by a simple algorithm (time/bytes sent).
Finally `mbuffer` forwards the output into another pipe to be consumed by `zfs receive`.

In my limited testing ZFS data may get corrupted, if you are not using any kind of data stream normalizers while sending ZFS blocks over the network.
That's why `mbuffer` was implemented in a first place.
And as a bonus we got a throttling mechanism, which helps us keep a stable connection up on the slow networks (or over the WAN).

I always recommend to set the replication speed just a bit lower than a pure line speed to keep things stable.

The replication speed is set using `--speed-limit` command line flag, which in it's own turn simply creates a dynamic system variable that is then picked up by the `mbuffer`.
Here is variable name for those interested: `SPEED_LIMIT_MB_PER_SECOND` (just in case you'd like to use in your own replication scripts).
