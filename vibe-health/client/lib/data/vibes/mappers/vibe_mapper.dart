import '../dto/vibe_dto.dart';
import '../../../../features/vibes/domain/vibe.dart';
import '../../../../features/vibes/domain/vibe_type.dart';
import '../../../../features/vibes/domain/vibe_error.dart';

/// Mapper class for converting between [VibeDto] and [Vibe] domain entity
class VibeMapper {
  /// Converts a [VibeDto] to a domain [Vibe] entity
  /// 
  /// Throws [VibeParsingError] if required fields are missing or invalid
  static Vibe fromDto(VibeDto dto) {
    try {
      // Validate required fields
      final id = dto.id;
      final userId = dto.userId;
      
      if (id == null || userId == null) {
        throw const VibeParsingError('Missing required fields (id or userId)');
      }
      
      // Map the type string to VibeType enum
      final VibeType type;
      switch (dto.type.toLowerCase()) {
        case 'sleep':
          type = VibeType.sleep;
          break;
        case 'mood':
          type = VibeType.mood;
          break;
        default:
          throw VibeParsingError('Invalid vibe type: ${dto.type}');
      }
      
      return Vibe(
        id: id,
        userId: userId,
        type: type,
        value: dto.value,
        note: dto.note,
        ts: dto.ts,
      );
    } catch (e) {
      if (e is VibeError) {
        rethrow;
      }
      throw VibeParsingError('Failed to parse VibeDto: ${e.toString()}');
    }
  }

  /// Converts a domain [Vibe] entity to a [VibeDto]
  static VibeDto toDto(Vibe vibe) {
    return VibeDto(
      id: vibe.id,
      userId: vibe.userId,
      type: vibe.type.name,
      value: vibe.value,
      note: vibe.note,
      ts: vibe.ts,
    );
  }
  
  /// Converts a list of [VibeDto] objects to a list of domain [Vibe] entities
  /// 
  /// If any DTO fails to convert, it will be skipped and an error will be logged
  static List<Vibe> fromDtoList(List<VibeDto> dtos) {
    final List<Vibe> vibes = [];
    
    for (final dto in dtos) {
      try {
        vibes.add(fromDto(dto));
      } catch (e) {
        // In a real app, log this error
        print('Error converting DTO to Vibe: ${e.toString()}');
      }
    }
    
    return vibes;
  }
  
  /// Converts a list of domain [Vibe] entities to a list of [VibeDto] objects
  static List<VibeDto> toDtoList(List<Vibe> vibes) {
    return vibes.map((vibe) => toDto(vibe)).toList();
  }
}
