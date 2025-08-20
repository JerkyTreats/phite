import 'package:equatable/equatable.dart';

/// Base class for all vibe-related errors
abstract class VibeError extends Equatable implements Exception {
  /// Error message
  final String message;
  
  /// Creates a new VibeError instance
  const VibeError(this.message);
  
  @override
  List<Object> get props => [message];
  
  @override
  String toString() => message;
}

/// Error that occurs when authentication fails
class VibeAuthError extends VibeError {
  /// Creates a new VibeAuthError instance
  const VibeAuthError([String message = 'Authentication failed']) : super(message);
}

/// Error that occurs when there's a network issue
class VibeNetworkError extends VibeError {
  /// Creates a new VibeNetworkError instance
  const VibeNetworkError([String message = 'Network connection failed']) : super(message);
}

/// Error that occurs when the server returns an error
class VibeServerError extends VibeError {
  /// HTTP status code
  final int? statusCode;
  
  /// Creates a new VibeServerError instance
  const VibeServerError({
    String message = 'Server error occurred',
    this.statusCode,
  }) : super(message);
  
  @override
  List<Object> get props => [message, if (statusCode != null) statusCode!];
}

/// Error that occurs when parsing data fails
class VibeParsingError extends VibeError {
  /// Creates a new VibeParsingError instance
  const VibeParsingError([String message = 'Failed to parse data']) : super(message);
}

/// Error that occurs when a vibe is not found
class VibeNotFoundError extends VibeError {
  /// Creates a new VibeNotFoundError instance
  const VibeNotFoundError([String message = 'Vibe not found']) : super(message);
}

/// Error that occurs when an operation is not permitted
class VibePermissionError extends VibeError {
  /// Creates a new VibePermissionError instance
  const VibePermissionError([String message = 'Operation not permitted']) : super(message);
}

/// Error that occurs for unexpected cases
class VibeUnexpectedError extends VibeError {
  /// Creates a new VibeUnexpectedError instance
  const VibeUnexpectedError([String message = 'An unexpected error occurred']) : super(message);
}
