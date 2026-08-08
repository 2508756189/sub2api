# Ensemble Diagnostics and Lifecycle Design

## Goal

Make native Ensemble groups reliable to operate and diagnose. A request must
use the saved group configuration on the authenticated gateway path, charge
every successful proposer and aggregator through the existing usage pipeline,
and expose enough runtime state for an administrator to understand failures.

## Scope

- Ensemble remains a native `platform=ensemble` group in sub2api.
- Runtime aggregation is in-group only. The Ensemble group routes through the
  accounts bound to that group; it does not dynamically route through another
  group.
- Candidate and aggregator calls use the existing gateway dispatch and billing
  behavior. The test action is a real request and is billed normally.
- The public model list of an Ensemble group remains the single public model
  `ensemble`. Internal members are shown in the Ensemble administration view.

## Backend Design

### Authentication snapshot

`ensemble_config` must be included in the API-key auth projection and in the
L1/L2 auth snapshot. A cache hit must preserve the aggregator flag, timeout,
minimum proposer count, token limit, and metadata flag. Tests cover both a
database lookup and a cache hit.

### Execution and diagnostics

The fan-out/fallback logic is extracted behind one execution boundary so the
normal gateway handler and the administrative test endpoint share the same
behavior. The administrative test endpoint streams authenticated SSE events:

- request accepted and execution plan
- proposer started, succeeded, or failed
- aggregator started, succeeded, or failed
- fallback selected, if applicable
- final response and usage summary

Member content, errors, duration, token counts, and cost are included only in
the authenticated diagnostic response. The normal OpenAI response contract is
unchanged. A cancellation signal stops waiting and cancels the execution
context; already-issued upstream usage remains recorded by the existing
billing path.

### Group lifecycle

Create and update use explicit semantics. Creating an Ensemble group is one
transaction that creates the group, copies the selected account bindings,
persists members and configuration, and attaches the billing channel. Editing
an existing group cannot silently create or replace another group's identity.
The server keeps the active-name unique index and adds a friendly conflict
error before the database constraint is reached where possible.

Deleting an Ensemble group removes its channel and account join rows, soft
deletes its Ensemble members, invalidates affected API-key caches, and leaves
no runnable Ensemble plan for the deleted group. The operation remains
consistent with the repository's existing soft-delete policy.

### Candidate availability

The admin candidate API reports model, concrete protocol, routable account
count, and billing source separately. Channel pricing is not presented as
proof that an upstream accepts a model. An account with an empty mapping is
shown as permissive/needs runtime verification rather than silently treated as
verified.

## Frontend Design

- Start in an explicit “new Ensemble group” mode.
- Selecting an existing group enters an explicit edit mode.
- Provide “save as new group” so editing never overwrites the source group.
- Use the shared `Select` component for all standard selects.
- Close candidate/source pickers after selection, on outside click, and on
  Escape.
- Open a modal for a real test request. Show a timeline of member states, a
  final answer panel, candidate details, and an error/fallback explanation.
- If the response lacks Ensemble metadata, show that the request did not
  complete through the Ensemble route instead of displaying `0/0`.

## Failure Behavior

- A proposer timeout or upstream 429 is visible per member and does not hide
  other successful members.
- Fewer successful proposers than `min_proposers` returns a clear 502-style
  API error and a completed diagnostic event.
- Aggregator failure falls back to the longest successful proposal and labels
  the fallback explicitly.
- Partial lifecycle writes are rolled back by the create/update transaction.

## Verification

- Go unit tests for auth snapshot persistence, runtime aggregation, failure and
  fallback events, duplicate names, and delete cleanup.
- Frontend typecheck and focused component tests for picker closure, create vs
  edit mode, test modal states, and missing metadata.
- Docker build and local health check.
- A user-triggered real test request is required for end-to-end proof; no paid
  request is made automatically during development verification.
