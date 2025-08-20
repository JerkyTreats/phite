import 'dart:io' show Platform;

import 'package:flutter/material.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:timezone/data/latest.dart' as tz;
import 'package:timezone/timezone.dart' as tz;

import '../../features/vibes/domain/vibe_type.dart';

/// Provider for the notification service
final notificationServiceProvider = Provider<NotificationService>((ref) {
  return NotificationService();
});

/// Service for handling local notifications
class NotificationService {
  final FlutterLocalNotificationsPlugin _notifications = FlutterLocalNotificationsPlugin();
  
  /// Channel ID for vibe check notifications
  static const String _vibeChannelId = 'vibe_checks';
  
  /// Channel name for vibe check notifications
  static const String _vibeChannelName = 'Vibe Checks';
  
  /// Channel description for vibe check notifications
  static const String _vibeChannelDescription = 'Daily vibe check notifications';
  
  /// Notification ID for sleep vibe checks
  static const int sleepVibeNotificationId = 1;
  
  /// Notification ID for mood vibe checks
  static const int moodVibeNotificationId = 2;

  /// Initialize the notification service
  Future<void> initialize() async {
    tz.initializeTimeZones();
    
    // Initialize the plugin
    const AndroidInitializationSettings androidSettings = 
        AndroidInitializationSettings('@mipmap/ic_launcher');
        
    const DarwinInitializationSettings iosSettings = DarwinInitializationSettings(
      requestAlertPermission: true,
      requestBadgePermission: true,
      requestSoundPermission: true,
    );
    
    const InitializationSettings initSettings = InitializationSettings(
      android: androidSettings,
      iOS: iosSettings,
    );
    
    await _notifications.initialize(
      initSettings,
      onDidReceiveNotificationResponse: _onNotificationResponse,
    );
    
    // Create the notification channel for Android
    await _createNotificationChannel();
  }
  
  /// Request notification permissions
  Future<bool> requestPermissions() async {
    bool permissionGranted = false;
    
    if (Platform.isAndroid) {
      final AndroidFlutterLocalNotificationsPlugin? androidPlugin = 
          _notifications.resolvePlatformSpecificImplementation<AndroidFlutterLocalNotificationsPlugin>();
          
      permissionGranted = await androidPlugin?.requestNotificationsPermission() ?? false;
    } else if (Platform.isIOS) {
      final IOSFlutterLocalNotificationsPlugin? iosPlugin = 
          _notifications.resolvePlatformSpecificImplementation<IOSFlutterLocalNotificationsPlugin>();
          
      permissionGranted = await iosPlugin?.requestPermissions(
        alert: true,
        badge: true,
        sound: true,
      ) ?? false;
    }
    
    return permissionGranted;
  }
  
  /// Create the notification channel for Android
  Future<void> _createNotificationChannel() async {
    const AndroidNotificationChannel channel = AndroidNotificationChannel(
      _vibeChannelId,
      _vibeChannelName,
      description: _vibeChannelDescription,
      importance: Importance.high,
    );
    
    final AndroidFlutterLocalNotificationsPlugin? androidPlugin = 
        _notifications.resolvePlatformSpecificImplementation<AndroidFlutterLocalNotificationsPlugin>();
        
    await androidPlugin?.createNotificationChannel(channel);
  }
  
  /// Handle notification responses
  void _onNotificationResponse(NotificationResponse response) {
    // Parse the payload to determine which vibe type was tapped
    final payload = response.payload;
    if (payload == null) return;
    
    // In a real app, you would navigate to the appropriate screen
    // based on the payload
    print('Notification tapped with payload: $payload');
    
    // Example of how to handle deep linking from notifications
    // final uri = Uri.parse(payload);
    // if (uri.scheme == 'vibehealth' && uri.host == 'vibe') {
    //   final type = uri.queryParameters['type'];
    //   final value = uri.queryParameters['value'];
    //   // Navigate to vibe entry screen with these parameters
    // }
  }
  
  /// Schedule a daily vibe check notification
  Future<void> scheduleDailyVibeCheck({
    required VibeType vibeType,
    required TimeOfDay time,
  }) async {
    final int notificationId = vibeType == VibeType.sleep 
        ? sleepVibeNotificationId 
        : moodVibeNotificationId;
    
    // Create the notification details
    final notificationDetails = NotificationDetails(
      android: AndroidNotificationDetails(
        _vibeChannelId,
        _vibeChannelName,
        channelDescription: _vibeChannelDescription,
        importance: Importance.high,
        priority: Priority.high,
        showWhen: true,
      ),
      iOS: const DarwinNotificationDetails(
        presentAlert: true,
        presentBadge: true,
        presentSound: true,
      ),
    );
    
    // Calculate the next occurrence of the specified time
    final now = DateTime.now();
    var scheduledDate = DateTime(
      now.year,
      now.month,
      now.day,
      time.hour,
      time.minute,
    );
    
    // If the time has already passed today, schedule for tomorrow
    if (scheduledDate.isBefore(now)) {
      scheduledDate = scheduledDate.add(const Duration(days: 1));
    }
    
    // Create the deep link payload
    final payload = 'vibehealth://vibe?type=${vibeType.name}';
    
    // Schedule the notification
    await _notifications.zonedSchedule(
      notificationId,
      'Time for your ${vibeType.displayName} check',
      vibeType.promptQuestion,
      tz.TZDateTime.from(scheduledDate, tz.local),
      notificationDetails,
      androidScheduleMode: AndroidScheduleMode.exactAllowWhileIdle,
      matchDateTimeComponents: DateTimeComponents.time,
      payload: payload,
    );
  }
  
  /// Cancel a scheduled vibe check notification
  Future<void> cancelVibeCheck(VibeType vibeType) async {
    final int notificationId = vibeType == VibeType.sleep 
        ? sleepVibeNotificationId 
        : moodVibeNotificationId;
        
    await _notifications.cancel(notificationId);
  }
  
  /// Cancel all scheduled notifications
  Future<void> cancelAllNotifications() async {
    await _notifications.cancelAll();
  }
}
