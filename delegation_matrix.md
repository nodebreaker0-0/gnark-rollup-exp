# Delegation Matrix — zkkit

Defines what the agent does autonomously vs. what stops for the human (Carl /
builnad). The user has stated they will not be involved during the build, so the
agent proceeds on everything except the Block column and records rationale in
`decisions/`.

| Area | Auto-merge OK | Notify-only | Block |
|---|---|---|---|
| Circuit / library code | ✅ when `make verify` passes | | |
| Tests (unit, e2e, negative) | ✅ | | |
| Spec updates (R-###, T###, SC-###) | ✅ | | |
| Docs (README, godoc, quickstart) | ✅ | | |
| New direct dependency | | ✅ one-line note in decisions/ | wallet/secret/HSM-related, or >5MB |
| Pinned `gnark` version bump | | ✅ | major version with API breaks mid-build |
| Public API shape of `rollup/` & harness | ✅ when documented + tested | breaking change after MVP tagged | |
| git commit + push to working branch | ✅ when `make verify` passes | | |
| Tag release `v*` | | | ✅ human only |
| Pushing to a real chain / spending funds | | | ✅ human only (out of scope anyway) |
| Solidity verifier / on-chain deploy | | | ✅ human only (out of scope) |
| Destructive ops (`rm -rf`, force-push, history rewrite) | | | ✅ human only |
| Trusted-setup ceremony / real keys | | | ✅ human only |

## Notes for this environment

- The mounted filesystem disallows `unlink`. The agent never relies on deleting
  files; it untracks via `git rm --cached` + `.gitignore` and renames instead of
  deletes. This is a constraint, recorded in `decisions/`.
- Commits happen per logical increment, not one giant commit, so the human can
  review history.
