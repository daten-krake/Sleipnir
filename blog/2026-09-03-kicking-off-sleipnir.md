# kicking off sleipnir

so we started building sleipnir. the name is odin's horse, eight legs, goes
everywhere. felt right for a thing that's supposed to move through a network
on its own.

the idea is simple to say and hard to do: an automated pentest platform
where agents do the work. recon, research, moving around active directory.
but we don't hand over the keys. anything that actually breaks something 
exploitation, getting a first foothold, escalating privs needs a human to
say yes first. recon and research run free, the dangerous stuff waits for
an operator. that line matters to us a lot.

we set ourselves some hard rules early, mostly to keep it honest:

- golang, standard library only. no external dependencies. if we need one
  we have to write down why in an adr first. the one exception so far is
  postgres, and that got its own decision record.
- htmx for the ui. server rendered, no frontend framework, no node build.
- everything runs in docker, and anything hostile lives in a throwaway
  container.
- it has to run on local ai. we want to prove this works on local models,
  not just someone else's api. cloud models are fine too, but then the data
  gets masked before it leaves. captured creds never leave, full stop.

we also decided the whole platform gets built by agents. a principal
engineer plans and reviews, and spawns implementers in parallel to do the
actual work. that means the code has to be easy for an agent to read and
debug, so we made errors describe themselves what went wrong and in which
function and we stick to idiomatic go. no cleverness for its own sake.

there's a lot we still don't know. how the handover between agent stages
should look, the exact program layout, the ui. we're tracking all of it in
a backlog and writing down every real decision as an adr, so future us (and
the agents) aren't guessing.

first session done. decisions made, threats written down, logo drawn. the
eight-legged horse is out of the gate.
