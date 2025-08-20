import 'package:flutter_test/flutter_test.dart';
import 'package:mockito/annotations.dart';
import 'package:mockito/mockito.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:vibe_health/features/vibes/controllers/vibe_controller.dart';
import 'package:vibe_health/features/vibes/domain/vibe.dart';
import 'package:vibe_health/features/vibes/domain/vibe_type.dart';
import 'package:vibe_health/features/vibes/domain/vibe_error.dart';
import 'package:vibe_health/features/vibes/domain/vibe_repository.dart';
import 'package:vibe_health/data/vibes/repositories/http_vibe_repository.dart';

import 'vibe_controller_test.mocks.dart';

@GenerateMocks([VibeRepository])
void main() {
  late MockVibeRepository mockRepository;
  late ProviderContainer container;

  // Sample test data
  final testVibe = Vibe(
    id: 'test-id-1',
    userId: 'test-user',
    type: VibeType.mood,
    value: 4,
    note: 'Test note',
    ts: '2025-08-20T12:00:00Z',
  );

  final testVibes = [
    testVibe,
    Vibe(
      id: 'test-id-2',
      userId: 'test-user',
      type: VibeType.sleep,
      value: 3,
      note: null,
      ts: '2025-08-20T08:00:00Z',
    ),
  ];

  setUp(() {
    mockRepository = MockVibeRepository();
    container = ProviderContainer(
      overrides: [
        vibeRepositoryProvider.overrideWithValue(mockRepository),
      ],
    );
  });

  tearDown(() {
    container.dispose();
  });

  group('VibeController', () {
    test('initial state is correct', () {
      final controller = container.read(vibeControllerProvider.notifier);
      final state = container.read(vibeControllerProvider);
      
      expect(state.isLoading, false);
      expect(state.vibes, isEmpty);
      expect(state.error, isNull);
    });

    test('loadVibes updates state correctly on success', () async {
      // Arrange
      when(mockRepository.listVibes(
        from: anyNamed('from'),
        to: anyNamed('to'),
      )).thenAnswer((_) async => testVibes);

      // Act
      await container.read(vibeControllerProvider.notifier).loadVibes();
      
      // Assert
      final state = container.read(vibeControllerProvider);
      expect(state.isLoading, false);
      expect(state.vibes, testVibes);
      expect(state.error, isNull);
    });

    test('loadVibes updates state correctly on error', () async {
      // Arrange
      final testError = VibeServerError(message: 'Server error');
      when(mockRepository.listVibes(
        from: anyNamed('from'),
        to: anyNamed('to'),
      )).thenThrow(testError);

      // Act
      await container.read(vibeControllerProvider.notifier).loadVibes();
      
      // Assert
      final state = container.read(vibeControllerProvider);
      expect(state.isLoading, false);
      expect(state.vibes, isEmpty);
      expect(state.error, testError);
    });

    test('createVibe calls repository and refreshes vibes', () async {
      // Arrange
      when(mockRepository.createVibe(any)).thenAnswer((_) async {});
      when(mockRepository.listVibes(
        from: anyNamed('from'),
        to: anyNamed('to'),
      )).thenAnswer((_) async => testVibes);

      // Act
      await container.read(vibeControllerProvider.notifier).createVibe(
        VibeType.mood,
        4,
        note: 'Test note',
      );
      
      // Assert
      verify(mockRepository.createVibe(any)).called(1);
      verify(mockRepository.listVibes(
        from: anyNamed('from'),
        to: anyNamed('to'),
      )).called(1);
      
      final state = container.read(vibeControllerProvider);
      expect(state.isLoading, false);
      expect(state.vibes, testVibes);
      expect(state.error, isNull);
    });

    test('updateVibe updates local state when vibe exists', () async {
      // Arrange
      final updatedVibe = testVibe.copyWith(value: 5, note: 'Updated note');
      
      // Set initial state with test vibes
      when(mockRepository.listVibes(
        from: anyNamed('from'),
        to: anyNamed('to'),
      )).thenAnswer((_) async => testVibes);
      await container.read(vibeControllerProvider.notifier).loadVibes();
      
      when(mockRepository.updateVibe(updatedVibe)).thenAnswer((_) async {});

      // Act
      await container.read(vibeControllerProvider.notifier).updateVibe(updatedVibe);
      
      // Assert
      verify(mockRepository.updateVibe(updatedVibe)).called(1);
      
      final state = container.read(vibeControllerProvider);
      expect(state.isLoading, false);
      expect(state.vibes.length, 2);
      
      final updatedVibeInState = state.vibes.firstWhere((v) => v.id == updatedVibe.id);
      expect(updatedVibeInState.value, 5);
      expect(updatedVibeInState.note, 'Updated note');
    });

    test('getVibesByType returns correct vibes', () async {
      // Arrange
      when(mockRepository.listVibes(
        from: anyNamed('from'),
        to: anyNamed('to'),
      )).thenAnswer((_) async => testVibes);
      await container.read(vibeControllerProvider.notifier).loadVibes();
      
      // Act
      final moodVibes = container.read(vibeControllerProvider.notifier)
          .getVibesByType(VibeType.mood);
      final sleepVibes = container.read(vibeControllerProvider.notifier)
          .getVibesByType(VibeType.sleep);
      
      // Assert
      expect(moodVibes.length, 1);
      expect(moodVibes.first.id, 'test-id-1');
      expect(sleepVibes.length, 1);
      expect(sleepVibes.first.id, 'test-id-2');
    });

    test('getMostRecentVibe returns correct vibe', () async {
      // Arrange
      when(mockRepository.listVibes(
        from: anyNamed('from'),
        to: anyNamed('to'),
      )).thenAnswer((_) async => testVibes);
      await container.read(vibeControllerProvider.notifier).loadVibes();
      
      // Act
      final recentMood = container.read(vibeControllerProvider.notifier)
          .getMostRecentVibe(VibeType.mood);
      
      // Assert
      expect(recentMood, isNotNull);
      expect(recentMood!.id, 'test-id-1');
    });

    test('clearError removes error from state', () async {
      // Arrange
      final testError = VibeServerError(message: 'Server error');
      when(mockRepository.listVibes(
        from: anyNamed('from'),
        to: anyNamed('to'),
      )).thenThrow(testError);
      await container.read(vibeControllerProvider.notifier).loadVibes();
      
      // Verify error is set
      expect(container.read(vibeControllerProvider).error, testError);
      
      // Act
      container.read(vibeControllerProvider.notifier).clearError();
      
      // Assert
      expect(container.read(vibeControllerProvider).error, isNull);
    });
  });
}
