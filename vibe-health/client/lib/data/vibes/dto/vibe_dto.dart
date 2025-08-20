import 'package:json_annotation/json_annotation.dart';

part 'vibe_dto.g.dart';

/// Data Transfer Object for Vibe entity
/// Used for serialization/deserialization with the API
@JsonSerializable()
class VibeDto {
  /// Unique identifier for the vibe
  final String? id;
  
  /// User identifier associated with this vibe
  final String? userId;
  
  /// Type of vibe: "sleep" or "mood"
  final String type;
  
  /// Numeric value representing the emoji selection (1-5 scale)
  final int value;
  
  /// Optional note provided by the user
  final String? note;
  
  /// Timestamp in UTC ISO8601 format
  final String ts;

  /// Creates a new VibeDto instance
  VibeDto({
    this.id,
    this.userId,
    required this.type,
    required this.value,
    this.note,
    required this.ts,
  });

  /// Creates a VibeDto from JSON
  factory VibeDto.fromJson(Map<String, dynamic> json) => _$VibeDtoFromJson(json);

  /// Converts this VibeDto to JSON
  Map<String, dynamic> toJson() => _$VibeDtoToJson(this);
}
