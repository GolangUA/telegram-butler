# Community Vote-Mute (`/report`)

Community-driven moderation: any member can reply to a message with `/report` and start a 3-minute vote. If **N** unique members vote `[Mute]`, the target user is muted for 3 hours.

Spec: [issue #104](https://github.com/GolangUA/telegram-butler/issues/104) — simplified scope (no Pardon side, no escalation).

## Packages and responsibilities

```mermaid
flowchart LR
    TG[Telegram] -->|"/report + callbacks"| H["handler/report"]
    H -->|"StartVote / CastVote / ActiveVote"| S["service/report"]
    S -->|"Create / GetActive / AddVoter / SetStatus / ListActive"| R["repository/firestore"]
    R -->|gRPC| FS[("Cloud Firestore<br/>report_votes")]
    H -.->|"on quorum or stale button"| TG
    S -.->|"on timer expiry (goroutine-side)"| TG
```

- **`handler/report`** — Telegram glue. Validates the command, posts/edits the vote message, routes button callbacks, posts the mute side-effect via `bot.RestrictChatMember`.
- **`service/report`** — business logic. Owns the per-vote `Coordinator` goroutines and the state machine. Also holds the `bot` reference for goroutine-side `EditMessageText` on timer expiry (the one Telegram call no handler is alive to make).
- **`repository/firestore`** — persistence. 5-method `Repository` interface, snake_case schema, 5s per-op timeout.

## Vote lifecycle

```mermaid
stateDiagram-v2
    [*] --> active: /report — reporter is voter 1
    active --> muted: quorum reached within 3 min
    active --> expired: 3 min elapsed without quorum
    muted --> [*]
    expired --> [*]
```

- `active` is the only state where the coordinator goroutine runs and the inline button collects taps.
- `muted` / `expired` are terminal. Docs are retained in Firestore for history (composite doc ID `{chatID}_{targetID}_{unixNano}` makes them sortable and inspectable).

## Happy-path: vote reaches quorum

```mermaid
sequenceDiagram
    participant Rep as Reporter
    participant V as Voter (Nth)
    participant TG as Telegram
    participant H as handler/report
    participant S as service/report
    participant C as Coordinator goroutine
    participant FS as Firestore

    Rep->>TG: /report (reply to target)
    TG->>H: Message update
    H->>H: validateReport (admin check, target check, ActiveVote check)
    H->>TG: SendMessage vote text + [Mute (1/N)] button
    H->>S: StartVote(vote)
    S->>FS: Create — status=active, voters=[reporter]
    S->>C: spawn goroutine — ctx with deadline = vote.ExpiresAt
    Note over C: select on CH_IN or ctx.Done()

    V->>TG: tap Mute button (k/N)
    TG->>H: CallbackQuery report_vote_chatID_targetID
    H->>S: ActiveVote then CastVote
    S->>C: send VoteAction (voter) on CH_IN
    C->>FS: AddVoter
    C->>S: send VoteResult (updated vote) on CH_OUT
    S-->>H: result

    alt voters count below Quorum
        H->>TG: EditMessageText — updated voter list + button
        H->>TG: AnswerCallbackQuery — empty toast
    else voters count reached Quorum
        C->>FS: SetStatus(muted)
        C->>S: send VoteResult Finished=true on CH_OUT
        H->>TG: AnswerCallbackQuery — empty toast
        H->>TG: RestrictChatMember — target, 3h
        H->>TG: GetChatAdministrators — for tagging
        H->>TG: EditMessageText — mute result + admin Cc tags
        Note over C: defer cleanup, goroutine exits
    end
```

## Expiry path: 3 min passes without quorum

```mermaid
sequenceDiagram
    participant TG as Telegram
    participant C as Coordinator goroutine
    participant S as service/report
    participant FS as Firestore

    Note over C: ctx.Deadline reached — no callback handler alive
    C->>C: handleExpire
    C->>FS: SetStatus(expired)
    C->>S: onExpire callback — service.notifyExpired
    S->>TG: EditMessageText "Голосування завершено — недостатньо голосів" — no ReplyMarkup so button is removed
    Note over C: defer cleanup, goroutine exits
```

## Callback dispatch

Telegram delivers all inline-button taps as `CallbackQuery` updates. Multiple handler packages register against the same stream — they fan out via telego's `CallbackDataPrefix` predicate.

```mermaid
flowchart TD
    Q[CallbackQuery arrives] --> Pre{callback_data prefix}
    Pre -->|agree_| Join1["handler/callback.handleAgree<br/>(join flow)"]
    Pre -->|decline_| Join2["handler/callback.handleDecline<br/>(join flow)"]
    Pre -->|report_vote_| Rep["handler/report.handleVote<br/>(this feature)"]
    Pre -->|no match| Drop[ignored]
```

`report/handler.go:Register` registers `voteGroup := bh.Group(th.CallbackDataPrefix(callbackPrefix))` so this package only sees its own callbacks. The earlier blocker (catch-all join handler swallowing report callbacks) was fixed in `e7565ad` by adopting the same prefix-grouping pattern in `handler/callback`.

## Coordinator goroutine internal loop

The for-select state machine that owns a single active vote from creation until quorum/expiry. One goroutine per active vote, registered in `Coordinator.registry` (a `sync.Map`) and cleaned up by a `defer` on exit.

```mermaid
flowchart TD
    Start["Start vote<br/>1. register in sync.Map<br/>2. spawn goroutine<br/>defer: registry.Delete"] --> Loop{select}
    Loop -->|"voter on CH_IN"| Add[repo.AddVoter]
    Add --> ErrCheck{err?}
    ErrCheck -->|yes| ErrOut["send VoteResult Error on CH_OUT"] --> Loop
    ErrCheck -->|no| Quorum{voters reached Quorum?}
    Quorum -->|no| Update["send VoteResult (updated) on CH_OUT"] --> Loop
    Quorum -->|yes| Mute["repo.SetStatus(muted)<br/>send VoteResult Finished=true on CH_OUT<br/>return"]
    Loop -->|"ctx.Done — deadline"| Expire["handleExpire<br/>repo.SetStatus(expired)<br/>onExpire callback<br/>return"]
```

Note: the goroutine talks to Telegram **only** via the `onExpire` callback (which invokes `service.notifyExpired` → `bot.EditMessageText`). Every other Telegram call happens in the handler goroutine after reading from `CH_OUT`. The expiry edit is the only case where no handler is alive to do it.

## Stale-button edit (orphan recovery)

When the vote's Firestore doc is gone (data wipe, manual cleanup, race orphan) but the Telegram message still has its button, tapping it triggers an explicit "this vote is over" edit instead of doing nothing.

```mermaid
flowchart TD
    Tap["User taps Mute button<br/>on a stale message"] --> CB["handler/report.handleVote"]
    CB --> AV["service.ActiveVote"]
    AV -->|"err == ErrVoteNotFound"| Stale["editStaleMessage<br/>+ AnswerCallbackQuery toast"]
    Stale --> Edit["Telegram EditMessageText<br/>❌ Голосування недоступне<br/>button removed"]
    Stale --> Toast["Toast: Голосування більше не активне"]
    AV -->|"err == other"| Infra["log + AnswerCallbackQuery<br/>'Сталася помилка, спробуйте пізніше'"]
    AV -->|ok| CastFlow["CastVote — quorum or update"]
```

Same branch runs if `CastVote` returns `ErrVoteNotFound` (e.g. coordinator goroutine died between the `ActiveVote` check and the `CastVote` call).

## Startup reconciliation

The Cloud Run instance is ephemeral — restarts happen on deploys, cold-starts, scaling-to-zero. Active votes survive via `Service.Reconcile`:

```mermaid
flowchart LR
    Boot["main.go setup<br/>after handler.Register"] --> Rec["Service.Reconcile<br/>10s timeout"]
    Rec --> LA["repo.ListActive"]
    LA -->|votes| Loop["for each vote:<br/>coordinator.Start"]
    Loop --> GR["per-vote goroutine<br/>ctx.WithDeadline = vote.ExpiresAt"]
    GR -->|past deadline| Imm["fires expire immediately<br/>SetStatus + EditMessage"]
    GR -->|future deadline| Wait["waits on CH_IN or ctx.Done"]
    LA -.->|"err — timeout, unauthorized"| Warn["log Warn, continue startup<br/>bot still serves /help, /mute, etc."]
```

## Concurrency model

- **One Cloud Run instance** (max instances = 1). No multi-process coordination needed today.
- **One coordinator goroutine per active vote**, keyed in a `sync.Map` by `{chatID}_{targetID}`. Tracked in `Coordinator.registry` (`coordinator.go`).
- **Per-vote serialization** — only the coordinator goroutine writes to that vote's Firestore doc. Callback handlers send via unbuffered channels and wait for the result. No transactions required at the repo layer.
- **Known race** in `validateReport`: two concurrent `/report`s on the same target can both pass the `ActiveVote` gate before either has written. Inline comment + `tasks/2026-05-16-firestore-phase2-progress.md` Phase F #7 document. Not a double-mute risk (only one coordinator collects votes); the orphan auto-expires.

## Timeouts at a glance

| Where | Bound | Why |
|---|---|---|
| Per-Firestore-op | `5s` (`repository/firestore.firestoreOpTimeout`) | Single `/report` or vote tap must fail fast on backend outage |
| Reconcile at startup | `10s` (`cmd/telegram-butler/setup.reconcileTimeout`) | Firestore outage must not block bot startup |
| Vote window | `3m` (`service/report.VoteWindow`) | Spec |
| Mute duration | `3h` (`service/report.MuteDuration`) | Spec |
| Quorum | `5` (`service/report.Quorum`) | Spec (issue #104) |

## Firestore indexing

The `validateReport → ActiveVote` query has 3 equality filters and no `OrderBy` / range:

```go
.Where("chat_id",        "==", X)
.Where("target_user_id", "==", Y)
.Where("status",         "==", "active")
.Limit(1)
```

Firestore satisfies this via **index merging** over auto-created single-field indexes — **no composite index required**. Verified live in `golangua-telegram-butler` against the real `report_votes` collection (no precreated indexes; query succeeded). See [Index types in Cloud Firestore](https://firebase.google.com/docs/firestore/query-data/index-overview).

A composite index would become necessary only if a future change adds:
- A range / inequality filter (e.g. `Where("created_at", ">", t)`)
- An `OrderBy` on a field different from the equality fields
- Operational scale where Firestore deems the merge plan too expensive

If that happens, the first failing query returns `FAILED_PRECONDITION` with an autosuggest URL — click once in the console to create the exact index.

## Failure modes

| Failure | What happens | Visible? |
|---|---|---|
| Firestore unreachable on `/report` | After 5s, handler replies `⚠️ Сервіс тимчасово недоступний` | Yes (chat reply) |
| Firestore unreachable on Reconcile | `Warn` log, bot starts anyway, in-flight votes lost until next restart | Logs only |
| Vote message exists but doc is gone (orphan) | Next button tap edits message to `❌ Голосування недоступне` | Yes (chat edit) |
| Bot restarts mid-vote | Reconcile respawns coordinator; if past deadline, fires expire immediately | Yes (message edited to expired) |
| Concurrent `/report` on same target | Both can pass validation → two docs + two messages; only one collects votes, the other auto-expires | Yes (two messages) |

## File map

```
internal/handler/report/
  handler.go    Telegram entry points (handleReport, handleVote) + edit/delete helpers
  parse.go      callback_data <-> chatID/targetID, voterFromUser
  render.go     HTML formatters (vote text, button, mute result, admin tags)

internal/service/report/
  types.go      Repository interface + VoteAction/VoteResult + Quorum/VoteWindow/MuteDuration
  service.go    StartVote / CastVote / ActiveVote / Reconcile + notifyExpired
  coordinator.go Per-vote goroutine state machine, sync.Map registry
  coordinator_test.go / service_test.go  memory-repo-backed unit tests

internal/repository/firestore/
  client.go     SDK client constructor + type alias
  vote.go       VoteRepository (5 methods, 5s per-op timeout, voteDoc/voterDoc schema)
  vote_test.go  testcontainers-backed integration tests (self-bootstrap emulator)

internal/repository/memory/
  vote.go       sync.Mutex map (test double — production uses repository/firestore)

internal/entity/
  vote.go       Vote, Voter, status constants, ErrVoteNotFound
```
