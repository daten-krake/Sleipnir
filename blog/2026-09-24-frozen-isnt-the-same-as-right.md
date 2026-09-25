# frozen isn't the same as right

today we wrote the first four foundation packages: `ids`, `timex`, `caps` and
`cjson`. small, boring, load-bearing. every id in the platform comes out of the
first, every timestamp out of the second, every size limit out of the third, and
the fourth is the one canonical json form — which matters more than it sounds,
because our event log is hash-chained and our approvals are fingerprints. two
implementations of "canonical" would mean two different chains, and the second
would win by silently invalidating the first.

the contracts covering all this were frozen two sessions ago. frozen means a
clause only changes by a new decision record. so it was mildly humbling that the
agents implementing them found three places where the frozen text was just wrong:

- the truncation marker `[truncated]` was documented as 12 bytes, in the same
  sentence that spelled out the literal. it's eleven.
- a clause justified a rule with "go normalizes leap seconds". it doesn't;
  `time.Parse` rejects `:60` outright. the rule was right, the reason wasn't, and
  the real reason — a numeric offset parses cleanly — is the one that matters.
- another said size *and* nesting depth are checked "before canonicalization".
  you cannot know the depth without parsing. it now says what a single-pass
  canonicalizer can actually promise: depth is enforced at every container
  boundary, and no bytes come back if any check fails.

all three are errata now, each with a decision behind it. none changed a MUST.

the best catch was in our code, not the contract. `cjson` had decided what "the
same key" means twice, and disagreed with itself: duplicate keys were compared
case-insensitively, the exclusion list byte-exact. worse, the case-insensitive
comparison used `strings.ToLower`, which is weaker than what `encoding/json`
does — `{"s":1,"ſ":2}` (that second character is a long s) sailed through, while
`json.Unmarshal` of `{"ſ":7}` lands happily in a struct field tagged `json:"s"`.
a canonical document that a downstream decoder reads differently is exactly the
producer/verifier split the clause exists to prevent, in the one package whose
whole job is byte-exactness.

the obvious fix is also a trap: comparing every key pair with
`strings.EqualFold` is quadratic in the keys of one object, and a 1 MB hostile
document carries about forty thousand of them — 111 ms at four thousand keys,
which extrapolates to eleven seconds of cpu per request. so we fold each key
once, to the minimum of its simple-fold orbit. linear, agrees with `EqualFold`
exactly, and for the ascii keys our contracts actually declare it allocates
nothing at all.

here's the part i want to remember. across four packages there were twenty-five
review findings, and not one said "this code computes the wrong thing". they were
all either *this test cannot fail* or *this comment says something untrue*. an
import allow-list that never actually required `crypto/rand`, so swapping in a
deterministic reader would have stayed green. a truncation corpus missing the one
boundary where the rune walk-back reaches zero. a source scan hardcoding the
parameter names `a` and `b`, so renaming them would quietly delete a security
assertion. a doc claiming a nil clock "can't happen because boot would crash
first" — it can, and nothing dereferences it at boot.

and the agents caught me six times: two of those contract errors, a claim that go
rounds sub-millisecond digits (it truncates), a test row the validity window made
impossible, and a fix i invented that contradicted itself — orbit-minimum plus
"lowercase ascii returns unchanged" disagree about the pair {A, a}, which would
have broken the very example the clause uses.

two rules came out of it, and both are in the repo now. a documented exception
has to be true, because the next agent reads it as licence. and for every guard
you write, ask what edit would make the test pass while breaking the rule — if
you can name one, it isn't a test yet.

everything is in the open at
[github.com/daten-krake/Sleipnir](https://github.com/daten-krake/Sleipnir):
today's work is [PR #4](https://github.com/daten-krake/Sleipnir/pull/4) and
[PR #5](https://github.com/daten-krake/Sleipnir/pull/5), both merged, with the
session record in [PR #6](https://github.com/daten-krake/Sleipnir/pull/6).
