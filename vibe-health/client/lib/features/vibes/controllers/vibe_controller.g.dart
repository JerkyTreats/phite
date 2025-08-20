// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'vibe_controller.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

/// Controller for managing vibes
@ProviderFor(VibeController)
const vibeControllerProvider = VibeControllerProvider._();

/// Controller for managing vibes
final class VibeControllerProvider
    extends $NotifierProvider<VibeController, VibeState> {
  /// Controller for managing vibes
  const VibeControllerProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'vibeControllerProvider',
        isAutoDispose: true,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$vibeControllerHash();

  @$internal
  @override
  VibeController create() => VibeController();

  /// {@macro riverpod.override_with_value}
  Override overrideWithValue(VibeState value) {
    return $ProviderOverride(
      origin: this,
      providerOverride: $SyncValueProvider<VibeState>(value),
    );
  }
}

String _$vibeControllerHash() => r'a5cb88a80917d94767c7ccef5b441b4fdfe907fc';

abstract class _$VibeController extends $Notifier<VibeState> {
  VibeState build();
  @$mustCallSuper
  @override
  void runBuild() {
    final created = build();
    final ref = this.ref as $Ref<VibeState, VibeState>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<VibeState, VibeState>,
              VibeState,
              Object?,
              Object?
            >;
    element.handleValue(ref, created);
  }
}

// ignore_for_file: type=lint
// ignore_for_file: subtype_of_sealed_class, invalid_use_of_internal_member, invalid_use_of_visible_for_testing_member, deprecated_member_use_from_same_package
