import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../core/network/http_client.dart';
import '../dto/vibe_dto.dart';
import '../../../features/vibes/domain/vibe_error.dart';

/// Provider for the Vibe API
final vibeApiProvider = Provider<VibeApi>((ref) {
  final dio = ref.watch(httpClientProvider);
  return VibeApi(dio);
});

/// API client for vibe-related endpoints
class VibeApi {
  final Dio _dio;
  static const _basePath = '/vibes';
  
  /// ISO8601 date format for API requests
  static final DateFormat _dateFormat = DateFormat('yyyy-MM-ddTHH:mm:ss.SSSZ');

  VibeApi(this._dio);

  /// Creates a new vibe entry
  /// 
  /// Returns the created vibe DTO
  /// Throws a [VibeError] on failure
  Future<VibeDto> createVibe(VibeDto vibe) async {
    try {
      final response = await _dio.post<Map<String, dynamic>>(
        _basePath,
        data: vibe.toJson(),
      );
      
      return VibeDto.fromJson(response.data!);
    } on DioException catch (e) {
      throw _handleDioError(e);
    } catch (e) {
      throw VibeUnexpectedError('Failed to create vibe: ${e.toString()}');
    }
  }

  /// Lists vibes within the specified date range
  /// 
  /// [from] - Start date (inclusive)
  /// [to] - End date (inclusive)
  /// 
  /// Returns a list of vibe DTOs
  /// Throws a [VibeError] on failure
  Future<List<VibeDto>> listVibes({
    required DateTime from,
    required DateTime to,
  }) async {
    try {
      final response = await _dio.get<List<dynamic>>(
        _basePath,
        queryParameters: {
          'from': _dateFormat.format(from),
          'to': _dateFormat.format(to),
        },
      );
      
      return response.data!
          .map((item) => VibeDto.fromJson(item as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      throw _handleDioError(e);
    } catch (e) {
      throw VibeUnexpectedError('Failed to list vibes: ${e.toString()}');
    }
  }

  /// Gets a vibe by its ID
  /// 
  /// [id] - The unique identifier of the vibe
  /// 
  /// Returns the vibe DTO
  /// Throws a [VibeError] on failure
  Future<VibeDto> getVibeById(String id) async {
    try {
      final response = await _dio.get<Map<String, dynamic>>(
        '$_basePath/$id',
      );
      
      return VibeDto.fromJson(response.data!);
    } on DioException catch (e) {
      throw _handleDioError(e);
    } catch (e) {
      throw VibeUnexpectedError('Failed to get vibe: ${e.toString()}');
    }
  }

  /// Updates an existing vibe
  /// 
  /// [id] - The unique identifier of the vibe
  /// [vibe] - The vibe DTO with updated values
  /// 
  /// Returns the updated vibe DTO
  /// Throws a [VibeError] on failure
  Future<VibeDto> updateVibe(String id, VibeDto vibe) async {
    try {
      final response = await _dio.put<Map<String, dynamic>>(
        '$_basePath/$id',
        data: vibe.toJson(),
      );
      
      return VibeDto.fromJson(response.data!);
    } on DioException catch (e) {
      throw _handleDioError(e);
    } catch (e) {
      throw VibeUnexpectedError('Failed to update vibe: ${e.toString()}');
    }
  }

  /// Deletes a vibe by its ID
  /// 
  /// [id] - The unique identifier of the vibe to delete
  /// 
  /// Throws a [VibeError] on failure
  Future<void> deleteVibe(String id) async {
    try {
      await _dio.delete('$_basePath/$id');
    } on DioException catch (e) {
      throw _handleDioError(e);
    } catch (e) {
      throw VibeUnexpectedError('Failed to delete vibe: ${e.toString()}');
    }
  }

  /// Handles Dio errors and converts them to domain-specific errors
  VibeError _handleDioError(DioException e) {
    switch (e.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
      case DioExceptionType.connectionError:
        return VibeNetworkError('Network connection error: ${e.message}');
      
      case DioExceptionType.badResponse:
        final statusCode = e.response?.statusCode;
        if (statusCode == 401 || statusCode == 403) {
          return VibeAuthError('Authentication error: ${e.message}');
        } else if (statusCode == 404) {
          return VibeNotFoundError('Resource not found: ${e.message}');
        } else if (statusCode == 400) {
          return VibeParsingError('Bad request: ${e.message}');
        } else {
          return VibeServerError(message: 'Server error: ${e.message}');
        }
      
      default:
        return VibeUnexpectedError('Unexpected error: ${e.message}');
    }
  }
}
