import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Core providers for dependency injection
/// These providers are used across the application

/// Provider for secure storage
final secureStorageProvider = Provider<FlutterSecureStorage>((ref) {
  return const FlutterSecureStorage();
});

/// Provider for shared preferences
final sharedPreferencesProvider = Provider<SharedPreferences>((ref) {
  throw UnimplementedError('Initialize this provider in main.dart');
});

/// Provider for the base API URL
final apiBaseUrlProvider = Provider<String>((ref) {
  // This would typically come from environment variables
  // For now, we'll use a default local development URL
  return const String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://localhost:8080',
  );
});

/// Provider for the auth token
final authTokenProvider = Provider<String?>((ref) {
  throw UnimplementedError('Initialize this provider in the auth service');
});
