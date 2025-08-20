import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../domain/vibe.dart';
import '../domain/vibe_type.dart';
import '../domain/vibe_error.dart';
import '../../../data/vibes/repositories/http_vibe_repository.dart';

part 'vibe_controller.g.dart';

/// State for the vibe controller
class VibeState {
  final bool isLoading;
  final List<Vibe> vibes;
  final VibeError? error;

  const VibeState({
    this.isLoading = false,
    this.vibes = const [],
    this.error,
  });

  /// Creates a copy of this state with the given fields replaced
  VibeState copyWith({
    bool? isLoading,
    List<Vibe>? vibes,
    VibeError? error,
  }) {
    return VibeState(
      isLoading: isLoading ?? this.isLoading,
      vibes: vibes ?? this.vibes,
      error: error,
    );
  }
}

/// Controller for managing vibes
@riverpod
class VibeController extends _$VibeController {
  @override
  VibeState build() {
    return const VibeState();
  }

  /// Creates a new vibe entry
  Future<void> createVibe(VibeType type, int value, {String? note}) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final repository = ref.read(vibeRepositoryProvider);
      
      // Create a new vibe with a temporary ID
      // The actual ID will be assigned by the server
      final vibe = Vibe(
        id: 'temp_${DateTime.now().millisecondsSinceEpoch}',
        userId: 'current_user', // This would come from auth service
        type: type,
        value: value,
        note: note,
        ts: DateTime.now().toUtc().toIso8601String(),
      );
      
      await repository.createVibe(vibe);
      
      // Refresh the list after creating
      await loadVibes();
    } on VibeError catch (e) {
      state = state.copyWith(isLoading: false, error: e);
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: VibeUnexpectedError(e.toString()),
      );
    }
  }

  /// Loads vibes for a specific date range
  Future<void> loadVibes({
    DateTime? from,
    DateTime? to,
  }) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final repository = ref.read(vibeRepositoryProvider);
      
      // Default to the current day if not specified
      final now = DateTime.now();
      final startDate = from ?? DateTime(now.year, now.month, now.day);
      final endDate = to ?? DateTime(now.year, now.month, now.day, 23, 59, 59);
      
      final vibes = await repository.listVibes(
        from: startDate,
        to: endDate,
      );
      
      state = state.copyWith(isLoading: false, vibes: vibes);
    } on VibeError catch (e) {
      state = state.copyWith(isLoading: false, error: e);
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: VibeUnexpectedError(e.toString()),
      );
    }
  }

  /// Updates an existing vibe
  Future<void> updateVibe(Vibe vibe) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final repository = ref.read(vibeRepositoryProvider);
      await repository.updateVibe(vibe);
      
      // Update the vibe in the local state
      final updatedVibes = [...state.vibes];
      final index = updatedVibes.indexWhere((v) => v.id == vibe.id);
      
      if (index != -1) {
        updatedVibes[index] = vibe;
        state = state.copyWith(isLoading: false, vibes: updatedVibes);
      } else {
        // If the vibe wasn't in our local state, refresh the list
        await loadVibes();
      }
    } on VibeError catch (e) {
      state = state.copyWith(isLoading: false, error: e);
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: VibeUnexpectedError(e.toString()),
      );
    }
  }

  /// Gets vibes for a specific type
  List<Vibe> getVibesByType(VibeType type) {
    return state.vibes.where((vibe) => vibe.type == type).toList();
  }

  /// Gets the most recent vibe for a specific type
  Vibe? getMostRecentVibe(VibeType type) {
    final typeVibes = getVibesByType(type);
    if (typeVibes.isEmpty) return null;
    
    // Sort by timestamp descending
    typeVibes.sort((a, b) => b.ts.compareTo(a.ts));
    return typeVibes.first;
  }

  /// Clears any error in the state
  void clearError() {
    if (state.error != null) {
      state = state.copyWith(error: null);
    }
  }
}
