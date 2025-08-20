import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../domain/vibe_type.dart';
import '../../../core/di/providers.dart';
import '../../../core/notifications/notification_service.dart';

part 'notification_controller.g.dart';

/// Keys for storing notification preferences
class NotificationPreferenceKeys {
  static const String sleepEnabled = 'sleep_notification_enabled';
  static const String sleepHour = 'sleep_notification_hour';
  static const String sleepMinute = 'sleep_notification_minute';
  static const String moodEnabled = 'mood_notification_enabled';
  static const String moodHour = 'mood_notification_hour';
  static const String moodMinute = 'mood_notification_minute';
}

/// State for the notification controller
class NotificationState {
  final bool sleepNotificationsEnabled;
  final TimeOfDay sleepNotificationTime;
  final bool moodNotificationsEnabled;
  final TimeOfDay moodNotificationTime;
  final bool isLoading;
  final String? error;

  const NotificationState({
    this.sleepNotificationsEnabled = false,
    this.sleepNotificationTime = const TimeOfDay(hour: 8, minute: 0),
    this.moodNotificationsEnabled = false,
    this.moodNotificationTime = const TimeOfDay(hour: 16, minute: 0),
    this.isLoading = false,
    this.error,
  });

  /// Creates a copy of this state with the given fields replaced
  NotificationState copyWith({
    bool? sleepNotificationsEnabled,
    TimeOfDay? sleepNotificationTime,
    bool? moodNotificationsEnabled,
    TimeOfDay? moodNotificationTime,
    bool? isLoading,
    String? error,
  }) {
    return NotificationState(
      sleepNotificationsEnabled: sleepNotificationsEnabled ?? this.sleepNotificationsEnabled,
      sleepNotificationTime: sleepNotificationTime ?? this.sleepNotificationTime,
      moodNotificationsEnabled: moodNotificationsEnabled ?? this.moodNotificationsEnabled,
      moodNotificationTime: moodNotificationTime ?? this.moodNotificationTime,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }
}

/// Controller for managing notifications
@riverpod
class NotificationController extends _$NotificationController {
  late SharedPreferences _prefs;
  late NotificationService _notificationService;

  @override
  NotificationState build() {
    _notificationService = ref.read(notificationServiceProvider);
    _prefs = ref.read(sharedPreferencesProvider);
    
    // Load saved preferences
    _loadPreferences();
    
    return const NotificationState(isLoading: true);
  }

  /// Loads saved notification preferences
  Future<void> _loadPreferences() async {
    try {
      final sleepEnabled = _prefs.getBool(NotificationPreferenceKeys.sleepEnabled) ?? false;
      final sleepHour = _prefs.getInt(NotificationPreferenceKeys.sleepHour) ?? 8;
      final sleepMinute = _prefs.getInt(NotificationPreferenceKeys.sleepMinute) ?? 0;
      
      final moodEnabled = _prefs.getBool(NotificationPreferenceKeys.moodEnabled) ?? false;
      final moodHour = _prefs.getInt(NotificationPreferenceKeys.moodHour) ?? 16;
      final moodMinute = _prefs.getInt(NotificationPreferenceKeys.moodMinute) ?? 0;
      
      state = state.copyWith(
        sleepNotificationsEnabled: sleepEnabled,
        sleepNotificationTime: TimeOfDay(hour: sleepHour, minute: sleepMinute),
        moodNotificationsEnabled: moodEnabled,
        moodNotificationTime: TimeOfDay(hour: moodHour, minute: moodMinute),
        isLoading: false,
      );
      
      // Schedule notifications based on saved preferences
      if (sleepEnabled) {
        _scheduleNotification(
          VibeType.sleep, 
          TimeOfDay(hour: sleepHour, minute: sleepMinute),
        );
      }
      
      if (moodEnabled) {
        _scheduleNotification(
          VibeType.mood, 
          TimeOfDay(hour: moodHour, minute: moodMinute),
        );
      }
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: 'Failed to load notification preferences: ${e.toString()}',
      );
    }
  }

  /// Saves notification preferences
  Future<void> _savePreferences() async {
    try {
      await _prefs.setBool(
        NotificationPreferenceKeys.sleepEnabled, 
        state.sleepNotificationsEnabled,
      );
      await _prefs.setInt(
        NotificationPreferenceKeys.sleepHour, 
        state.sleepNotificationTime.hour,
      );
      await _prefs.setInt(
        NotificationPreferenceKeys.sleepMinute, 
        state.sleepNotificationTime.minute,
      );
      
      await _prefs.setBool(
        NotificationPreferenceKeys.moodEnabled, 
        state.moodNotificationsEnabled,
      );
      await _prefs.setInt(
        NotificationPreferenceKeys.moodHour, 
        state.moodNotificationTime.hour,
      );
      await _prefs.setInt(
        NotificationPreferenceKeys.moodMinute, 
        state.moodNotificationTime.minute,
      );
    } catch (e) {
      state = state.copyWith(
        error: 'Failed to save notification preferences: ${e.toString()}',
      );
    }
  }

  /// Toggles notifications for a specific vibe type
  Future<void> toggleNotifications(VibeType type, bool enabled) async {
    state = state.copyWith(isLoading: true, error: null);
    
    try {
      // Request permissions if enabling notifications
      if (enabled) {
        final hasPermission = await _notificationService.requestPermissions();
        if (!hasPermission) {
          state = state.copyWith(
            isLoading: false,
            error: 'Notification permissions denied',
          );
          return;
        }
      }
      
      if (type == VibeType.sleep) {
        state = state.copyWith(sleepNotificationsEnabled: enabled);
        
        if (enabled) {
          await _scheduleNotification(type, state.sleepNotificationTime);
        } else {
          await _notificationService.cancelVibeCheck(type);
        }
      } else {
        state = state.copyWith(moodNotificationsEnabled: enabled);
        
        if (enabled) {
          await _scheduleNotification(type, state.moodNotificationTime);
        } else {
          await _notificationService.cancelVibeCheck(type);
        }
      }
      
      await _savePreferences();
      state = state.copyWith(isLoading: false);
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: 'Failed to toggle notifications: ${e.toString()}',
      );
    }
  }

  /// Updates the notification time for a specific vibe type
  Future<void> updateNotificationTime(VibeType type, TimeOfDay time) async {
    state = state.copyWith(isLoading: true, error: null);
    
    try {
      if (type == VibeType.sleep) {
        state = state.copyWith(sleepNotificationTime: time);
        
        if (state.sleepNotificationsEnabled) {
          await _scheduleNotification(type, time);
        }
      } else {
        state = state.copyWith(moodNotificationTime: time);
        
        if (state.moodNotificationsEnabled) {
          await _scheduleNotification(type, time);
        }
      }
      
      await _savePreferences();
      state = state.copyWith(isLoading: false);
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: 'Failed to update notification time: ${e.toString()}',
      );
    }
  }

  /// Schedules a notification for a specific vibe type
  Future<void> _scheduleNotification(VibeType type, TimeOfDay time) async {
    await _notificationService.scheduleDailyVibeCheck(
      vibeType: type,
      time: time,
    );
  }

  /// Clears any error in the state
  void clearError() {
    if (state.error != null) {
      state = state.copyWith(error: null);
    }
  }
}
