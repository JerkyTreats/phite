import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../features/vibes/domain/vibe.dart';
import '../../../features/vibes/domain/vibe_repository.dart';
import '../api/vibe_api.dart';
import '../api/mock_vibe_api.dart';
import '../mappers/vibe_mapper.dart';

/// Provider for the HTTP implementation of the VibeRepository
final vibeRepositoryProvider = Provider<VibeRepository>((ref) {
  // Use mock API for development
  final vibeApi = ref.watch(mockVibeApiProvider);
  return HttpVibeRepository(vibeApi);
});

/// HTTP implementation of the VibeRepository interface
class HttpVibeRepository implements VibeRepository {
  final VibeApi _vibeApi;

  HttpVibeRepository(this._vibeApi);

  @override
  Future<void> createVibe(Vibe vibe) async {
    final dto = VibeMapper.toDto(vibe);
    await _vibeApi.createVibe(dto);
  }

  @override
  Future<List<Vibe>> listVibes({
    required DateTime from,
    required DateTime to,
  }) async {
    final dtos = await _vibeApi.listVibes(from: from, to: to);
    return VibeMapper.fromDtoList(dtos);
  }

  @override
  Future<Vibe> getVibeById(String id) async {
    final dto = await _vibeApi.getVibeById(id);
    return VibeMapper.fromDto(dto);
  }

  @override
  Future<void> updateVibe(Vibe vibe) async {
    final dto = VibeMapper.toDto(vibe);
    await _vibeApi.updateVibe(vibe.id, dto);
  }

  @override
  Future<void> deleteVibe(String id) async {
    await _vibeApi.deleteVibe(id);
  }
}
