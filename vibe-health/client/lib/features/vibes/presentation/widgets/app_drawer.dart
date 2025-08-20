import 'package:flutter/material.dart';

/// App drawer for navigation
class AppDrawer extends StatelessWidget {
  const AppDrawer({super.key});

  @override
  Widget build(BuildContext context) {
    return Drawer(
      child: ListView(
        padding: EdgeInsets.zero,
        children: [
          DrawerHeader(
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.primary,
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Vibe Health',
                  style: TextStyle(
                    color: Colors.white,
                    fontSize: 24,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  'Track your daily vibes',
                  style: TextStyle(
                    color: Colors.white.withOpacity(0.8),
                    fontSize: 16,
                  ),
                ),
              ],
            ),
          ),
          ListTile(
            leading: const Icon(Icons.dashboard),
            title: const Text('Dashboard'),
            onTap: () {
              Navigator.pushReplacementNamed(context, '/');
            },
          ),
          ListTile(
            leading: const Icon(Icons.bedtime),
            title: const Text('Sleep Check'),
            onTap: () {
              Navigator.pop(context);
              Navigator.pushNamed(
                context, 
                '/vibe/entry',
                arguments: {'type': 'sleep'},
              );
            },
          ),
          ListTile(
            leading: const Icon(Icons.mood),
            title: const Text('Mood Check'),
            onTap: () {
              Navigator.pop(context);
              Navigator.pushNamed(
                context, 
                '/vibe/entry',
                arguments: {'type': 'mood'},
              );
            },
          ),
          const Divider(),
          ListTile(
            leading: const Icon(Icons.notifications),
            title: const Text('Notification Settings'),
            onTap: () {
              Navigator.pop(context);
              Navigator.pushNamed(context, '/settings/notifications');
            },
          ),
          ListTile(
            leading: const Icon(Icons.help_outline),
            title: const Text('About'),
            onTap: () {
              Navigator.pop(context);
              showAboutDialog(
                context: context,
                applicationName: 'Vibe Health',
                applicationVersion: '1.0.0',
                applicationIcon: const FlutterLogo(size: 48),
                children: [
                  const Text(
                    'Vibe Health helps you track your daily mood and sleep quality to improve your overall well-being.',
                  ),
                ],
              );
            },
          ),
        ],
      ),
    );
  }
}
