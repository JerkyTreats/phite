import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../controllers/notification_controller.dart';
import '../../domain/vibe_type.dart';
import '../widgets/app_drawer.dart';

/// Screen for configuring notification settings
class NotificationSettingsScreen extends ConsumerWidget {
  const NotificationSettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(notificationControllerProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Notification Settings'),
      ),
      drawer: const AppDrawer(),
      body: state.isLoading
          ? const Center(child: CircularProgressIndicator())
          : state.error != null
              ? _buildErrorView(context, state.error!, ref)
              : _buildSettingsForm(context, state, ref),
    );
  }

  Widget _buildErrorView(BuildContext context, String error, WidgetRef ref) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(
            Icons.error_outline,
            color: Colors.red,
            size: 60,
          ),
          const SizedBox(height: 16),
          Text(
            'Error: $error',
            textAlign: TextAlign.center,
            style: const TextStyle(fontSize: 16),
          ),
          const SizedBox(height: 24),
          ElevatedButton(
            onPressed: () {
              ref.read(notificationControllerProvider.notifier).clearError();
            },
            child: const Text('Try Again'),
          ),
        ],
      ),
    );
  }

  Widget _buildSettingsForm(
    BuildContext context,
    NotificationState state,
    WidgetRef ref,
  ) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Configure when you want to receive vibe check reminders',
            style: Theme.of(context).textTheme.bodyLarge,
          ),
          const SizedBox(height: 24),
          
          // Sleep notifications
          _buildNotificationSection(
            context: context,
            ref: ref,
            title: 'Sleep Check Notifications',
            subtitle: 'Receive a daily reminder to record your sleep quality',
            type: VibeType.sleep,
            enabled: state.sleepNotificationsEnabled,
            time: state.sleepNotificationTime,
          ),
          
          const Divider(height: 32),
          
          // Mood notifications
          _buildNotificationSection(
            context: context,
            ref: ref,
            title: 'Mood Check Notifications',
            subtitle: 'Receive a daily reminder to record your mood',
            type: VibeType.mood,
            enabled: state.moodNotificationsEnabled,
            time: state.moodNotificationTime,
          ),
          
          const SizedBox(height: 24),
          
          // Information about permissions
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      const Icon(Icons.info_outline, color: Colors.blue),
                      const SizedBox(width: 8),
                      Text(
                        'About Notifications',
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  const Text(
                    'You can change notification permissions in your device settings if notifications are not working properly.',
                  ),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () {
                      // In a real app, this would open the device notification settings
                      ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(
                          content: Text('This would open device settings'),
                        ),
                      );
                    },
                    child: const Text('Open Device Settings'),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildNotificationSection({
    required BuildContext context,
    required WidgetRef ref,
    required String title,
    required String subtitle,
    required VibeType type,
    required bool enabled,
    required TimeOfDay time,
  }) {
    final controller = ref.read(notificationControllerProvider.notifier);
    
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          title,
          style: Theme.of(context).textTheme.titleLarge,
        ),
        const SizedBox(height: 4),
        Text(
          subtitle,
          style: Theme.of(context).textTheme.bodyMedium,
        ),
        const SizedBox(height: 16),
        SwitchListTile(
          title: const Text('Enable Notifications'),
          value: enabled,
          onChanged: (value) {
            controller.toggleNotifications(type, value);
          },
        ),
        ListTile(
          enabled: enabled,
          title: const Text('Notification Time'),
          subtitle: Text(_formatTimeOfDay(time)),
          trailing: const Icon(Icons.access_time),
          onTap: enabled
              ? () async {
                  final newTime = await showTimePicker(
                    context: context,
                    initialTime: time,
                  );
                  if (newTime != null) {
                    controller.updateNotificationTime(type, newTime);
                  }
                }
              : null,
        ),
      ],
    );
  }

  String _formatTimeOfDay(TimeOfDay time) {
    final hour = time.hour.toString().padLeft(2, '0');
    final minute = time.minute.toString().padLeft(2, '0');
    return '$hour:$minute ${time.period.name.toUpperCase()}';
  }
}
