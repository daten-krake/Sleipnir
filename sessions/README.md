# Session Tracker

One file per working session: `YYYY-MM-DD-<topic>.md`.

Each record contains:
- Date, session id, model
- Goal of the session
- Decisions made (link to ADRs)
- Open questions carried forward
- Token usage (input / output / cache / reasoning), snapshotted at start and end

`update-usage.sh` re-reads the live pi session log and refreshes the usage
table in a session file:

    ./sessions/update-usage.sh sessions/2026-09-03-architecture-kickoff.md
