import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:dio/dio.dart';

import '../di/providers.dart';

/// Provider for the auth service
final authServiceProvider = Provider<AuthService>((ref) {
  final secureStorage = ref.watch(secureStorageProvider);
  final apiBaseUrl = ref.watch(apiBaseUrlProvider);
  return AuthService(secureStorage, apiBaseUrl);
});

/// Provider for the current auth token
/// This overrides the placeholder in providers.dart
final authTokenProvider = Provider<String?>((ref) {
  final authService = ref.watch(authServiceProvider);
  return authService.currentToken;
});

/// Authentication service for managing user authentication
class AuthService {
  final FlutterSecureStorage _secureStorage;
  final String _apiBaseUrl;
  String? _token;
  
  static const _tokenKey = 'auth_token';
  static const _userIdKey = 'user_id';
  
  AuthService(this._secureStorage, this._apiBaseUrl);
  
  /// The current authentication token
  String? get currentToken => _token;
  
  /// Initialize the auth service
  Future<void> initialize() async {
    // Load token from secure storage if available
    _token = await _secureStorage.read(key: _tokenKey);
  }
  
  /// Login with email and password
  Future<bool> login(String email, String password) async {
    try {
      final dio = Dio();
      final response = await dio.post<Map<String, dynamic>>(
        '$_apiBaseUrl/auth/login',
        data: {
          'email': email,
          'password': password,
        },
      );
      
      if (response.statusCode == 200 && response.data != null) {
        final token = response.data!['token'] as String?;
        final userId = response.data!['userId'] as String?;
        
        if (token != null && userId != null) {
          await _secureStorage.write(key: _tokenKey, value: token);
          await _secureStorage.write(key: _userIdKey, value: userId);
          _token = token;
          return true;
        }
      }
      
      return false;
    } catch (e) {
      print('Login error: $e');
      return false;
    }
  }
  
  /// Register a new user
  Future<bool> register(String email, String password, String name) async {
    try {
      final dio = Dio();
      final response = await dio.post<Map<String, dynamic>>(
        '$_apiBaseUrl/auth/register',
        data: {
          'email': email,
          'password': password,
          'name': name,
        },
      );
      
      if (response.statusCode == 201 && response.data != null) {
        return true;
      }
      
      return false;
    } catch (e) {
      print('Registration error: $e');
      return false;
    }
  }
  
  /// Logout the current user
  Future<void> logout() async {
    await _secureStorage.delete(key: _tokenKey);
    await _secureStorage.delete(key: _userIdKey);
    _token = null;
  }
  
  /// Get the current user ID
  Future<String?> getCurrentUserId() async {
    return await _secureStorage.read(key: _userIdKey);
  }
  
  /// Check if the user is authenticated
  bool get isAuthenticated => _token != null;
  
  /// For development purposes - set a mock token
  Future<void> setMockToken(String token) async {
    await _secureStorage.write(key: _tokenKey, value: token);
    _token = token;
  }
}
