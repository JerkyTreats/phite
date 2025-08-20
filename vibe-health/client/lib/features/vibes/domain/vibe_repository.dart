import 'package:flutter/foundation.dart';
import 'vibe.dart';

/// Repository interface for vibe operations
abstract class VibeRepository {
  /// Creates a new vibe entry
  /// 
  /// Returns a [Future] that completes with void when the operation is successful
  /// or throws a [VibeError] when the operation fails
  Future<void> createVibe(Vibe vibe);
  
  /// Lists vibes within the specified date range
  /// 
  /// [from] - Start date (inclusive)
  /// [to] - End date (inclusive)
  /// 
  /// Returns a [Future] that completes with a list of [Vibe] objects
  /// or throws a [VibeError] when the operation fails
  Future<List<Vibe>> listVibes({
    required DateTime from,
    required DateTime to,
  });
  
  /// Gets a vibe by its ID
  /// 
  /// [id] - The unique identifier of the vibe
  /// 
  /// Returns a [Future] that completes with the [Vibe] object
  /// or throws a [VibeError] when the operation fails
  Future<Vibe> getVibeById(String id);
  
  /// Updates an existing vibe
  /// 
  /// [vibe] - The vibe to update
  /// 
  /// Returns a [Future] that completes with void when the operation is successful
  /// or throws a [VibeError] when the operation fails
  Future<void> updateVibe(Vibe vibe);
  
  /// Deletes a vibe by its ID
  /// 
  /// [id] - The unique identifier of the vibe to delete
  /// 
  /// Returns a [Future] that completes with void when the operation is successful
  /// or throws a [VibeError] when the operation fails
  @visibleForTesting
  Future<void> deleteVibe(String id);
}
