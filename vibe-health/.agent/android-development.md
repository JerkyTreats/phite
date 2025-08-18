# Android Development

Flutter Fast Prototyping (Mac + VS Code + Physical Device)

Scope: Define tools and workflows to achieve near‑zero iteration delay for the vibe health Android prototype using Flutter.

Integrates with a Go backend over HTTP/WebSocket.

Includes push notifications, device mirroring, distribution, and testing.

## Executive Summary

- __Why Flutter__: Hot Reload, strong VS Code tooling, rapid UI assembly, cross‑platform optionality (Android first; iOS later).
- __Fast loop__: Physical device + Hot Reload + scrcpy mirroring + DevTools.
- __Push path__: Start with local notifications → add FCM HTTP v1 for real push (curl/Postman) → optional OneSignal/Beams.
- __Distribution__: Firebase App Distribution for testers; Codemagic for fast CI/CD when needed.

## Tooling Stack

| Tool | Purpose | Key commands/links |
|---|---|---|
| Flutter SDK + Dart | Build/run app; Hot Reload/DevTools | Docs: https://docs.flutter.dev/get-started/install |
| VS Code + extensions | IDE + Flutter/Dart tooling | Flutter, Dart, Error Lens; DevTools integrates |
| Android SDK/Platform Tools | ADB, platform images | Installed via Android Studio (SDK only) |
| Setup scripts | Bootstrap Flutter/Android SDK | client/scripts/setup_macos.sh • client/scripts/setup_linux.sh • client/scripts/setup_win.ps1 |

## Project Scaffolding & Packages

- __Scaffold__
  - `flutter create vibe_health`
  - Add flavors (dev, prod) for endpoints/keys.
- __Recommended Packages__

| Category | Packages |
|---|---|
| State | riverpod (selected) |
| HTTP | dio |
| JSON | json_serializable, build_runner |
| Notifications | firebase_messaging, flutter_local_notifications |
| Secure storage | flutter_secure_storage |
| Date/time & intl | intl |
| Deep links | uni_links (or firebase_dynamic_links) |

- __Structure__
  - `lib/` organized into `features/`, `core/`, `data/`, `services/` with dependency inversion (UI → state → repos → HTTP).

## State Management

Decision: Use Riverpod as the default state manager for this project.

| Option | Model | Boilerplate | Testability | Notable strengths | Caveats | Fit |
|---|---|---|---|---|---|---|
| Riverpod (selected) | DI‑first provider graph | Low–Medium | High | Compile‑time safety; `AsyncValue`; fine‑grained rebuilds | Requires provider graph discipline | Best overall (UI → state → repo → HTTP) |
| BLoC/Cubit | Event→State or simple state | Medium/Low | High | Explicit transitions; mature | More wiring than Riverpod | Good for explicit event flows |
| Provider | InheritedWidget wrapper | Low | Medium | Simple, common | Easier rebuild pitfalls; ad‑hoc DI | OK for small areas only |

## Firebase

### Why Firebase
- Push notifications via FCM for real-time prompts and background delivery.
- Crash reporting and stability via Crashlytics.
- Fast tester distribution via Firebase App Distribution.
- Optional analytics for UX iteration (can be disabled in dev).

### Prerequisites
- Google account with access to Firebase Console.
- Android app package name (e.g., `com.vibehealth.app`).
- SHA-1/SHA-256 fingerprints for debug/release (improves auth/FCM reliability): run `./gradlew signingReport` in `android/`.

### Project Setup
- Create Firebase project and add Android app in Firebase Console.
- Download `google-services.json` and place it at `android/app/google-services.json`.
- Ensure Android uses a Google Play Services system image for emulator testing.

### Flutter Packages
- Add core SDKs:
  - `flutter pub add firebase_core firebase_messaging firebase_crashlytics`
  - Optional: `flutter pub add firebase_analytics`

### Android Gradle Integration
- In `android/build.gradle` add Google Services classpath in `dependencies`:
  - `classpath "com.google.gms:google-services:4.4.2"`
- In `android/app/build.gradle` apply the plugin at the bottom:
  - `apply plugin: "com.google.gms.google-services"`

### Initialization
- Ensure `Firebase.initializeApp()` runs before using Firebase SDKs (Flutter default templates often handle this).

### Push Messaging Notes
- FCM testing workflow is detailed in `## Push Notifications (Development)`.
- For HTTP v1 tests: create a service account, enable the API, and follow the curl/Postman flow referenced there.

### Distribution Notes
- Tester rollout via `## Builds, Distribution, CI/CD → Firebase App Distribution`.

## Push Notifications (Development)

| Path | Packages | Key steps | Links |
|---|---|---|---|
| Local (fastest) | flutter_local_notifications | Request `POST_NOTIFICATIONS` (Android 13+); create channel; schedule test notifications | Permission: https://developer.android.com/develop/ui/views/notifications/notification-permission |
| FCM HTTP v1 | firebase_messaging | Integrate SDK; retrieve device token; send test via OAuth-authenticated curl/Postman; handle fg/bg | First message: https://firebase.google.com/docs/cloud-messaging/android/first-message • Send: https://firebase.google.com/docs/cloud-messaging/send-message • How‑to: https://apoorv487.medium.com/testing-fcm-push-notification-http-v1-through-oauth-2-0-playground-postman-terminal-part-2-7d7a6a0e2fa0 |
| Alternatives | OneSignal, Pusher Beams | Optional vendor‑managed push | OneSignal: https://documentation.onesignal.com/reference/push-notification • Beams: https://pusher.com/beams/ |


## Environment & Networking

- __Local dev__
  - See “Testing & Feedback → Local Networking.” Summary: use `adb reverse tcp:8080 tcp:8080` or configure API base as LAN IP. Ensure reachability (e.g., Tailscale if applicable).
- __Config per flavor__
  - Use `--dart-define` or env files to inject API base URLs and flags per flavor.

## Builds, Distribution, CI/CD

- __Builds__
  - Debug for dev; sign release when distributing beyond testers.
- __Firebase App Distribution__
  - Fast tester onboarding, install links, crash metrics (with Crashlytics).
  - Docs: https://firebase.google.com/docs/app-distribution
- __CI/CD (when ready)__
  - Codemagic: fast Flutter builds, easy setup, caching.
    https://codemagic.io/start/
  - Bitrise (Android Gradle caching, scalable):
    https://bitrise.io/

## Testing & Feedback

- Local Device & Emulator (VS Code)
  - Use the VS Code device picker (status bar) to select a target device or Android Emulator.
  - Launch emulator via Command Palette: “Flutter: Launch Emulator”.
  - CLI: `flutter emulators`, `flutter emulators --launch <id>`.
  - Hot Reload/Restart supported; open Flutter DevTools from VS Code (inspector, layout, perf/memory).
- Physical Device
  - Enable Developer Options and USB debugging.
  - Verify device: USB `adb devices`; Wi‑Fi: `adb tcpip 5555` → `adb connect <device_ip>:5555`.
  - Mirror/control: `scrcpy` (USB or Wi‑Fi). Useful flags: `--record out.mp4`, `--bit-rate 8M`, `--max-size 1080`.
- Local Networking
  - Backend on Mac localhost: `adb reverse tcp:8080 tcp:8080`.
  - Or use LAN IP via `--dart-define=API_BASE_URL`.
- Notifications Testing
  - Local: request `POST_NOTIFICATIONS` (Android 13+), create channels; schedule test notifications.
  - FCM HTTP v1: retrieve device token; send via curl/Postman with OAuth token; verify fg/bg handlers.
  - Quick‑reply actions: confirm emoji replies post without launching app; check background isolate logs.
- Deep Links
  - Scheme: `vibehealth://vibe/new?type=sleep|mood&prefill=<value?>`.
  - Launch via ADB: `adb shell am start -a android.intent.action.VIEW -d "vibehealth://vibe/new?type=sleep"`.
- Auth & Storage
  - Tokens stored with `flutter_secure_storage`; auth interceptor injects token on device builds.
- Performance & Logs
  - Enable perf overlays in DevTools; profile jank.
  - Logs: `adb logcat | grep -i flutter` (or tag‑specific); ensure crashes surface in logcat.
- Troubleshooting
  - Wi‑Fi ADB drops: `adb connect <device_ip>:5555` again.
  - Notifications missing: check channel importance, OS permission, battery optimizations.
  - Deep links fail: validate intent filter/scheme; retry with explicit `adb shell am start`.
- Tools for Testing & Feedback
  - Maestro (flow-based E2E; simple authoring; local or cloud): https://maestro.dev/
  - Instabug (in-app feedback, crash, performance): https://www.instabug.com
- Unit & Widget tests
  - Use `flutter_test` and `integration_test` for critical flows.

## Security Notes

- __Tokens__: store in `flutter_secure_storage`; never in plaintext prefs.
- __Permissions__: request `POST_NOTIFICATIONS` on Android 13+; declare minimal permissions only.
- __Networking__: TLS everywhere; consider certificate pinning later.


