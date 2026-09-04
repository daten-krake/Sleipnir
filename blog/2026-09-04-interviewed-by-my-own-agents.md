# the day i got interviewed by my own agents

today we didn't write a single line of code. we did something more
valuable: we decided how the thing actually thinks.

the setup: my architect and the principal engineer each read the entire
repo — spec, all twenty adrs, the threat model — and came back with
research, options, recommendations. then they interviewed me. fifteen
questions, one by one, and i said yes, no, or "explain that again".

some of what we locked down:

- agents hand knowledge to each other as fixed views over a graph, not as
  chat history. the platform decides what a stage gets to see. no free
  querying in v1 — the leash stays short until it proves itself on a real
  target.
- findings and hypotheses are two different things now. a guess is not an
  evidence. both live in the graph, both carry provenance, nothing gets
  overwritten — only superseded.
- the workers are dumb on purpose. they execute and report back. they
  can't read the graph, can't ask for more work, can't approve anything.
  the less a worker can do, the less a compromised worker can do.
- machine tokens will never touch scopes, blacklists, approvals or the
  kill switch. that rule is permanent, and it's written down in three
  layers so a bug in one layer can't open the door.
- our event log is hash-chained from day one. if the chain ever breaks,
  customer reports don't leave the building — unless a human overrides it,
  and that override gets logged and stamped on the report. no silent
  exports with a broken chain.

we also gave ourselves a design rule i care about: keep it simple, don't
over-engineer. every abstraction has to justify itself. boring stdlib
beats clever. the next reader of this code is an agent, and agents don't
forgive cleverness.

one nice detail: every work package from here on has to be small enough
that a junior dev — or a junior agent — can do it without architectural
judgment. that's how you keep quality when machines write the code.

next up: the contract documents, then ci/cd, monitoring, and something i
find genuinely exciting — benchmarks for the models, so we notice when a
model gets worse before our results do.

no code today. and it was one of the most productive days so far.

if you want to watch this happen: everything — spec, decisions, design
rules, this blog — lives in the open at
[github.com/daten-krake/Sleipnir](https://github.com/daten-krake/Sleipnir),
and today's decisions are up for review in
[PR #1](https://github.com/daten-krake/Sleipnir/pull/1).
