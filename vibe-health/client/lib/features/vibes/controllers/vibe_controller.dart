import 'dart:async';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../domain/vibe.dart';
import '../domain/vibe_type.dart';
import '../domain/vibe_error.dart';
import '../../../data/vibes/repositories/http_vibe_repository.dart';

part 'vibe_controller.g.dart';

/// Enum representing the various operations that can be performed on vibes
enum VibeOperation {
  loading,
  creating,
  updating,
  deleting,
  none
}

/// State for the vibe controller
class VibeState {
  /// Current operation being performed
  final VibeOperation operation;
  
  /// List of vibes
  final List<Vibe> vibes;
  
  /// Error that occurred during an operation, if any
  final VibeError? error;
  
  /// Whether any operation is in progress
  bool get isLoading => operation != VibeOperation.none;
  
  /// Whether a specific vibe is being created
  bool get isCreating => operation == VibeOperation.creating;
  
  /// Whether vibes are being loaded
  bool get isLoadingVibes => operation == VibeOperation.loading;
  
  /// Whether a specific vibe is being updated
  bool get isUpdating => operation == VibeOperation.updating;
  
  /// Whether a specific vibe is being deleted
  bool get isDeleting => operation == VibeOperation.deleting;

  const VibeState({
    this.operation = VibeOperation.none,
    this.vibes = const [],
    this.error,
  });

  /// Creates a copy of this state with the given fields replaced
  VibeState copyWith({
    VibeOperation? operation,
    List<Vibe>? vibes,
    VibeError? error,
  }) {
    return VibeState(
      operation: operation ?? this.operation,
      vibes: vibes ?? this.vibes,
      error: error,
    );
  }
}

/// Controller for managing vibes
@riverpod
class VibeController extends _$VibeController {
  // Track the last operation for retry functionality
  VibeOperation _lastFailedOperation = VibeOperation.none;
  Vibe? _lastCreateAttempt;
  @override
  VibeState build() {
    return const VibeState();
  }

  /// Creates a new vibe entry
  Future<void> createVibe(VibeType type, int value, {String? note}) async {
    state = state.copyWith(operation: VibeOperation.creating, error: null);

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
      
      // Store the vibe for potential retry
      _lastCreateAttempt = vibe;
      
      await repository.createVibe(vibe);
      
      // Refresh the list after creating
      await loadVibes();
      
      // Clear last operation on success
      _lastFailedOperation = VibeOperation.none;
      _lastCreateAttempt = null;
    } on VibeError catch (e) {
      _lastFailedOperation = VibeOperation.creating;
      state = state.copyWith(operation: VibeOperation.none, error: e);
    } catch (e) {
      _lastFailedOperation = VibeOperation.creating;
      state = state.copyWith(
        operation: VibeOperation.none,
        error: VibeUnexpectedError(e.toString()),
      );
    }
  }

  /// Loads vibes for a specific date range
  Future<void> loadVibes({
    DateTime? from,
    DateTime? to,
    bool forceRefresh = false,
  }) async {
    // Skip loading if we already have data and aren't forcing a refresh
    if (!forceRefresh && state.vibes.isNotEmpty && !state.isLoading) {
      return;
    }
    
    state = state.copyWith(operation: VibeOperation.loading, error: null);

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
      
      state = state.copyWith(operation: VibeOperation.none, vibes: vibes);
      
      // Clear last operation on success
      _lastFailedOperation = VibeOperation.none;
    } on VibeError catch (e) {
      _lastFailedOperation = VibeOperation.loading;
      state = state.copyWith(operation: VibeOperation.none, error: e);
    } catch (e) {
      _lastFailedOperation = VibeOperation.loading;
      state = state.copyWith(
        operation: VibeOperation.none,
        error: VibeUnexpectedError(e.toString()),
      );
    }
  }

  /// Updates an existing vibe
  Future<void> updateVibe(Vibe vibe) async {
    state = state.copyWith(operation: VibeOperation.updating, error: null);

    try {
      final repository = ref.read(vibeRepositoryProvider);
      await repository.updateVibe(vibe);
      
      // Update the vibe in the local state
      final updatedVibes = [...state.vibes];
      final index = updatedVibes.indexWhere((v) => v.id == vibe.id);
      
      if (index != -1) {
        updatedVibes[index] = vibe;
        state = state.copyWith(operation: VibeOperation.none, vibes: updatedVibes);
      } else {
        // If the vibe wasn't in our local state, refresh the list
        await loadVibes(forceRefresh: true);
      }
      
      // Clear last operation on success
      _lastFailedOperation = VibeOperation.none;
    } on VibeError catch (e) {
      _lastFailedOperation = VibeOperation.updating;
      state = state.copyWith(operation: VibeOperation.none, error: e);
    } catch (e) {
      _lastFailedOperation = VibeOperation.updating;
      state = state.copyWith(
        operation: VibeOperation.none,
        error: VibeUnexpectedError(e.toString()),
      );
    }
  }
  
  /// Removes a vibe from the local state
  /// 
  /// This is a production-safe alternative to deleteVibe that only removes
  /// the vibe from the local state without calling the repository's deleteVibe method
  /// which is marked as @visibleForTesting.
  Future<void> removeVibeFromState(String id) async {
    state = state.copyWith(operation: VibeOperation.deleting, error: null);
    
    try {
      // Only remove from local state without calling the API
      final updatedVibes = state.vibes.where((vibe) => vibe.id != id).toList();
      state = state.copyWith(operation: VibeOperation.none, vibes: updatedVibes);
      
      // Clear last operation on success
      _lastFailedOperation = VibeOperation.none;
    } catch (e) {
      _lastFailedOperation = VibeOperation.deleting;
      state = state.copyWith(
        operation: VibeOperation.none,
        error: VibeUnexpectedError(e.toString()),
      );
    }
  }
  
  /// Deletes a vibe by ID - FOR TESTING ONLY
  /// 
  /// Note: This method uses the repository's deleteVibe method which is marked
  /// as @visibleForTesting. This should only be used in tests.
  @visibleForTesting
  Future<void> deleteVibe(String id) async {
    state = state.copyWith(operation: VibeOperation.deleting, error: null);
    
    try {
      final repository = ref.read(vibeRepositoryProvider);
      // Using a method marked as visibleForTesting - only for development/testing
      await repository.deleteVibe(id);
      
      // Remove the vibe from the local state
      await removeVibeFromState(id);
      
      // Clear last operation on success
      _lastFailedOperation = VibeOperation.none;
    } on VibeError catch (e) {
      _lastFailedOperation = VibeOperation.deleting;
      state = state.copyWith(operation: VibeOperation.none, error: e);
    } catch (e) {
      _lastFailedOperation = VibeOperation.deleting;
      state = state.copyWith(
        operation: VibeOperation.none,
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
  
  /// Retries the last failed operation
  Future<void> retry() async {
    if (state.error == null) return;
    
    // Clear the error
    state = state.copyWith(error: null);
    
    // Retry based on the last failed operation
    switch (_lastFailedOperation) {
      case VibeOperation.creating:
        if (_lastCreateAttempt != null) {
          final repository = ref.read(vibeRepositoryProvider);
          await repository.createVibe(_lastCreateAttempt!);
          await loadVibes(forceRefresh: true);
        } else {
          // If we don't have the last create attempt, just refresh
          await loadVibes(forceRefresh: true);
        }
        break;
        
      case VibeOperation.updating:
        // For now, just refresh the list as we don't track the last update attempt
        await loadVibes(forceRefresh: true);
        break;
        
      case VibeOperation.deleting:
        // For now, just refresh the list as we don't track the last delete attempt
        await loadVibes(forceRefresh: true);
        break;
        
      case VibeOperation.loading:
      case VibeOperation.none:
        // Default to loading vibes
        await loadVibes(forceRefresh: true);
        break;
    }
  }

  /// Clears any error in the state
  void clearError() {
    if (state.error != null) {
      state = state.copyWith(error: null);
    }
  }
}
