import 'package:flutter_test/flutter_test.dart';
import 'package:mockito/annotations.dart';
import 'package:mockito/mockito.dart';

import 'package:vibe_health/data/vibes/api/vibe_api.dart';
import 'package:vibe_health/data/vibes/dto/vibe_dto.dart';
import 'package:vibe_health/data/vibes/repositories/http_vibe_repository.dart';
import 'package:vibe_health/features/vibes/domain/vibe.dart';
import 'package:vibe_health/features/vibes/domain/vibe_type.dart';

import 'http_vibe_repository_test.mocks.dart';

@GenerateMocks([VibeApi])
void main() {
  late MockVibeApi mockVibeApi;
  late HttpVibeRepository repository;

  // Sample test data
  final testVibe = Vibe(
    id: 'test-id-1',
    userId: 'test-user',
    type: VibeType.mood,
    value: 4,
    note: 'Test note',
    ts: '2025-08-20T12:00:00Z',
  );

  final testVibeDto = VibeDto(
    id: 'test-id-1',
    userId: 'test-user',
    type: 'mood',
    value: 4,
    note: 'Test note',
    ts: '2025-08-20T12:00:00Z',
  );

  final testVibeDtos = [
    testVibeDto,
    VibeDto(
      id: 'test-id-2',
      userId: 'test-user',
      type: 'sleep',
      value: 3,
      note: null,
      ts: '2025-08-20T08:00:00Z',
    ),
  ];

  setUp(() {
    mockVibeApi = MockVibeApi();
    repository = HttpVibeRepository(mockVibeApi);
  });

  group('HttpVibeRepository', () {
    test('createVibe calls API with correct DTO', () async {
      // Arrange
      when(mockVibeApi.createVibe(any)).thenAnswer((_) async => testVibeDtos[0]);

      // Act
      await repository.createVibe(testVibe);

      // Assert
      final captured = verify(mockVibeApi.createVibe(captureAny)).captured;
      expect(captured.length, 1);
      
      final capturedDto = captured.first as VibeDto;
      expect(capturedDto.id, testVibe.id);
      expect(capturedDto.userId, testVibe.userId);
      expect(capturedDto.type, testVibe.type.name);
      expect(capturedDto.value, testVibe.value);
      expect(capturedDto.note, testVibe.note);
      expect(capturedDto.ts, testVibe.ts);
    });

    test('listVibes returns correctly mapped domain objects', () async {
      // Arrange
      final testFrom = DateTime(2025, 8, 20);
      final testTo = DateTime(2025, 8, 21);
      
      when(mockVibeApi.listVibes(
        from: anyNamed('from'),
        to: anyNamed('to'),
      )).thenAnswer((_) async => testVibeDtos);

      // Act
      final result = await repository.listVibes(from: testFrom, to: testTo);

      // Assert
      verify(mockVibeApi.listVibes(from: testFrom, to: testTo)).called(1);
      expect(result.length, 2);
      
      // Verify first vibe
      expect(result[0].id, 'test-id-1');
      expect(result[0].type, VibeType.mood);
      expect(result[0].value, 4);
      expect(result[0].note, 'Test note');
      
      // Verify second vibe
      expect(result[1].id, 'test-id-2');
      expect(result[1].type, VibeType.sleep);
      expect(result[1].value, 3);
      expect(result[1].note, null);
    });

    test('getVibeById returns correctly mapped domain object', () async {
      // Arrange
      when(mockVibeApi.getVibeById('test-id-1')).thenAnswer((_) async => testVibeDto);

      // Act
      final result = await repository.getVibeById('test-id-1');

      // Assert
      verify(mockVibeApi.getVibeById('test-id-1')).called(1);
      expect(result.id, 'test-id-1');
      expect(result.type, VibeType.mood);
      expect(result.value, 4);
      expect(result.note, 'Test note');
    });

    test('updateVibe calls API with correct DTO', () async {
      // Arrange
      when(mockVibeApi.updateVibe(any, any)).thenAnswer((_) async => testVibeDtos[0]);

      // Act
      await repository.updateVibe(testVibe);

      // Assert
      final captured = verify(mockVibeApi.updateVibe('test-id-1', captureAny)).captured;
      expect(captured.length, 1);
      
      final capturedDto = captured.first as VibeDto;
      expect(capturedDto.id, testVibe.id);
      expect(capturedDto.userId, testVibe.userId);
      expect(capturedDto.type, testVibe.type.name);
      expect(capturedDto.value, testVibe.value);
      expect(capturedDto.note, testVibe.note);
      expect(capturedDto.ts, testVibe.ts);
    });

    test('deleteVibe calls API with correct ID', () async {
      // Arrange
      when(mockVibeApi.deleteVibe(any)).thenAnswer((_) async {});

      // Act
      await repository.deleteVibe('test-id-1');

      // Assert
      verify(mockVibeApi.deleteVibe('test-id-1')).called(1);
    });
  });
}
