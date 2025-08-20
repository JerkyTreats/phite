/// Represents the type of vibe check
enum VibeType {
  /// Sleep check, typically prompted in the morning
  sleep,
  
  /// Mood check, typically prompted in the afternoon
  mood;
  
  /// Returns a user-friendly display name for the vibe type
  String get displayName {
    switch (this) {
      case VibeType.sleep:
        return 'Sleep';
      case VibeType.mood:
        return 'Mood';
    }
  }
  
  /// Returns a prompt question based on the vibe type
  String get promptQuestion {
    switch (this) {
      case VibeType.sleep:
        return 'How did you sleep?';
      case VibeType.mood:
        return 'How\'s the day going?';
    }
  }
}
