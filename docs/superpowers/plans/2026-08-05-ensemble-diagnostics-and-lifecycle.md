# Ensemble Diagnostics and Lifecycle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make native Ensemble groups correctly execute their saved configuration, expose real progress and billing details during tests, and prevent accidental overwrite or incomplete deletion.

**Architecture:** Preserve the existing in-group Ensemble gateway path and usage recorder. Repair the API-key auth snapshot first, then add a shared diagnostic event sink to the handler and an authenticated admin test stream. Make group lifecycle operations explicit in the frontend and transactional where the current repository boundaries permit it.

**Tech Stack:** Go/Gin, Ent/PostgreSQL, existing API-key auth cache, Vue 3 Composition API, TypeScript, existing `Select.vue`, SSE, Docker-based Go verification.

## Global Constraints

- Ensemble members are served only by accounts bound to the Ensemble group.
- Proposer and aggregator calls use the existing gateway dispatch and billing path.
- Public Ensemble model exposure remains `ensemble`; internal members are admin-only metadata.
- No automatic paid end-to-end request is made during verification.
- Preserve unrelated upstream and local branch changes.

---

### Task 1: Preserve Ensemble Configuration in Auth Snapshots

**Files:**
- Modify: `backend/internal/repository/api_key_repo.go`
- Modify: `backend/internal/service/api_key_auth_cache.go`
- Modify: `backend/internal/service/api_key_auth_cache_impl.go`
- Test: `backend/internal/service/api_key_auth_cache_test.go` or the existing auth-cache test file

**Interfaces:**
- `GetByKeyForAuth` must populate `APIKey.Group.EnsembleConfig`.
- `APIKeyAuthGroupSnapshot` must serialize and restore `EnsembleConfig`.

- [ ] **Step 1: Add a regression test** that builds an API-key auth snapshot with `AggregatorEnabled=true`, a non-default timeout, and `MinProposers=2`, then restores it and asserts all values survive.
- [ ] **Step 2: Add `group.FieldEnsembleConfig` to the auth-only Ent projection.**
- [ ] **Step 3: Add `EnsembleConfig` to `APIKeyAuthGroupSnapshot` and copy it in both snapshot conversion directions.**
- [ ] **Step 4: Run the focused auth-cache tests and the existing Ensemble runtime tests.**

Expected result: a cache hit supplies the same aggregator-enabled plan as a database lookup.

### Task 2: Make Ensemble Runtime Diagnostics Observable

**Files:**
- Modify: `backend/internal/handler/ensemble_chat_completions.go`
- Modify: `backend/internal/handler/ensemble_chat_completions_test.go`
- Modify: `backend/internal/service/ensemble_runtime.go` only if the event plan needs a stable exported representation

**Interfaces:**
- Introduce an internal event callback with event name, member identity, status, timing, usage, cost, content, and error.
- Existing `ChatCompletions` remains compatible when no callback is installed.

- [ ] **Step 1: Add tests for proposer start/success/failure, aggregator start/success, aggregator fallback, and minimum-proposer failure events.**
- [ ] **Step 2: Add the event sink to `EnsembleHandler` and emit events around each sub-call without changing the normal JSON/SSE response.**
- [ ] **Step 3: Ensure cancellation propagates into the execution context and that failed members are represented in the final diagnostics.**
- [ ] **Step 4: Run all `ensemble_chat_completions` tests.**

Expected result: one execution can feed both the normal handler and an administrator-facing progress stream.

### Task 3: Add an Authenticated Administrative Test Stream

**Files:**
- Create or modify: `backend/internal/handler/admin/ensemble_test_handler.go`
- Modify: `backend/internal/handler/admin/group_handler.go` or its Ensemble handler companion
- Modify: the admin route registration file where `EnsembleHandler` routes are wired
- Modify: `frontend/src/api/admin/ensemble.ts`
- Test: `backend/internal/handler/admin/*ensemble*_test.go`

**Interfaces:**
- `POST /api/v1/admin/groups/:id/ensemble-test` accepts a test API key, messages, and optional model.
- The response is authenticated SSE with stable event types: `started`, `member_started`, `member_finished`, `aggregator_started`, `completed`, and `error`.
- The endpoint invokes the same execution and billing path as the public gateway request.

- [ ] **Step 1: Add handler tests for admin authentication, malformed input, SSE headers, and terminal completion/error events.**
- [ ] **Step 2: Implement the endpoint with request cancellation and a bounded event channel.**
- [ ] **Step 3: Wire the endpoint and expose a typed frontend SSE client using `AbortController`.**
- [ ] **Step 4: Run focused admin handler tests.**

Expected result: the page can show actual member progress rather than a timer pretending progress.

### Task 4: Make Group Lifecycle Explicit and Clean

**Files:**
- Modify: `backend/internal/repository/group_repo.go`
- Modify: `backend/internal/service/admin_group.go` if a friendly duplicate-name check belongs above the repository
- Modify: `backend/internal/handler/admin/group_handler.go` if a dedicated Ensemble lifecycle payload is needed
- Test: repository/service/handler group tests

**Interfaces:**
- Existing normal group lifecycle semantics remain intact.
- Ensemble deletion additionally cleans Ensemble-specific joins and members.
- Duplicate active names return the existing conflict error in a user-readable form.

- [ ] **Step 1: Add tests proving active duplicate Ensemble names fail without creating a second group.**
- [ ] **Step 2: Add tests proving deletion clears `ensemble_proposers`, `channel_groups`, and `account_groups` for the target group while preserving unrelated groups.**
- [ ] **Step 3: Add the cleanup statements inside the existing deletion transaction and invalidate the group API-key cache after commit.**
- [ ] **Step 4: Run repository/service group tests.**

Expected result: deleting an Ensemble group leaves no runnable member plan or billing-channel association.

### Task 5: Rework Ensemble Configuration Page Semantics

**Files:**
- Modify: `frontend/src/views/admin/EnsembleConfigView.vue`
- Modify: `frontend/src/utils/ensemble.ts`
- Modify: `frontend/src/api/admin/ensemble.ts`
- Test: focused Vue tests if this view has a test harness

**Interfaces:**
- New mode creates a group with `POST` only.
- Edit mode updates the selected group only.
- Save-as-new resets the target ID and creates a new group.

- [ ] **Step 1: Add a failing UI test or pure-function test for create mode not calling `groupsAPI.update`.**
- [ ] **Step 2: Add explicit new/edit mode state and a save-as-new action.**
- [ ] **Step 3: Close source/model pickers after selection and on outside click/Escape.**
- [ ] **Step 4: Keep the public model list as `ensemble`, while showing internal members separately.**
- [ ] **Step 5: Render missing metadata as a routing/configuration warning instead of `0/0`.**
- [ ] **Step 6: Run frontend typecheck and focused tests.**

Expected result: choosing an existing group cannot accidentally overwrite it when the user intends to create a new one.

### Task 6: Add the Real Progress Modal

**Files:**
- Create: `frontend/src/components/admin/ensemble/EnsembleTestDialog.vue`
- Modify: `frontend/src/views/admin/EnsembleConfigView.vue`
- Modify: `frontend/src/api/admin/ensemble.ts`
- Modify: relevant Chinese locale file if new shared labels are introduced

**Interfaces:**
- Dialog accepts the test API key, group name, and selected members.
- Dialog displays event status, duration, input/output tokens, cost, errors, candidate content, final content, and cancellation state.

- [ ] **Step 1: Add component tests for running, partial failure, fallback, success, and cancellation states.**
- [ ] **Step 2: Implement the dialog with the project `BaseDialog` and existing icon/select styles.**
- [ ] **Step 3: Connect the SSE client, update rows by stable member key, and close/retain the final result deliberately.**
- [ ] **Step 4: Run frontend typecheck, unit tests, and production build.**

Expected result: an administrator can identify whether the request is waiting on a proposer, aggregator, upstream rate limit, timeout, or routing failure.

### Task 7: Verification and Local Deployment

**Files:**
- No source changes unless a verification failure requires a targeted fix.

- [ ] **Step 1: Run Go focused tests, then the backend test/build command used by this repository.**
- [ ] **Step 2: Run frontend typecheck and build.**
- [ ] **Step 3: Rebuild the local Docker image and restart only the local stack.**
- [ ] **Step 4: Verify `/health`, container status, migration status, and admin Ensemble endpoints.**
- [ ] **Step 5: Report the exact untested paid path and wait for the user to trigger the real test from the UI.**

## Self-Review

- The plan covers the approved design's auth snapshot, diagnostics, lifecycle, candidate semantics, frontend UX, deletion cleanup, and verification requirements.
- No task relies on a placeholder implementation or on a client-only fake progress indicator.
- Public gateway compatibility is preserved by keeping diagnostics opt-in and keeping the normal response contract unchanged.
