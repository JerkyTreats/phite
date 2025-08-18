# Agent Brief: Vibe Check Feature (Flutter)

## Objective
- Implement the “Vibe Check” for the Android prototype with strict dependency inversion per `vibe-health/.agent/android-development.md`.
- Deliver scheduled prompts (morning Sleep Check, afternoon Mood Check), notification quick‑reply emojis that post without opening the app, and a ➕ action that deep‑links into the Vibe Entry screen.
- Integrate with the Go backend (`POST /vibes`, `GET /vibes?from=&to=`), with secure token handling, local notifications first, then FCM HTTP v1.

## Core Architecture

- Dependency direction: UI → state → repository interface → repository implementation → HTTP/IO. Higher‑level layers depend on abstractions, not concrete details.

- Directory tree (proposed):
```text
lib/
  core/
    di/
      providers.dart              // Composition root (Dio, Auth, Notification, VibeApi, VibeRepository)
    config/
      config.dart                 // --dart-define surfaced as providers (API_BASE_URL, flags)
    errors/
      failures.dart               // Shared/typed failures
    routing/
      routes.dart                 // App routes + deep-link mapping
    time/
      time_utils.dart             // ISO8601 / TZ helpers
  services/
    http/
      dio_client.dart             // Configured Dio instance
      auth_interceptor.dart       // Injects token from secure storage
    notifications/
      notification_service.dart   // Local + FCM init, channels, schedule, action handlers
    storage/
      secure_storage.dart         // Token read/write lifecycle
    vibes/
      vibe_api.dart               // POST /vibes, GET /vibes
  data/
    vibes/
      dto/
        vibe_dto.dart             // Transport model
      mappers/
        vibe_mapper.dart          // DTO <-> Entity conversion
      http_vibe_repository.dart   // Implements VibeRepository using VibeApi
  features/
    vibes/
      domain/
        vibe.dart                 // Entity
        vibe_type.dart            // enum {sleep, mood}
        vibe_repository.dart      // Abstract repository interface
        vibe_error.dart           // Feature-scoped error type
      application/
        vibe_list_controller.dart   // Riverpod notifier (list today’s vibes)
        vibe_create_controller.dart // Riverpod notifier (create vibe)
        providers.dart              // Feature-level providers
      presentation/
        home_screen.dart            // Today summary, entry points
        vibe_entry_screen.dart      // Emoji grid, note, submit
        widgets/
          emoji_bar.dart            // Reusable emoji row with a11y
  app.dart                           // MaterialApp, route setup
  main_dev.dart                      // Dev flavor
  main_prod.dart                     // Prod flavor
```

## Data Model

- Entity: `Vibe`
  - `id: string`
  - `userId: string`
  - `type: "sleep" | "mood"`
  - `value: int` (emoji scale; see mapping)
  - `note?: string`
  - `ts: string` (UTC ISO8601)
- Emoji value mapping (default, configurable):
  - 😁 → 5, 🙂 → 4, 😐 → 3, 🙁 → 2, 😡 → 1
- API (Go backend)
  - Create: `POST /vibes` body `{ type, value, note?, ts }`
  - List: `GET /vibes?from=&to=` returns `Vibe[]`
- Error normalization: transport errors mapped to domain failures (`AuthError`, `NetworkError`, `ServerError`, `ParsingError`).

## Notifications

- Phase A: Local notifications
  - Android 13+ permission `POST_NOTIFICATIONS`
  - Channel: `vibes_prompt` (importance: high)
  - Schedules:
    - Sleep Check: 07:30 local
    - Mood Check: 15:30 local
  - Actions:
    - Quick‑reply emojis (5 actions) → background handler posts to `POST /vibes`
    - ➕ action → deep‑link: `vibehealth://vibe/new?type=sleep|mood`
- Phase B: FCM HTTP v1
  - `firebase_messaging` integration, device token registration with backend
  - Data payload:
    - `action=quick_vibe`, `type`, `value`, `ts` (optional), `deeplink`
  - Background/foreground handlers mirror local flow

## Deep Links and Navigation

- Scheme: `vibehealth://`
- Routes:
  - `vibehealth://vibe/new?type=sleep|mood&prefill=<value?>`
- Behavior:
  - If `prefill` present, preselect emoji; focus note field
  - Works from background/terminated states via route resolver

## Configuration

- `--dart-define` per flavor:
  - `API_BASE_URL`
  - `ENABLE_LOCAL_NOTIFS` (bool)
  - `ENABLE_FCM` (bool)
  - `VIBE_AM_SCHEDULE` (e.g., `07:30`)
  - `VIBE_PM_SCHEDULE` (e.g., `15:30`)
- Dev networking: `adb reverse tcp:8080 tcp:8080` or LAN URL

## Technical Specifications

Decision: Use Riverpod as the default state manager for this feature.

- Providers (Riverpod):
  - `dioProvider`, `authTokenProvider`, `vibeApiProvider`, `vibeRepositoryProvider`
  - `vibeListControllerProvider(DateTimeRange)`, `vibeCreateControllerProvider`
  - `notificationServiceProvider`, config providers
- HTTP:
  - Dio BaseOptions from config; auth interceptor pulls token from `flutter_secure_storage`
  - Errors mapped to domain failures
- Notification Service:
  - `init()`, permission prompt, channel creation
  - `scheduleDaily(name, time)`; `handleAction(action, payload)`
- Repository Interface (`features/vibes/domain/vibe_repository.dart`):
  - `Future<void> createVibe(Vibe v)`
  - `Future<List<Vibe>> listVibes(DateTimeRange range)`
- Repository Implementation (`data/vibes/http_vibe_repository.dart`):
  - Uses `services/vibes/vibe_api.dart`
  - DTO mapping and error normalization
- Time:
  - Store and send UTC ISO8601; convert to local for UI

## Implementation Plan & Status

| Phase | Step | Task | Files | Status |
|------|------|------|-------|--------|
| Core Setup | 1.1 | Define `Vibe`, `VibeType`, `VibeRepository`, `VibeError` | `features/vibes/domain/*` | TODO |
|  | 1.2 | Create DTOs and mappers | `data/vibes/dto/*`, `data/vibes/mappers/*` | TODO |
|  | 1.3 | DI providers | `core/di/providers.dart` | TODO |
| Notifications A (Local) | 2.1 | Permission + channel creation | `services/notifications/notification_service.dart` | TODO |
|  | 2.2 | Daily schedules (AM/PM) | same as above | TODO |
|  | 2.3 | Quick‑reply actions + handler | same as above | TODO |
|  | 2.4 | ➕ deep‑link wiring | `core/routing/routes.dart`, `features/vibes/presentation/vibe_entry_screen.dart` | TODO |
| HTTP/Repo | 3.1 | Dio base + auth interceptor | `services/http/*` | TODO |
|  | 3.2 | VibeApi endpoints | `services/vibes/vibe_api.dart` | TODO |
|  | 3.3 | HttpVibeRepository | `data/vibes/http_vibe_repository.dart` | TODO |
| State & UI | 4.1 | Controllers (list/create) | `features/vibes/application/*` | TODO |
|  | 4.2 | Home + Vibe Entry screens | `features/vibes/presentation/*` | TODO |
|  | 4.3 | Emoji component + A11y | `features/vibes/presentation/widgets/*` | TODO |
| FCM (Phase B) | 5.1 | Firebase init + token | `services/notifications/notification_service.dart` | TODO |
|  | 5.2 | Background/foreground handlers | same as above | TODO |
|  | 5.3 | Backend token registration | `services/vibes/vibe_api.dart` (optional) | TODO |
| Config & Flags | 6.1 | `--dart-define` wiring | `core/config/config.dart` | TODO |
|  | 6.2 | Feature toggles in DI | `core/di/providers.dart` | TODO |
| QA & Polish | 7.1 | Error/empty states | UI files | TODO |
|  | 7.2 | Haptics, theming basics | UI files | TODO |

## Testing Strategy

- Unit
  - Controllers with fake `VibeRepository` (success/error, optimistic update)
  - DTO ↔ Entity mapping; error normalization
- Integration
  - `HttpVibeRepository` against mock server
  - Notification action handler → repository creates vibe with correct payload
- E2E (Maestro)
  - Scheduled prompt fires
  - Quick‑reply posts vibe without opening app
  - ➕ deep‑link opens Vibe Entry; submitting creates vibe and returns Home
- Handlers
  - Foreground/background payload handling with test data
- Accessibility
  - Emoji component semantics/labels

## Success Criteria

- Scheduled prompts (local or FCM) present emoji quick actions and complete end‑to‑end posting.
- ➕ deep‑link opens Vibe Entry and supports prefill; submit creates vibe and updates Home.
- UI/state import no HTTP packages; repository abstraction boundary enforced.
- Clear user‑facing error handling; offline failure modes defined.

## Risks and Assumptions

- Notification action delivery may vary across OEMs; ensure idempotency on backend.
- Timezone and DST edge cases addressed by using UTC storage and local display.
- Background execution limits: test quick‑reply posting across Android versions.

## Open Questions

- Final emoji mapping (values, labels, localization).
- Long‑term schedule source of truth: local vs server (FCM).
- Analytics/telemetry events, if any.

## References

- `vibe-health/.agent/android-development.md`
- `vibe-health/.agent/client.md`
- Firebase Messaging docs
- flutter_local_notifications docs
