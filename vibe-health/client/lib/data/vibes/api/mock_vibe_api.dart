import 'dart:async';
import 'dart:math';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../dto/vibe_dto.dart';
import 'vibe_api.dart';

/// Provider for the mock vibe API
final mockVibeApiProvider = Provider<VibeApi>((ref) {
  return MockVibeApi();
});

/// Mock implementation of the VibeApi for development and testing
class MockVibeApi implements VibeApi {
  // In-memory storage for mock data
  final List<VibeDto> _vibes = [];
  final _uuid = const Uuid();
  
  // Mock user ID
  final String _mockUserId = 'mock-user-123';
  
  MockVibeApi() {
    // Initialize with some sample data
    _initializeSampleData();
  }
  
  @override
  Future<VibeDto> createVibe(VibeDto vibe) async {
    // Simulate network delay
    await Future.delayed(const Duration(milliseconds: 300));
    
    // Generate a real UUID for the vibe
    final newVibe = VibeDto(
      id: _uuid.v4(),
      userId: _mockUserId,
      type: vibe.type,
      value: vibe.value,
      note: vibe.note,
      ts: DateTime.now().toUtc().toIso8601String(),
    );
    
    _vibes.add(newVibe);
    return newVibe;
  }
  
  @override
  Future<List<VibeDto>> listVibes({
    required DateTime from,
    required DateTime to,
  }) async {
    // Simulate network delay
    await Future.delayed(const Duration(milliseconds: 500));
    
    // Filter vibes by date range
    return _vibes.where((vibe) {
      final vibeDate = DateTime.parse(vibe.ts);
      return vibeDate.isAfter(from) && vibeDate.isBefore(to.add(const Duration(days: 1)));
    }).toList();
  }
  
  @override
  Future<VibeDto> getVibeById(String id) async {
    // Simulate network delay
    await Future.delayed(const Duration(milliseconds: 200));
    
    // Find vibe by ID
    final vibe = _vibes.firstWhere(
      (vibe) => vibe.id == id,
      orElse: () => throw Exception('Vibe not found with ID: $id'),
    );
    
    return vibe;
  }
  
  @override
  Future<VibeDto> updateVibe(String id, VibeDto vibe) async {
    // Simulate network delay
    await Future.delayed(const Duration(milliseconds: 300));
    
    // Find the vibe index
    final index = _vibes.indexWhere((v) => v.id == id);
    if (index == -1) {
      throw Exception('Vibe not found with ID: $id');
    }
    
    // Update the vibe
    final updatedVibe = VibeDto(
      id: id,
      userId: _mockUserId,
      type: vibe.type,
      value: vibe.value,
      note: vibe.note,
      ts: vibe.ts,
    );
    
    _vibes[index] = updatedVibe;
    return updatedVibe;
  }
  
  @override
  Future<void> deleteVibe(String id) async {
    // Simulate network delay
    await Future.delayed(const Duration(milliseconds: 200));
    
    // Remove the vibe
    _vibes.removeWhere((vibe) => vibe.id == id);
  }
  
  // Initialize with some sample data
  void _initializeSampleData() {
    final random = Random();
    final now = DateTime.now();
    
    // Add some mood vibes for the past week
    for (var i = 0; i < 7; i++) {
      final date = now.subtract(Duration(days: i));
      final value = random.nextInt(5) + 1; // 1-5 scale
      
      _vibes.add(VibeDto(
        id: _uuid.v4(),
        userId: _mockUserId,
        type: 'mood',
        value: value,
        note: value > 3 ? 'Had a good day' : 'Feeling a bit down',
        ts: DateTime(date.year, date.month, date.day, 
                     random.nextInt(12) + 8, // Random hour between 8 and 20
                     random.nextInt(60))     // Random minute
            .toUtc().toIso8601String(),
      ));
    }
    
    // Add some sleep vibes for the past week
    for (var i = 0; i < 7; i++) {
      final date = now.subtract(Duration(days: i));
      final value = random.nextInt(5) + 1; // 1-5 scale
      
      _vibes.add(VibeDto(
        id: _uuid.v4(),
        userId: _mockUserId,
        type: 'sleep',
        value: value,
        note: value > 3 ? 'Slept well' : 'Restless night',
        ts: DateTime(date.year, date.month, date.day, 7, random.nextInt(30))
            .toUtc().toIso8601String(),
      ));
    }
  }
}
