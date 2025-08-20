import 'package:equatable/equatable.dart';
import 'vibe_type.dart';

/// Represents a vibe check entry
class Vibe extends Equatable {
  /// Unique identifier for the vibe
  final String id;
  
  /// User identifier associated with this vibe
  final String userId;
  
  /// Type of vibe (sleep or mood)
  final VibeType type;
  
  /// Numeric value representing the emoji selection (1-5 scale)
  final int value;
  
  /// Optional note provided by the user
  final String? note;
  
  /// Timestamp in UTC ISO8601 format
  final String ts;

  /// Creates a new Vibe instance
  const Vibe({
    required this.id,
    required this.userId,
    required this.type,
    required this.value,
    this.note,
    required this.ts,
  });
  
  /// Creates a copy of this Vibe with the given fields replaced with new values
  Vibe copyWith({
    String? id,
    String? userId,
    VibeType? type,
    int? value,
    String? note,
    String? ts,
  }) {
    return Vibe(
      id: id ?? this.id,
      userId: userId ?? this.userId,
      type: type ?? this.type,
      value: value ?? this.value,
      note: note ?? this.note,
      ts: ts ?? this.ts,
    );
  }
  
  /// Returns emoji representation based on the value
  String get emoji {
    switch (value) {
      case 5: return '😁'; // Very happy
      case 4: return '🙂'; // Happy
      case 3: return '😐'; // Neutral
      case 2: return '🙁'; // Sad
      case 1: return '😡'; // Angry
      default: return '❓'; // Unknown
    }
  }
  
  @override
  List<Object?> get props => [id, userId, type, value, note, ts];
}
