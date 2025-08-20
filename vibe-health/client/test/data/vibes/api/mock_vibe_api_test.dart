import 'package:flutter_test/flutter_test.dart';
import 'package:vibe_health/data/vibes/api/mock_vibe_api.dart';
import 'package:vibe_health/data/vibes/dto/vibe_dto.dart';

void main() {
  group('MockVibeApi', () {
    late MockVibeApi mockVibeApi;

    setUp(() {
      mockVibeApi = MockVibeApi();
    });

    test('listVibes returns sample data', () async {
      // Get vibes for the current day
      final now = DateTime.now();
      final startOfDay = DateTime(now.year, now.month, now.day);
      final endOfDay = DateTime(now.year, now.month, now.day, 23, 59, 59);

      final vibes = await mockVibeApi.listVibes(
        from: startOfDay,
        to: endOfDay,
      );

      // Should have some sample data
      expect(vibes, isNotEmpty);
    });

    test('createVibe adds a new vibe', () async {
      // Create a new vibe
      final newVibe = VibeDto(
        id: 'temp-id', // Will be replaced by the API
        userId: 'test-user',
        type: 'mood',
        value: 4,
        note: 'Test note',
        ts: DateTime.now().toUtc().toIso8601String(),
      );

      // Get the initial count
      final now = DateTime.now();
      final startOfDay = DateTime(now.year, now.month, now.day);
      final endOfDay = DateTime(now.year, now.month, now.day, 23, 59, 59);
      final initialVibes = await mockVibeApi.listVibes(
        from: startOfDay,
        to: endOfDay,
      );
      final initialCount = initialVibes.length;

      // Create the vibe
      final createdVibe = await mockVibeApi.createVibe(newVibe);

      // Verify the vibe was created with a real UUID
      expect(createdVibe.id, isNot('temp-id'));
      expect(createdVibe.value, 4);
      expect(createdVibe.note, 'Test note');

      // Verify the vibe was added to the list
      final updatedVibes = await mockVibeApi.listVibes(
        from: startOfDay,
        to: endOfDay,
      );
      expect(updatedVibes.length, initialCount + 1);
    });

    test('updateVibe updates an existing vibe', () async {
      // Create a vibe to update
      final newVibe = VibeDto(
        id: 'temp-id',
        userId: 'test-user',
        type: 'mood',
        value: 3,
        note: 'Original note',
        ts: DateTime.now().toUtc().toIso8601String(),
      );

      final createdVibe = await mockVibeApi.createVibe(newVibe);
      final vibeId = createdVibe.id;
      
      // Ensure vibeId is not null
      expect(vibeId, isNotNull);
      
      // Update the vibe
      final updatedVibe = VibeDto(
        id: vibeId!,
        userId: 'test-user',
        type: 'mood',
        value: 5,
        note: 'Updated note',
        ts: createdVibe.ts,
      );

      final result = await mockVibeApi.updateVibe(vibeId, updatedVibe);

      // Verify the update
      expect(result.id, vibeId);
      expect(result.value, 5);
      expect(result.note, 'Updated note');

      // Verify the update is reflected in the list
      final vibe = await mockVibeApi.getVibeById(vibeId);
      expect(vibe.value, 5);
      expect(vibe.note, 'Updated note');
    });

    test('deleteVibe removes a vibe', () async {
      // Create a vibe to delete
      final newVibe = VibeDto(
        id: 'temp-id',
        userId: 'test-user',
        type: 'mood',
        value: 3,
        note: 'To be deleted',
        ts: DateTime.now().toUtc().toIso8601String(),
      );

      final createdVibe = await mockVibeApi.createVibe(newVibe);
      final vibeId = createdVibe.id;
      
      // Ensure vibeId is not null
      expect(vibeId, isNotNull);

      // Verify it exists
      final vibe = await mockVibeApi.getVibeById(vibeId!);
      expect(vibe.id, vibeId);

      // Delete the vibe
      await mockVibeApi.deleteVibe(vibeId);

      // Verify it was deleted
      expect(() async => await mockVibeApi.getVibeById(vibeId), 
             throwsA(isA<Exception>()));
    });
  });
}
